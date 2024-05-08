package templates

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/tmwalaszek/hload/model"
	"github.com/tmwalaszek/hload/storage"
)

var (
	//go:embed default/list_template.tmpl
	ListTemplate string
)

type LoaderSummaries struct {
	Loader    *model.Loader    `json:"loader"`
	Summaries []*model.Summary `json:"Summaries"`
}

type Loaders struct {
	Loaders []LoaderSummaries

	Short               bool
	ShowFullStats       bool
	ShowAggregatedStats bool
}

type RenderTemplate struct {
	content string
}

func NewRenderTemplate(name, dbFile string) (*RenderTemplate, error) {
	var content string
	if name == "default" {
		content = ListTemplate
	} else {
		s, err := storage.NewStorage(dbFile)
		if err != nil {
			return nil, fmt.Errorf("error creating new storage: %v", err)
		}

		templ, err := s.GetTemplateByName(name)
		if err != nil {
			return nil, fmt.Errorf("error loading template: %v", err)
		}

		if templ == nil {
			return nil, fmt.Errorf("error loading template: template not found")
		}

		content = templ.Content
	}

	return &RenderTemplate{content}, nil
}

func (r *RenderTemplate) render(templateName string, data any) ([]byte, error) {
	var buff bytes.Buffer

	funcsAdd := template.FuncMap{
		"bold": func(x string) string {
			return text.Bold.Sprint(x)
		},
		"timeInLoc": func(x time.Time) time.Time {
			loc, err := time.LoadLocation("Local")
			if err != nil {
				log.Fatal(err)
			}

			return x.In(loc)
		},
	}

	t := template.Must(template.New("new").Funcs(funcsAdd).Parse(r.content))
	err := t.ExecuteTemplate(&buff, templateName, data)
	return buff.Bytes(), err
}

func (r *RenderTemplate) RenderOutput(loaders *Loaders) ([]byte, error) {
	return r.render("loaders", loaders)
}

func (r *RenderTemplate) RenderTags(loaderUUID string, tags []*model.LoaderTag) ([]byte, error) {
	var s = struct {
		LoaderUUID string
		Tags       []*model.LoaderTag
	}{
		LoaderUUID: loaderUUID,
		Tags:       tags,
	}

	return r.render("tags", s)
}

func (r *RenderTemplate) RenderTagsMap(tags map[string]*model.LoaderTag) ([]byte, error) {
	return r.render("tags_map", tags)
}

func (r *RenderTemplate) RenderSummary(summary *model.Summary, fullStats, aggStats bool) ([]byte, error) {
	l := &Loaders{
		Loaders: []LoaderSummaries{
			{
				Summaries: []*model.Summary{summary},
			},
		},
		ShowFullStats:       fullStats,
		ShowAggregatedStats: aggStats,
	}

	return r.render("summaries", l)
}
