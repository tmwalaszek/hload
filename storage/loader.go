package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	u "github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/tmwalaszek/hload/model"
)

type options struct {
	withRequests bool

	limit int

	from int64
	to   int64
}

type Option func(options *options)

func WithRequests() Option {
	return func(o *options) {
		o.withRequests = true
	}
}

func WithFrom(from int64) Option {
	return func(o *options) {
		o.from = from
	}
}

func WithTo(to int64) Option {
	return func(o *options) {
		o.to = to
	}
}

func WithLimit(limit int) Option {
	return func(o *options) {
		o.limit = limit
	}
}

func (s *Storage) DeleteLoader(ID string) error {
	_, err := s.db.Exec(deleteLoader, ID)
	return err
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

func (s *Storage) getSummaryWithAggRequests(summaryUUID string) ([]*model.AggregatedStat, error) {
	var aggregateStats []*model.AggregatedStat

	err := s.db.Select(&aggregateStats, selectAggregatedStats, summaryUUID)
	if err != nil {
		return nil, err
	}

	return aggregateStats, nil
}

func (s *Storage) getSummaryWithFullRequests(summaryUUID string) ([]*model.RequestStat, error) {
	var requestsStats []*model.RequestStat

	err := s.db.Select(&requestsStats, selectRequestsStats, summaryUUID)
	if err != nil {
		return nil, err
	}

	return requestsStats, nil
}

func (s *Storage) mapSummaries(summariesModelsAgg []*summaryAggregated) ([]*model.Summary, error) {
	summaries := make([]*model.Summary, 0)

	for _, summariesModel := range summariesModelsAgg {
		errs := make(map[string]int)
		httpCodes := make(map[int]int)

		if summariesModel.ErrorName != "" && summariesModel.ErrorCount != "" {
			names := strings.Split(summariesModel.ErrorName, ",")
			counts := strings.Split(summariesModel.ErrorCount, ",")

			if len(names) != len(counts) {
				return nil, fmt.Errorf("errors data looks broken, errors names should match errors counts")
			}

			for i := 0; i < len(names); i++ {
				count, err := strconv.Atoi(counts[i])
				if err != nil {
					return nil, err
				}

				errs[names[i]] = count
			}

			summariesModel.Errors = errs
		}

		if summariesModel.HTTPCount != "" && summariesModel.HTTPCode != "" {
			codes := strings.Split(summariesModel.HTTPCode, ",")
			counts := strings.Split(summariesModel.HTTPCount, ",")

			if len(codes) != len(counts) {
				return nil, fmt.Errorf("error data looks broken")
			}

			for i := 0; i < len(codes); i++ {
				code, err := strconv.Atoi(codes[i])
				if err != nil {
					return nil, err
				}

				count, err := strconv.Atoi(counts[i])
				if err != nil {
					return nil, err
				}

				httpCodes[code] = count
			}

			summariesModel.HTTPCodes = httpCodes
		}

		summaries = append(summaries, &summariesModel.Summary)
	}

	return summaries, nil
}

func (s *Storage) getSummariesRequests(summaries []*model.Summary) error {
	for _, summary := range summaries {
		requests, err := s.getSummaryWithFullRequests(summary.UUID)
		if err != nil {
			return err
		}

		aggregatedRequests, err := s.getSummaryWithAggRequests(summary.UUID)
		if err != nil {
			return err
		}

		summary.AggregatedStats = aggregatedRequests
		summary.RequestStats = requests
	}

	return nil
}

func (s *Storage) GetLoaderByDescription(description string) ([]*model.Loader, error) {
	var loaderConfAgg []*loaderAggregated

	sqlQuery, err := generateSQLFromTemplate(loaderConfigurationTemplate, "by_loader.description", nil)
	if err != nil {
		return nil, err
	}

	err = s.db.Select(&loaderConfAgg, sqlQuery, description)
	if err != nil {
		return nil, err
	}

	opts, err := mapLoader(loaderConfAgg)
	if err != nil {
		return nil, err
	}

	return opts, nil
}

func (s *Storage) GetLoaderByName(name string) ([]*model.Loader, error) {
	var loaderConfAgg []*loaderAggregated

	sqlQuery, err := generateSQLFromTemplate(loaderConfigurationTemplate, "by_loader.name", nil)
	if err != nil {
		return nil, err
	}

	err = s.db.Select(&loaderConfAgg, sqlQuery, name)
	if err != nil {
		return nil, err
	}

	opts, err := mapLoader(loaderConfAgg)
	if err != nil {
		return nil, err
	}

	return opts, nil
}

func (s *Storage) GetLoaderByID(loaderUUID string) (*model.Loader, error) {
	var loaderConfAgg []*loaderAggregated

	sqlQuery, err := generateSQLFromTemplate(loaderConfigurationTemplate, "by_loader.id", nil)
	if err != nil {
		return nil, err
	}

	err = s.db.Select(&loaderConfAgg, sqlQuery, loaderUUID)
	if err != nil {
		return nil, err
	}

	confs, err := mapLoader(loaderConfAgg)
	if err != nil {
		return nil, err
	}

	if len(confs) == 0 {
		return nil, fmt.Errorf("loader configuration %s not found", loaderUUID)
	}

	return confs[0], nil
}

func (s *Storage) GetLoaders(limit int) ([]*model.Loader, error) {
	var loaderConfAgg []*loaderAggregated
	sqlQuery, err := generateSQLFromTemplate(loaderConfigurationTemplate, "limit", nil)
	if err != nil {
		return nil, err
	}

	err = s.db.Select(&loaderConfAgg, sqlQuery, limit)
	if err != nil {
		return nil, err
	}

	confUUIDs, err := mapLoader(loaderConfAgg)
	if err != nil {
		return nil, err
	}

	return confUUIDs, nil
}

func (s *Storage) GetLoadersByRange(from, to int64, limit int) ([]*model.Loader, error) {
	sqlQuery, err := generateSQLFromTemplate(loaderConfigurationTemplate, "by_time_scope", nil)
	if err != nil {
		return nil, err
	}

	var loaderConfAgg []*loaderAggregated
	err = s.db.Select(&loaderConfAgg, sqlQuery, from, to, limit)
	if err != nil {
		return nil, err
	}

	loaders, err := mapLoader(loaderConfAgg)
	if err != nil {
		return nil, err
	}
	return loaders, nil
}

func (s *Storage) GetSummaries(loaderConfUUID string, opts ...Option) ([]*model.Summary, error) {
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	var summariesModelsAgg []*summaryAggregated

	if options.from != 0 {
		sqlQuery, err := generateSQLFromTemplate(summaryTemplate, "range", nil)
		if err != nil {
			return nil, err
		}

		err = s.db.Select(&summariesModelsAgg, sqlQuery, loaderConfUUID, options.from, options.to, options.limit)
		if err != nil {
			return nil, err
		}
	} else {
		sqlQuery, err := generateSQLFromTemplate(summaryTemplate, "loader_uuid", nil)
		if err != nil {
			return nil, err
		}

		err = s.db.Select(&summariesModelsAgg, sqlQuery, loaderConfUUID, options.limit)
		if err != nil {
			return nil, err
		}
	}

	summaries, err := s.mapSummaries(summariesModelsAgg)
	if err != nil {
		return nil, err
	}

	if options.withRequests {
		err = s.getSummariesRequests(summaries)
		if err != nil {
			return nil, err
		}
	}

	return summaries, nil
}

func (s *Storage) GetLoaderByTags(tags []*model.LoaderTag) ([]*model.Loader, error) {
	type TagsKeysValues struct {
		Keys   []string
		Values []string
	}

	var loaderConfAgg []*loaderAggregated

	var keys, values []string
	for _, tag := range tags {
		keys = append(keys, tag.Key)
		if len(tag.Value) > 0 {
			values = append(values, tag.Value)
		}
	}

	kvTags := TagsKeysValues{
		Keys:   keys,
		Values: values,
	}

	sqlQuery, err := generateSQLFromTemplate(loaderConfigurationTemplate, "by_loader.tag", kvTags)
	if err != nil {
		return nil, err
	}

	err = s.db.Select(&loaderConfAgg, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	optIDs, err := mapLoader(loaderConfAgg)
	return optIDs, err
}
