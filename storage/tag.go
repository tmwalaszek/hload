package storage

import (
	"fmt"

	"github.com/tmwalaszek/hload/model"
)

func (s *Storage) DeleteLoaderTag(loaderUUID string, tags []*model.LoaderTag) (err error) {
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

	for _, tag := range tags {
		_, err = tx.Exec(deleteLoaderTag, tag.Key, tag.Value, loaderUUID)
		if err != nil {
			return fmt.Errorf("delete error loader_tag: %w", err)
		}
	}

	err = tx.Commit()
	return err
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
