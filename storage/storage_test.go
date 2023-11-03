package storage

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/tmwalaszek/hload/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type tempSummary struct {
	model.Summary

	Start string `json:"start"`
	End   string `json:"end"`
}

func TestStorageTags(t *testing.T) {
	testDir := "testdata/tags"
	testOpt := testDir + "/loader.json"
	testTags := testDir + "/tags.json"

	optsBytes, err := os.ReadFile(testOpt)
	require.Nil(t, err)

	tagsBytes, err := os.ReadFile(testTags)
	require.Nil(t, err)

	loaderOpts := &model.Loader{}
	err = json.Unmarshal(optsBytes, loaderOpts)
	require.Nil(t, err)

	tags := make([]*model.LoaderTag, 0)
	err = json.Unmarshal(tagsBytes, &tags)
	require.Nil(t, err)

	loaderOpts.Tags = tags

	uid, err := uuid.NewUUID()
	require.Nil(t, err)
	loaderOpts.UUID = uid.String()
	loaderOpts.ID = 1

	store, err := NewStorage("test_file.db")
	defer os.Remove("test_file.db")

	require.Nil(t, err)

	_, err = store.InsertLoaderConfiguration(loaderOpts)
	require.Nil(t, err)

	err = store.InsertLoaderConfigurationTags(uid.String(), tags)
	require.Nil(t, err)

	now := time.Now()
	format := "2006-01-02 15:04"
	timeOnly := now.UTC().Format(format)
	parsedTime, err := time.Parse(format, timeOnly)
	require.Nil(t, err)

	loaders, err := store.GetLoaderByTags(tags)
	require.Nil(t, err)
	require.Len(t, loaders, 1)
	require.Equal(t, parsedTime, loaders[0].CreateDate.Truncate(time.Minute))
	loaderOpts.CreateDate = loaders[0].CreateDate
	require.Equal(t, loaderOpts, loaders[0])
}

// One LoaderConf and zero, one or more summaries
func TestStorageOneLoaderOpts(t *testing.T) {
	var tt = []struct {
		Name        string
		Directory   string
		LoaderFile  string
		SummaryFile string
	}{
		{
			Name:        "Test 1 - directory opt1",
			Directory:   "opt1",
			LoaderFile:  "loader.json",
			SummaryFile: "summaries.json",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			store, err := NewStorage("test_file.db")
			defer os.Remove("test_file.db")

			require.Nil(t, err)

			testDir := "testdata/" + tc.Directory
			testOpt := testDir + "/" + tc.LoaderFile
			testSummary := testDir + "/" + tc.SummaryFile
			loaderOpts := &model.Loader{}
			var summaries []*model.Summary

			optsBytes, err := os.ReadFile(testOpt)
			require.Nil(t, err)
			summariesBytes, err := os.ReadFile(testSummary)
			require.Nil(t, err)

			err = json.Unmarshal(optsBytes, loaderOpts)
			require.Nil(t, err)

			tempSummaries := make([]*tempSummary, 0)
			err = json.Unmarshal(summariesBytes, &tempSummaries)
			require.Nil(t, err)

			timeFormat := "2006-01-02 15:04"
			for _, tempSummary := range tempSummaries {
				start, err := time.Parse(timeFormat, tempSummary.Start)
				require.Nil(t, err)
				end, err := time.Parse(timeFormat, tempSummary.End)
				require.Nil(t, err)

				summary := &model.Summary{
					URL:             tempSummary.URL,
					Description:     tempSummary.Description,
					Start:           start,
					End:             end,
					TotalTime:       tempSummary.TotalTime,
					ReqCount:        tempSummary.ReqCount,
					SuccessReq:      tempSummary.SuccessReq,
					FailReq:         tempSummary.FailReq,
					DataTransferred: tempSummary.DataTransferred,
					ReqPerSec:       tempSummary.ReqPerSec,
					AvgReqTime:      tempSummary.AvgReqTime,
					MinReqTime:      tempSummary.MinReqTime,
					MaxReqTime:      tempSummary.MaxReqTime,
					P50ReqTime:      tempSummary.P50ReqTime,
					P75ReqTime:      tempSummary.P75ReqTime,
					P90ReqTime:      tempSummary.P90ReqTime,
					P99ReqTime:      tempSummary.P99ReqTime,
					StdDeviation:    tempSummary.StdDeviation,
					Errors:          tempSummary.Errors,
					HTTPCodes:       tempSummary.HTTPCodes,
					AggregatedStats: tempSummary.AggregatedStats,
					RequestStats:    tempSummary.RequestStats,
				}
				summaries = append(summaries, summary)
			}

			uid, err := uuid.NewUUID()
			require.Nil(t, err)
			loaderOpts.UUID = uid.String()
			loaderOpts.ID = 1
			u, err := store.InsertLoaderConfiguration(loaderOpts)
			require.Nil(t, err)

			for i, summary := range summaries {
				u, err := store.InsertSummary(u, summary, false, false)
				require.Nil(t, err)
				summaries[i].UUID = u
			}

			now := time.Now()
			format := "2006-01-02 15:04"
			timeOnly := now.UTC().Format(format)
			parsedTime, err := time.Parse(format, timeOnly)

			// --- Read the data from the database and compare with the JSONs we inserted
			loader, err := store.GetLoaderByID(loaderOpts.UUID)
			require.Nil(t, err)
			require.Equal(t, parsedTime, loader.CreateDate.Truncate(time.Minute))
			loaderOpts.CreateDate = loader.CreateDate
			require.Equal(t, loaderOpts, loader)

			loaders, err := store.GetLoaders(10)
			require.Nil(t, err)
			require.Len(t, loaders, 1)
			require.Equal(t, parsedTime, loaders[0].CreateDate.Truncate(time.Minute))
			loaderOpts.CreateDate = loaders[0].CreateDate
			require.Equal(t, loaderOpts, loaders[0])

			loaders, err = store.GetLoaderByDescription(loaderOpts.Description)
			require.Nil(t, err)
			require.Len(t, loaders, 1)
			require.Equal(t, parsedTime, loaders[0].CreateDate.Truncate(time.Minute))
			loaderOpts.CreateDate = loaders[0].CreateDate
			require.Equal(t, loaderOpts, loaders[0])

			loaders, err = store.GetLoaderByName(loaderOpts.Name)
			require.Nil(t, err)
			require.Len(t, loaders, 1)
			require.Equal(t, parsedTime, loaders[0].CreateDate.Truncate(time.Minute))
			loaderOpts.CreateDate = loaders[0].CreateDate
			require.Equal(t, loaderOpts, loaders[0])

			loaderWithSummaries, err := store.GetSummaries(loaderOpts.UUID, WithLimit(10), WithRequests())
			require.Nil(t, err)
			require.NotNil(t, loaderWithSummaries)
			require.EqualValues(t, summaries, loaderWithSummaries)

			loaderWithSummaries, err = store.GetSummaries(loaderOpts.UUID, WithLimit(10))
			require.Nil(t, err)
			require.NotNil(t, loaderWithSummaries)
			require.EqualValues(t, summaries, loaderWithSummaries)
		})
	}
}
