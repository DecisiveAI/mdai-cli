package cmd

import (
	"errors"
	"testing"
)

func TestFilterAddCommandErr(t *testing.T) {
	errTests := testCmdErrs{
		{
			name: "filter add command with both pipeline and telemetry flags",
			args: []string{"filter", "add", "--service", "foo", "--name", "foo", "--description", "foo", "--pipeline", "traces", "--telemetry", "traces"},
			err:  errors.New("if any flags in the group [pipeline telemetry] are set none of the others can be; [pipeline telemetry] were all set"),
		},
		{
			name: "filter add command with no flags",
			args: []string{"filter", "add"},
			err:  errors.New(filterAddUsage),
		},
		{
			name: "filter add command with name flag without description and pipeline flags",
			args: []string{"filter", "add", "--name", "test-filter"},
			err:  errors.New(`if any flags in the group [name description pipeline] are set they must all be set; missing [description pipeline]`),
		},
		{
			name: "filter add command with name and description flags without pipeline flag",
			args: []string{"filter", "add", "--name", "test-filter", "--description", "test filter"},
			err:  errors.New(`if any flags in the group [name description pipeline] are set they must all be set; missing [pipeline]`),
		},
		{
			name: "filter add command with name and pipeline flags without description flag",
			args: []string{"filter", "add", "--name", "test-filter", "--pipeline", "logs"},
			err:  errors.New(`if any flags in the group [name description pipeline] are set they must all be set; missing [description]`),
		},
		{
			name: "filter add command with description flag without name and pipeline flags",
			args: []string{"filter", "add", "--description", "test-filter"},
			err:  errors.New(`if any flags in the group [name description pipeline] are set they must all be set; missing [name pipeline]`),
		},
		{
			name: "filter add command with description and pipeline flags without name flag",
			args: []string{"filter", "add", "--description", "test-filter", "--pipeline", "logs"},
			err:  errors.New(`if any flags in the group [name description pipeline] are set they must all be set; missing [name]`),
		},
		{
			name: "filter add command with pipeline flag without description and name flags",
			args: []string{"filter", "add", "--pipeline", "logs"},
			err:  errors.New(`if any flags in the group [name description pipeline] are set they must all be set; missing [description name]`),
		},

		{
			name: "filter add command with service flag without description and name flags",
			args: []string{"filter", "add", "--service", "foo"},
			err:  errors.New(`if any flags in the group [name description service] are set they must all be set; missing [description name]`),
		},
		{
			name: "filter add command with description and service flags without name flag",
			args: []string{"filter", "add", "--description", "test-filter", "--service", "logs"},
			err:  errors.New(`if any flags in the group [name description service] are set they must all be set; missing [name]`),
		},
		{
			name: "filter add command with name and service flags without description flag",
			args: []string{"filter", "add", "--name", "test-filter", "--service", "logs"},
			err:  errors.New(`if any flags in the group [name description service] are set they must all be set; missing [description]`),
		},
	}

	errTests.Run(t)
}

var filterAddUsage = `Usage:
  mdai filter add

Examples:
  add --name filter-1 --description filter-1 --pipeline logs
  add --name filter-1 --description filter-1 --pipeline logs --service service-1
  add --name filter-1 --description filter-1 --telemetry logs --service service-1

Flags:
  -d, --description string   description of the filter
  -h, --help                 help for add
  -n, --name string          name of the filter
  -p, --pipeline strings     pipeline to mute
  -s, --service string       service pattern
  -t, --telemetry strings    telemetry type

Global Flags:
      --kubeconfig string    Path to a kubeconfig
      --kubecontext string   Kubernetes context to use
`
