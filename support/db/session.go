package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/hcnet/go/support/db/sqlutils"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/support/log"
)

// Begin binds this session to a new transaction.
func (s *Session) Begin() error {
	if s.tx != nil {
		return errors.New("already in transaction")
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return errors.Wrap(err, "beginx failed")
	}
	s.logBegin()

	s.tx = tx
	return nil
}

// BeginTx binds this session to a new transaction which is configured with the
// given transaction options
func (s *Session) BeginTx(opts *sql.TxOptions) error {
	if s.tx != nil {
		return errors.New("already in transaction")
	}

	tx, err := s.DB.BeginTxx(s.Ctx, opts)
	if err != nil {
		return errors.Wrap(err, "beginTx failed")
	}
	s.logBegin()

	s.tx = tx
	return nil
}

func (s *Session) GetTx() *sqlx.Tx {
	return s.tx
}

// Clone clones the receiver, returning a new instance backed by the same
// context and db. The result will not be bound to any transaction that the
// source is currently within.
func (s *Session) Clone() *Session {
	return &Session{
		DB:  s.DB,
		Ctx: s.Ctx,
	}
}

// Close delegates to the underlying database Close method, closing the database
// and releasing any resources. It is rare to Close a DB, as the DB handle is meant
// to be long-lived and shared between many goroutines.
func (s *Session) Close() error {
	return s.DB.Close()
}

// Commit commits the current transaction
func (s *Session) Commit() error {
	if s.tx == nil {
		return errors.New("not in transaction")
	}

	err := s.tx.Commit()
	s.logCommit()
	s.tx = nil
	return err
}

// Dialect returns the SQL dialect that this session is configured to use
func (s *Session) Dialect() string {
	return s.DB.DriverName()
}

// DeleteRange deletes a range of rows from a sql table between `start` and
// `end` (exclusive).
func (s *Session) DeleteRange(
	start, end int64,
	table string,
	idCol string,
) error {
	del := sq.Delete(table).Where(
		fmt.Sprintf("%s >= ? AND %s < ?", idCol, idCol),
		start,
		end,
	)
	_, err := s.Exec(del)
	return err
}

// Get runs `query`, setting the first result found on `dest`, if
// any.
func (s *Session) Get(dest interface{}, query sq.Sqlizer) error {
	sql, args, err := s.build(query)
	if err != nil {
		return err
	}
	return s.GetRaw(dest, sql, args...)
}

// GetRaw runs `query` with `args`, setting the first result found on
// `dest`, if any.
func (s *Session) GetRaw(dest interface{}, query string, args ...interface{}) error {
	query, err := s.ReplacePlaceholders(query)
	if err != nil {
		return errors.Wrap(err, "replace placeholders failed")
	}

	start := time.Now()
	err = s.conn().Get(dest, query, args...)
	s.log("get", start, query, args)

	if err == nil {
		return nil
	}

	if s.NoRows(err) {
		return err
	}

	return errors.Wrap(err, "get failed")
}

// GetTable translates the provided struct into a Table,
func (s *Session) GetTable(name string) *Table {
	return &Table{
		Name:    name,
		Session: s,
	}
}

func (s *Session) TruncateTables(tables []string) error {
	truncateCmd := fmt.Sprintf("truncate %s cascade", strings.Join(tables[:], ","))
	_, err := s.ExecRaw(truncateCmd)
	return err
}

// Exec runs `query`
func (s *Session) Exec(query sq.Sqlizer) (sql.Result, error) {
	sql, args, err := s.build(query)
	if err != nil {
		return nil, err
	}
	return s.ExecRaw(sql, args...)
}

// ExecAll runs all sql commands in `script` against `r` within a single
// transaction.
func (s *Session) ExecAll(script string) error {
	err := s.Begin()
	if err != nil {
		return err
	}

	defer s.Rollback()

	for _, cmd := range sqlutils.AllStatements(script) {
		_, err = s.ExecRaw(cmd)
		if err != nil {
			return err
		}
	}

	return s.Commit()
}

// ExecRaw runs `query` with `args`
func (s *Session) ExecRaw(query string, args ...interface{}) (sql.Result, error) {
	query, err := s.ReplacePlaceholders(query)
	if err != nil {
		return nil, errors.Wrap(err, "replace placeholders failed")
	}

	start := time.Now()
	result, err := s.conn().Exec(query, args...)
	s.log("exec", start, query, args)

	if err == nil {
		return result, nil
	}

	if s.NoRows(err) {
		return nil, err
	}

	return nil, errors.Wrap(err, "exec failed")
}

// NoRows returns true if the provided error resulted from a query that found
// no results.
func (s *Session) NoRows(err error) bool {
	return err == sql.ErrNoRows
}

// Query runs `query`, returns a *sqlx.Rows instance
func (s *Session) Query(query sq.Sqlizer) (*sqlx.Rows, error) {
	sql, args, err := s.build(query)
	if err != nil {
		return nil, err
	}
	return s.QueryRaw(sql, args...)
}

// QueryRaw runs `query` with `args`
func (s *Session) QueryRaw(query string, args ...interface{}) (*sqlx.Rows, error) {
	query, err := s.ReplacePlaceholders(query)
	if err != nil {
		return nil, errors.Wrap(err, "replace placeholders failed")
	}

	start := time.Now()
	result, err := s.conn().Queryx(query, args...)
	s.log("query", start, query, args)

	if err == nil {
		return result, nil
	}

	if s.NoRows(err) {
		return nil, err
	}

	return nil, errors.Wrap(err, "query failed")
}

// ReplacePlaceholders replaces the '?' parameter placeholders in the provided
// sql query with a sql dialect appropriate version. Use '??' to escape a
// placeholder.
func (s *Session) ReplacePlaceholders(query string) (string, error) {
	var format sq.PlaceholderFormat = sq.Question

	if s.DB.DriverName() == "postgres" {
		format = sq.Dollar
	}
	return format.ReplacePlaceholders(query)
}

// Rollback rolls back the current transaction
func (s *Session) Rollback() error {
	if s.tx == nil {
		return errors.New("not in transaction")
	}

	err := s.tx.Rollback()
	s.logRollback()
	s.tx = nil
	return err
}

// Select runs `query`, setting the results found on `dest`.
func (s *Session) Select(dest interface{}, query sq.Sqlizer) error {
	sql, args, err := s.build(query)
	if err != nil {
		return err
	}
	return s.SelectRaw(dest, sql, args...)
}

// SelectRaw runs `query` with `args`, setting the results found on `dest`.
func (s *Session) SelectRaw(
	dest interface{},
	query string,
	args ...interface{},
) error {
	s.clearSliceIfPossible(dest)
	query, err := s.ReplacePlaceholders(query)
	if err != nil {
		return errors.Wrap(err, "replace placeholders failed")
	}

	start := time.Now()
	err = s.conn().Select(dest, query, args...)
	s.log("select", start, query, args)

	if err == nil {
		return nil
	}

	if s.NoRows(err) {
		return err
	}

	return errors.Wrap(err, "select failed")
}

// build converts the provided sql builder `b` into the sql and args to execute
// against the raw database connections.
func (s *Session) build(b sq.Sqlizer) (sql string, args []interface{}, err error) {
	sql, args, err = b.ToSql()

	if err != nil {
		err = errors.Wrap(err, "to-sql failed")
	}
	return
}

// clearSliceIfPossible is a utility function that clears a slice if the
// provided interface wraps one. In the event that `dest` is not a pointer to a
// slice this func will fail with a warning, this allowing the forthcoming db
// select fail more concretely due to an incompatible destination.
func (s *Session) clearSliceIfPossible(dest interface{}) {
	v := reflect.ValueOf(dest)
	vt := v.Type()

	if vt.Kind() != reflect.Ptr {
		log.Warn("cannot clear slice: dest is not pointer")
		return
	}

	if vt.Elem().Kind() != reflect.Slice {
		log.Warn("cannot clear slice: dest is a pointer, but not to a slice")
		return
	}

	reflect.Indirect(v).SetLen(0)
}

func (s *Session) conn() Conn {
	if s.tx != nil {
		return s.tx
	}

	return s.DB
}

func (s *Session) log(typ string, start time.Time, query string, args []interface{}) {
	log.
		Ctx(s.logCtx()).
		WithField("args", args).
		WithField("sql", query).
		WithField("dur", time.Since(start).String()).
		Debugf("sql: %s", typ)
}

func (s *Session) logBegin() {
	log.Ctx(s.logCtx()).Debug("sql: begin")
}

func (s *Session) logCommit() {
	log.Ctx(s.logCtx()).Debug("sql: commit")
}

func (s *Session) logRollback() {
	log.Ctx(s.logCtx()).Debug("sql: rollback")
}

func (s *Session) logCtx() context.Context {
	if s.Ctx != nil {
		return s.Ctx
	}

	return context.Background()
}
