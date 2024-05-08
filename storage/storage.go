package storage

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/tmwalaszek/hload/model"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed migrations/*.sql
	fs embed.FS
)

type Storage struct {
	db *sqlx.DB
}

type StorageError struct {
	Err           error
	RollbackError error
}

func (s StorageError) Error() string {
	if s.RollbackError != nil && s.Err != nil {
		return fmt.Sprintf("rollback error %v follwed the orignial storage error %v", s.RollbackError, s.Err)
	}

	if s.RollbackError != nil {
		return fmt.Sprintf("rollback error: %v", s.RollbackError)
	}

	if s.Err != nil {
		return s.Err.Error()
	}

	return ""
}

// loaderAggregated is a struct that will be used to store the result of the query in select_loader.tmpl
// from this struct we will build the model.Loader with the tags, headers and parameters in their respective models
type loaderAggregated struct {
	Header     string `db:"headers_agg"`
	Parameter  string `db:"parameters_agg"`
	TagsKeys   string `db:"tags_keys_agg"`
	TagsValues string `db:"tags_values_agg"`

	model.Loader
}

// summaryAggregated is a struct that will be used to store the result of the query in select_summaries.tmpl
// from this struct we will build the model.Summary with the errors and http codes in their respective models
type summaryAggregated struct {
	ErrorName  string `db:"errors_name_agg"`
	ErrorCount string `db:"errors_count_agg"`

	HTTPCode  string `db:"http_codes_code_agg"`
	HTTPCount string `db:"http_codes_count_agg"`

	model.Summary
}

type requestStatTable struct {
	SummaryUUID string `db:"summary_uuid"`

	model.RequestStat
}

type headerTable struct {
	ID int64 `db:"id"`

	Header            string `db:"header"`
	ConfigurationUUID string `db:"loader_uuid"`
}

type parameterTable struct {
	ID int64 `db:"id"`

	Parameters        string `db:"parameters"`
	ConfigurationUUID string `db:"loader_uuid"`
}

type errorsTable struct {
	ID int64 `db:"id"`

	Name  string `db:"name"`
	Count int    `db:"count"`

	Summary string `db:"summary"`
}

type aggregatedStatTable struct {
	ID          int64  `db:"id"`
	SummaryUUID string `db:"summary_uuid"`

	model.AggregatedStat
}

type httpCodesTable struct {
	ID int64 `db:"id"`

	Code  int `db:"code"`
	Count int `db:"count"`

	Summary string `db:"summary"`
}

type loaderTagTable struct {
	ID                      int64  `db:"id"`
	LoaderConfigurationUUID string `db:"loader_uuid"`

	model.LoaderTag
}

func Migrate(file string) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", d, "sqlite3://"+file)
	if err != nil {
		return err
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func NewStorage(file string) (*Storage, error) {
	db, err := sqlx.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("could not set foreign_keys pragma on: %w", err)
	}
	err = Migrate(file)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

// mapLoader function maps aggregated loaderConfiguration from database query to model.Loader
func mapLoader(loaderAgg []*loaderAggregated) ([]*model.Loader, error) {
	confs := make([]*model.Loader, 0)
	var err error
	for _, confAgg := range loaderAgg {
		if confAgg.Header != "" {
			header := make(model.Headers)
			headersSplit := strings.Split(confAgg.Header, ",")

			for _, h := range headersSplit {
				err = header.Set(h)
				if err != nil {
					return nil, err
				}
			}

			confAgg.Loader.Headers = header
		}

		if confAgg.Parameter != "" {
			parameter := make(model.Parameters, 0)
			parametersSplit := strings.Split(confAgg.Parameter, ",")

			for _, p := range parametersSplit {
				err = parameter.Set(p)
				if err != nil {
					return nil, err
				}
			}

			confAgg.Loader.Parameters = parameter
		}

		if confAgg.TagsKeys != "" && confAgg.TagsValues != "" {
			tagsKeys := strings.Split(confAgg.TagsKeys, ",")
			tagsValues := strings.Split(confAgg.TagsValues, ",")
			if len(tagsKeys) != len(tagsValues) {
				return nil, fmt.Errorf("error data looks broken")
			}

			tags := make([]*model.LoaderTag, len(tagsKeys))
			for i := 0; i < len(tagsKeys); i++ {
				tags[i] = &model.LoaderTag{
					Key:   tagsKeys[i],
					Value: tagsValues[i],
				}
			}

			confAgg.Loader.Tags = tags
		}

		confs = append(confs, &confAgg.Loader)
	}

	return confs, nil
}

// insertTablePrimaryUUID will perform a query to insert a table with primary key type TEXT and return it
// the query inserted here has to provide RETURNING with the uuid return
func (s *Storage) insertTablePrimaryUUID(tx *sqlx.Tx, query string, args any) (string, error) {
	var uuid string
	rows, err := tx.NamedQuery(query, args)
	if err != nil {
		return "", fmt.Errorf("error insert: %w", err)
	}

	rows.Next()
	err = rows.Scan(&uuid)
	if err != nil {
		return "", fmt.Errorf("error insert UUID result: %w", err)
	}

	if rows.Err() != nil {
		return "", fmt.Errorf("error insert: %w", err)
	}

	return uuid, nil
}

// insertTable insert table but do not care about the return
func (s *Storage) insertTable(tx *sqlx.Tx, query string, args any) error {
	_, err := tx.NamedExec(query, args)
	if err != nil {
		return fmt.Errorf("error insert: %w", err)
	}

	return nil
}
