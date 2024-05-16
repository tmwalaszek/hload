package storage

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	//go:embed sql/insert_loader.sql
	optsInsert string
	//go:embed sql/insert_loader_requests_details.sql
	optsLoadInsert string
	//go:embed sql/insert_header.sql
	headerInsert string
	//go:embed sql/insert_parameter.sql
	parameterInsert string
	//go:embed sql/insert_error.sql
	errInsert string
	//go:embed sql/insert_http_codes.sql
	httpCodesInsert string
	//go:embed sql/insert_summary.sql
	summaryInsert string
	//go:embed sql/insert_loader_tags.sql
	loaderConfigurationTagInsert string
	//go:embed sql/delete_loader.sql
	deleteLoader string
	//go:embed sql/delete_loader_tag.sql
	deleteLoaderTag string
	//go:embed sql/insert_aggregate_stats.sql
	insertAggregateStat string
	//go:embed sql/insert_request_stats.sql
	insertRequestStat string
	//go:embed sql/select_aggregated_stats.sql
	selectAggregatedStats string
	//go:embed sql/select_requests_stats.sql
	selectRequestsStats string
	//go:embed sql/select_loader_tags.sql
	selectLoaderTags string
	//go:embed sql/select_loader_tags_by_name.sql
	selectLoaderTagsByName string
	//go:embed sql/select_summaries.tmpl
	summaryTemplate string
	//go:embed sql/select_loader.tmpl
	loaderConfigurationTemplate string
	//go:embed sql/insert_template.sql
	insertTemplate string
	//go:embed sql/select_temaples.tmpl
	selectTemplates string
	//go:embed sql/delete_template.sql
	deleteTemplate string
	//go:embed sql/update_template.sql
	updateTemplate string
	//go:embed sql/update_loader_tag.sql
	updateLoaderTag string
)

// data is optional depending on the template
func generateSQLFromTemplate(tmplFile, tmpl string, data any) (string, error) {
	var buff bytes.Buffer

	funcAdd := template.FuncMap{
		"add": func(x, y int) int {
			return x + y
		},
		"bold": func(x string) string {
			return text.Bold.Sprint(x)
		},
	}

	t := template.Must(template.New("opts").Funcs(funcAdd).Parse(tmplFile))
	err := t.ExecuteTemplate(&buff, tmpl, data)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buff.String(), nil
}
