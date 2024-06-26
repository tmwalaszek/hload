package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/tmwalaszek/hload/model"
)

func (s *Storage) UpdateLoaderTag(loaderUUID string, key, value string) error {
	_, err := s.db.Exec(updateLoaderTag, value, key, loaderUUID)
	if err != nil {
		return fmt.Errorf("update loader tag: %w", err)
	}

	return nil
}

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

	var deletedCount int64
	for _, tag := range tags {
		res, err := tx.Exec(deleteLoaderTag, tag.Key, tag.Value, loaderUUID)
		if err != nil {
			return fmt.Errorf("delete error loader_tag: %w", err)
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("rows affected error: %w", err)
		}

		deletedCount += rows
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if deletedCount == 0 {
		return fmt.Errorf("loader tags not found")
	}

	return nil
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
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				if errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
					if strings.Contains(sqliteErr.Error(), "UNIQUE constraint failed") {
						return errors.New("tag for the loader UUID already exists")
					} else if strings.Contains(sqliteErr.Error(), "FOREIGN KEY constraint failed") {
						return errors.New("loader UUID does not exists")
					}
				}
			}

			return fmt.Errorf("insert error loader_tag: %w", err)
		}
	}

	err = tx.Commit()

	return
}
