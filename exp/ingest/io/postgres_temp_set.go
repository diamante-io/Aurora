package io

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/errors"
)

const (
	// postgresTempSetCacheSize defines the maximum number of
	// entries in the cache. When the number of entries exceed part of
	// the cache is dumped to the DB.
	// Change the value to lower: smaller memory requirements but slower.
	// Change the value to higher: higher memory requirements but faster.
	postgresTempSetCacheSize = 1000000
	// postgresMaxQueryParams defines maximum number of params in a single
	// postgres query.
	postgresMaxQueryParams = 65535
)

// PostgresTempSet is a postgres implementation of TempSet interface.
// It's around 4x slower than MemoryStateReaderTempStore but has much
// lower memory requirements. `postgresTempSetCacheSize` can be changed
// to achieve better speed at the cost of higher memory usage.
// If `DSN` is passed, a new `db.Session` will be created.
// If `Session` is passed, it will be cloned.
type PostgresTempSet struct {
	DSN     string
	Session *db.Session

	afterFirstDump bool
	// cache can contain both true and false values. false values can be set
	// when key is not in cache: then the value will be checked in a database
	// and later cached.
	cache     map[string]bool
	tableName string
}

// Open connects to a DB and creates a temporary table and data structures.
func (s *PostgresTempSet) Open() error {
	var err error
	if s.Session == nil {
		s.Session, err = db.Open("postgres", s.DSN)
		if err != nil {
			return err
		}
	} else {
		// Clone existing session to prevent deadlock if the existing session
		// is in a transaction.
		s.Session = s.Session.Clone()
	}

	// Begin transaction - without it `ON COMMIT DROP` won't work.
	err = s.Session.Begin()
	if err != nil {
		return err
	}

	// Create a temporary table name
	r := make([]byte, 32)
	_, err = rand.Read(r)
	if err != nil {
		return err
	}

	s.tableName = fmt.Sprintf("exp_state_reader_%s", hex.EncodeToString(r))

	// Create table
	_, err = s.Session.ExecRaw(fmt.Sprintf(`
	CREATE TEMPORARY TABLE %s (
	    key character varying(255) NOT NULL,
	    PRIMARY KEY (key)
	) ON COMMIT DROP;`, s.tableName))
	if err != nil {
		return errors.Wrap(err, "Error creating table")
	}

	s.cache = make(map[string]bool)
	return nil
}

// Preload preloads `keys` from a database. It ignores the keys that
// are already in cache because these are the most recent values.
func (s *PostgresTempSet) Preload(keys []string) error {
	// Before first dump, there are no keys in a database.
	if !s.afterFirstDump {
		return nil
	}

	// The cache has always the latest data that (maybe) was not dumped to
	// a database yet so check it first. Then constuct a slice of keys
	// that actually need to be loaded from a DB.
	loadKeys := make([]string, 0, len(keys))
	keysMap := make(map[string]bool)
	for _, key := range keys {
		if _, exist := s.cache[key]; exist {
			continue
		}

		keysMap[key] = true
		loadKeys = append(loadKeys, key)
	}

	// At this point `keysMap` and `loadKeys` contains only the keys that
	// are not currently cached.

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	rows, err := psql.Select("key").
		From(s.tableName).
		Where(map[string]interface{}{"key": loadKeys}).
		RunWith(s.Session.GetTx().Tx).
		Query()

	if err != nil {
		return err
	}

	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			return err
		}

		// If value is found in a DB we cache with `true` value
		s.cache[key] = true
		delete(keysMap, key)
	}

	// For all the keys that are left (not found in a DB) we set `false` value.
	for key := range keysMap {
		s.cache[key] = false
	}

	return nil
}

// Add adds a key to TempSet.
func (s *PostgresTempSet) Add(key string) error {
	s.cache[key] = true
	return s.dumpCacheIfNeeded()
}

// dumpCacheIfNeeded dumps cache to a database if cache size exceeds
// postgresTempSetCacheSize.
func (s *PostgresTempSet) dumpCacheIfNeeded() error {
	cacheLen := len(s.cache)
	if cacheLen < postgresTempSetCacheSize {
		return nil
	}

	s.afterFirstDump = true

	query := s.newInsertBuilder()
	dumpedEntries := 0
	queryParams := 0

	for k, v := range s.cache {
		// We omit `false` values set in `Get`. We only store `true` values
		// in a DB.
		if !v {
			continue
		}

		query = query.Values(k)
		delete(s.cache, k)

		dumpedEntries++
		queryParams++

		// The number comes from the fact that the maximum number of params per
		// postgres query is 65535. When we are approaching the max params,
		// insert rows we have so far and create a new builder.
		if queryParams > postgresMaxQueryParams-500 {
			_, err := s.Session.Exec(query)
			if err != nil {
				return err
			}

			query = s.newInsertBuilder()
			queryParams = 0
		}

		// We dump only 1/2 (random) keys in cache. This is to hold _some_ keys
		// in memory and to not spend too much time dumping data. This 1:1 ratio
		// was confirmed to be the best by experimenting with different options.
		if dumpedEntries >= cacheLen/2 {
			break
		}
	}

	// Insert the last batch.
	if queryParams > 0 {
		_, err := s.Session.Exec(query)
		return err
	}

	return nil
}

// newInsertBuilder creates a new InsertBuilder.
func (s *PostgresTempSet) newInsertBuilder() sq.InsertBuilder {
	return sq.Insert(s.tableName).
		Columns("key").
		Suffix("ON CONFLICT (key) DO NOTHING")
}

// Exist check if the key exists in a TempSet.
func (s *PostgresTempSet) Exist(key string) (bool, error) {
	// Cache has the latest data: check it first.
	value, exist := s.cache[key]
	if exist {
		// This can be `true` or `false`. `false` values can be set below.
		return value, nil
	}

	// Before first dump all `true` entries should be in cache. If it's
	// not found then return false. It improves performance a lot before
	// the first data is dumped. Otherwise each `Exist` would send a
	// corresponding DB query.
	if !s.afterFirstDump {
		return false, nil
	}

	value, err := s.dbGet(key)
	if err != nil {
		return false, err
	}

	s.cache[key] = value

	err = s.dumpCacheIfNeeded()
	return value, err
}

// dbGet gets a single row from a database.
func (s *PostgresTempSet) dbGet(key string) (bool, error) {
	query := sq.Select("1").
		From(s.tableName).
		Where("key = ?", key)

	var value int
	if err := s.Session.Get(&value, query); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "could not get value")
	}

	// Value found in a store
	return true, nil
}

// Close closes a database connection what also removes a temporary table.
func (s *PostgresTempSet) Close() error {
	// Remove reference to a map
	s.cache = nil
	// This will drop temp table. Do not use `s.Session.Close` as
	// DB in `s.Session` can be shared by many sessions.
	return s.Session.Rollback()
}
