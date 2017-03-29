package cmd

import (
	"fmt"
	"os"

	"github.com/ddn0/peanut/config"
	"github.com/ddn0/peanut/git"
	"github.com/ddn0/peanut/logwriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addDirCmd = &cobra.Command{
	Use:   "add-dir [dirs]",
	Short: "add directory to config",
	RunE:  runAddDir,
}

func runAddDir(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	cfg, err := readConf()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		args = append(args, wd)
	}

	seen := make(map[string]bool)
	for _, dir := range cfg.RepoPaths() {
		seen[dir] = true
	}

	lw := logwriter.NewColorWriter("")
	defer lw.Flush()
	client := git.NewClient(nil)
	for _, arg := range args {
		wt, err := client.WorkTree(arg)
		if err != nil {
			fmt.Fprintf(lw, "[warn] error adding %s: %s", arg, err)
			continue
		}
		if seen[wt.Repo] {
			continue
		}
		seen[wt.Repo] = true
		cfg.Repos = append(cfg.Repos, &config.Repo{
			Path: wt.Repo,
		})
	}

	return nil
}

func init() {
	c := addDirCmd
	//flags := c.Flags()

	RootCmd.AddCommand(c)
}
