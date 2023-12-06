package storage

import (
	"fmt"
	"strconv"
	"strings"

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
		return nil, nil
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

func (s *Storage) GetLoaderTagsByKey(key string) (map[string]*model.LoaderTag, error) {
	var loaderConfTags []*loaderTagTable

	err := s.db.Select(&loaderConfTags, selectLoaderTagsByName, key)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]*model.LoaderTag)
	for _, tag := range loaderConfTags {
		tags[tag.LoaderConfigurationUUID] = &tag.LoaderTag
	}

	return tags, nil
}

func (s *Storage) GetLoaderTags(loaderUUID string) ([]*model.LoaderTag, error) {
	var loaderConfTags []*loaderTagTable

	err := s.db.Select(&loaderConfTags, selectLoaderTags, loaderUUID)
	if err != nil {
		return nil, err
	}

	tags := make([]*model.LoaderTag, 0)
	for _, tag := range loaderConfTags {
		tags = append(tags, &tag.LoaderTag)
	}

	return tags, nil
}

func (s *Storage) GetLoaderByTags(tags []*model.LoaderTag) ([]*model.Loader, error) {
	type wrappedTags struct {
		Tags []*model.LoaderTag
	}

	var loaderConfAgg []*loaderAggregated

	wTags := wrappedTags{
		Tags: tags,
	}

	sqlQuery, err := generateSQLFromTemplate(loaderConfigurationTemplate, "by_loader.tag", wTags)
	if err != nil {
		return nil, err
	}

	args := make([]any, len(tags)*2)
	i := 0
	for _, tag := range tags {
		args[i] = tag.Key
		i++
		args[i] = tag.Value
		i++
	}

	err = s.db.Select(&loaderConfAgg, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	optIDs, err := mapLoader(loaderConfAgg)
	return optIDs, err
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
