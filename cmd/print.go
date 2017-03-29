package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/ddn0/peanut/plog"
	"github.com/ghodss/yaml"
)

func print(obj interface{}, format, filter string) error {
	switch {
	case format == "text" && len(filter) != 0:
		if t, err := template.New("").Parse(filter); err != nil {
			return err
		} else if err := t.Execute(plog.Out, obj); err != nil {
			return err
		} else if _, err := fmt.Fprintf(plog.Out, "\n"); err != nil {
			return err
		}
	case format == "text" && len(filter) == 0:
		if _, err := fmt.Fprintf(plog.Out, "%+v\n", obj); err != nil {
			return err
		}
	case format == "json":
		if bs, err := json.Marshal(obj); err != nil {
			return err
		} else if _, err := fmt.Fprintln(plog.Out, string(bs)); err != nil {
			return err
		}
	case format == "yaml":
		if bs, err := yaml.Marshal(obj); err != nil {
			return err
		} else if _, err := fmt.Fprint(plog.Out, string(bs)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown output format: %q", format)
	}
	return nil
}
