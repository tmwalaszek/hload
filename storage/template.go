package storage

import (
	"fmt"

	"github.com/tmwalaszek/hload/model"
)

func (s *Storage) DeleteTemplate(name string) error {
	_, err := s.db.Exec(deleteTemplate, name)
	return err
}

func (s *Storage) GetTemplates(limit int) ([]*model.Template, error) {
	sqlQuery, err := generateSQLFromTemplate(selectTemplates, "all", nil)
	if err != nil {
		return nil, err
	}

	var templates []*model.Template

	err = s.db.Select(&templates, sqlQuery, limit)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

func (s *Storage) GetTemplateByName(name string) (*model.Template, error) {
	sqlQuery, err := generateSQLFromTemplate(selectTemplates, "by_name", nil)
	if err != nil {
		return nil, err
	}

	var template model.Template

	err = s.db.Get(&template, sqlQuery, name)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (s *Storage) InsertTemplate(name, content string) error {
	template := model.Template{
		Name:    name,
		Content: content,
	}

	_, err := s.db.NamedExec(insertTemplate, template)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateTemplate(name, content string) error {
	res, err := s.db.Exec(updateTemplate, content, name)
	if err != nil {
		return fmt.Errorf("could not update template: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not update template: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("could not update template: no rows updated")
	}

	return nil
}
