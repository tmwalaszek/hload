package storage

import (
	"fmt"

	"github.com/tmwalaszek/hload/model"
)

func (s *Storage) DeleteLoader(ID string) error {
	_, err := s.db.Exec(deleteLoader, ID)
	return err
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

	for _, tag := range tags {
		_, err = tx.Exec(deleteLoaderTag, tag.Key, tag.Value, loaderUUID)
		if err != nil {
			return fmt.Errorf("delete error loader_tag: %w", err)
		}
	}

	err = tx.Commit()
	return err
}
