package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tmwalaszek/hload/model"

	u "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

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

func (s *Storage) InsertLoaderConfigurationTags(loaderConfUUID string, loaderConfigurationTags []*model.LoaderTag) (err error) {
	tx := s.db.MustBegin()
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = StorageError{
					Err:           err,
					RollbackError: rollbackErr,
				}
			}
			return
		}
	}()

	for _, t := range loaderConfigurationTags {
		tag := loaderTagTable{
			LoaderConfigurationUUID: loaderConfUUID,
			LoaderTag:               *t,
		}

		err = s.insertTable(tx, loaderConfigurationTagInsert, tag)
		if err != nil {
			return fmt.Errorf("insert error loader_tag: %w", err)
		}
	}

	err = tx.Commit()

	return
}

// InsertLoaderConfiguration method is saving Loader structure in the database
// It will create parameters and headers in separate tables
func (s *Storage) InsertLoaderConfiguration(loaderConfiguration *model.Loader) (uuid string, err error) {
	tx := s.db.MustBegin()
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = StorageError{
					Err:           err,
					RollbackError: rollbackErr,
				}
			}
			return
		}
	}()

	if loaderConfiguration.UUID == "" {
		id := u.New()
		loaderConfiguration.UUID = id.String()
	}

	uuid, err = s.insertTablePrimaryUUID(tx, optsInsert, loaderConfiguration)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
				return "", fmt.Errorf("loader configuration name %s for URL %s already exists", loaderConfiguration.Name, loaderConfiguration.URL)
			}
		}
		return "", err
	}

	loaderConfiguration.LoaderReqDetails.LoaderConfigurationUUID = uuid
	err = s.insertTable(tx, optsLoadInsert, loaderConfiguration.LoaderReqDetails)
	if err != nil {
		return "", err
	}

	if len(loaderConfiguration.Headers) > 0 {
		headers := make([]string, 0)
		for k, values := range loaderConfiguration.Headers {
			for _, v := range values {
				header := strings.Join([]string{k, v}, ":")
				headers = append(headers, header)
			}
		}

		for _, header := range headers {
			headerModel := &headerTable{
				ConfigurationUUID: uuid,
				Header:            header,
			}

			err = s.insertTable(tx, headerInsert, headerModel)
			if err != nil {
				return "", err
			}
		}
	}

	if len(loaderConfiguration.Parameters) > 0 {
		parameters := make([]string, len(loaderConfiguration.Parameters))
		for i, parameterMap := range loaderConfiguration.Parameters {
			var parameter string
			for k, v := range parameterMap {
				parameter += strings.Join([]string{k, v}, "=")
			}

			parameters[i] = parameter
		}

		for _, parameter := range parameters {
			parameterModel := &parameterTable{
				ConfigurationUUID: uuid,
				Parameters:        parameter,
			}

			err = s.insertTable(tx, parameterInsert, parameterModel)
			if err != nil {
				return "", err
			}
		}
	}

	err = tx.Commit()

	return uuid, err
}

func (s *Storage) InsertSummary(optsUUID string, summary *model.Summary, saveRequests, saveAggRequests bool) (uuid string, err error) {
	summary.LoaderConf = optsUUID

	tx := s.db.MustBegin()
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = StorageError{
					Err:           err,
					RollbackError: rollbackErr,
				}
			}
			return
		}
	}()

	id := u.New()
	summary.UUID = id.String()

	uuid, err = s.insertTablePrimaryUUID(tx, summaryInsert, summary)
	if err != nil {
		return "", err
	}

	for errName, errCount := range summary.Errors {
		errModel := errorsTable{
			Name:    errName,
			Count:   errCount,
			Summary: uuid,
		}

		err = s.insertTable(tx, errInsert, errModel)
		if err != nil {
			return "", err
		}
	}

	for code, count := range summary.HTTPCodes {
		httpCodeModel := httpCodesTable{
			Code:    code,
			Count:   count,
			Summary: uuid,
		}

		err = s.insertTable(tx, httpCodesInsert, httpCodeModel)
		if err != nil {
			return "", err
		}
	}

	if saveRequests {
		for _, reqStat := range summary.RequestStats {
			reqStatDB := requestStatTable{
				SummaryUUID: uuid,
				RequestStat: model.RequestStat{
					Start:    reqStat.Start,
					End:      reqStat.End,
					Duration: reqStat.Duration,
					Error:    reqStat.Error,
					BodySize: reqStat.BodySize,
					RetCode:  reqStat.RetCode,
				},
			}

			err = s.insertTable(tx, insertRequestStat, reqStatDB)
			if err != nil {
				return "", err
			}
		}
	}

	if saveAggRequests {
		for _, aggStat := range summary.AggregatedStats {
			aggStatDB := aggregatedStatTable{
				SummaryUUID: uuid,
				AggregatedStat: model.AggregatedStat{
					Start:          aggStat.Start,
					End:            aggStat.End,
					Duration:       aggStat.Duration,
					AvgRequestTime: aggStat.AvgRequestTime,
					MaxRequestTime: aggStat.MaxRequestTime,
					MinRequestTime: aggStat.MinRequestTime,
					RequestCount:   aggStat.RequestCount,
				},
			}
			err = s.insertTable(tx, insertAggregateStat, aggStatDB)
			if err != nil {
				return "", err
			}
		}
	}

	err = tx.Commit()
	return uuid, err
}
