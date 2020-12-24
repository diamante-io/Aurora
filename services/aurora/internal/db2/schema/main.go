package schema

import (
	"database/sql"
	"errors"
	stdLog "log"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/diamnet/go/support/db"
)

//go:generate go-bindata -ignore .+\.go$ -pkg schema -o bindata.go ./...

// MigrateDir represents a direction in which to perform schema migrations.
type MigrateDir string

const (
	// MigrateUp causes migrations to be run in the "up" direction.
	MigrateUp MigrateDir = "up"
	// MigrateDown causes migrations to be run in the "down" direction.
	MigrateDown MigrateDir = "down"
	// MigrateRedo causes migrations to be run down, then up
	MigrateRedo MigrateDir = "redo"
)

// Migrations represents all of the schema migration for aurora
var Migrations migrate.MigrationSource = &migrate.AssetMigrationSource{
	Asset:    Asset,
	AssetDir: AssetDir,
	Dir:      "migrations",
}

// Init installs the latest schema into db after clearing it first
func Init(db *db.Session) error {
	return db.ExecAll(string(MustAsset("latest.sql")))
}

// Migrate performs schema migration.  Migrations can occur in one of three
// ways:
//
// - up: migrations are performed from the currently installed version upwards.
// If count is 0, all unapplied migrations will be run.
//
// - down: migrations are performed from the current version downard. If count
// is 0, all applied migrations will be run in a downard direction.
//
// - redo: migrations are first ran downard `count` times, and then are rand
// upward back to the current version at the start of the process. If count is
// 0, a count of 1 will be assumed.
func Migrate(db *sql.DB, dir MigrateDir, count int) (int, error) {
	switch dir {
	case MigrateUp:
		return migrate.ExecMax(db, "postgres", Migrations, migrate.Up, count)
	case MigrateDown:
		return migrate.ExecMax(db, "postgres", Migrations, migrate.Down, count)
	case MigrateRedo:

		if count == 0 {
			count = 1
		}

		down, err := migrate.ExecMax(db, "postgres", Migrations, migrate.Down, count)
		if err != nil {
			return down, err
		}

		return migrate.ExecMax(db, "postgres", Migrations, migrate.Up, down)
	default:
		return 0, errors.New("Invalid migration direction")
	}
}

// GetMigrationsUp returns a list of names of any migrations needed in the
// "up" direction (more recent schema versions).
func GetMigrationsUp(dbUrl string) (migrationIds []string) {
	// Get a DB handle
	db, dbErr := sql.Open("postgres", dbUrl)
	if dbErr != nil {
		stdLog.Fatal(dbErr)
	}

	// Get the possible migrations
	possibleMigrations, _, migrateErr := migrate.PlanMigration(db, "postgres", Migrations, migrate.Up, 0)
	if migrateErr != nil {
		stdLog.Fatal(migrateErr)
	}

	// Extract a list of the possible migration names
	for _, m := range possibleMigrations {
		migrationIds = append(migrationIds, m.Id)
	}

	return migrationIds
}

// GetNumMigrationsDown returns the number of migrations to apply in the
// "down" direction to return to the older schema version expected by this
// version of Aurora. To keep the code simple, it does not provide a list of
// migration names.
func GetNumMigrationsDown(dbUrl string) (nMigrations int) {
	// Get a DB handle
	db, dbErr := sql.Open("postgres", dbUrl)
	if dbErr != nil {
		stdLog.Fatal(dbErr)
	}

	// Get the set of migrations recorded in the database
	migrationRecords, recordErr := migrate.GetMigrationRecords(db, "postgres")
	if recordErr != nil {
		stdLog.Fatal(recordErr)
	}

	// Get the list of migrations needed by this version of Aurora
	allNeededMigrations, _, migrateErr := migrate.PlanMigration(db, "postgres", Migrations, migrate.Down, 0)
	if migrateErr != nil {
		stdLog.Fatal(migrateErr)
	}

	// Return the size difference between the two sets of migrations
	return len(migrationRecords) - len(allNeededMigrations)
}
