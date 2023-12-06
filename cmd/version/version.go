package version

import (
	"encoding/json"
	"fmt"

	"github.com/tmwalaszek/hload/cmd/cliio"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type OutputValue string

const (
	OutputVarJSON OutputValue = "json"
	OutputVarYAML OutputValue = "yaml"
)

var (
	Major     string
	Minor     string
	Patch     string
	GitCommit string
	BuildDate string
	GoVersion string
)

type VersionInfo struct {
	Major     string `json:"major"`
	Minor     string `json:"minor"`
	Patch     string `json:"patch"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

type Options struct {
	Short  bool
	Output Output

	cliio.IO
}

type Output struct {
	Value OutputValue
}

func (o *Output) String() string {
	return string(o.Value)
}

func (o *Output) Set(value string) error {
	switch v := value; v {
	case "json":
		o.Value = OutputVarJSON
	case "yaml":
		o.Value = OutputVarYAML
	default:
		return fmt.Errorf("unknwon value: %s", v)
	}
	return nil
}

func (o *Output) Type() string {
	return "output"
}

func NewVersionCmd(cliIO cliio.IO) *cobra.Command {
	opts := Options{
		IO: cliIO,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version",
		Long:  "Print the version of the CLI",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Run()
		},
	}

	cmd.Flags().BoolVarP(&opts.Short, "short", "s", opts.Short, "Show only the CLI version")
	cmd.Flags().VarP(&opts.Output, "output", "o", "Out in one of 'yaml' or 'json'")
	return cmd
}

func (o *Options) Run() {
	versionInfo := VersionInfo{
		Major:     Major,
		Minor:     Minor,
		Patch:     Patch,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
	}

	if o.Short {
		fmt.Fprintf(o.Out, "HLoad version %s", fmt.Sprintf("%s.%s.%s", Major, Minor, Patch))
		return
	}

	var b []byte
	var err error

	switch v := o.Output; v.Value {
	case OutputVarJSON:
		b, err = json.MarshalIndent(versionInfo, "", " ")
	case OutputVarYAML:
		b, err = yaml.Marshal(versionInfo)
	default:
		b, err = json.MarshalIndent(versionInfo, "", " ")
	}

	if err != nil {
		fmt.Fprintf(o.Out, "error: %v\n", err)
		return
	}

	fmt.Fprintf(o.Out, "%s\n", string(b))
}
