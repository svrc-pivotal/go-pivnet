package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/olekukonko/tablewriter"
)

type ReleaseTypesCommand struct {
}

func (command *ReleaseTypesCommand) Execute([]string) error {
	client := NewClient()
	releaseTypes, err := client.ReleaseTypes()
	if err != nil {
		return err
	}

	switch Pivnet.PrintAs {
	case printAsTable:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ReleaseTypes"})

		for _, r := range releaseTypes {
			table.Append([]string{r})
		}
		table.Render()
		return nil
	case printAsJSON:
		b, err := json.Marshal(releaseTypes)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", string(b))
		return nil
	case printAsYAML:
		b, err := yaml.Marshal(releaseTypes)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", string(b))
		return nil
	}

	return nil
}
