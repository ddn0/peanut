package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ddn0/peanut/git"
	"github.com/ddn0/peanut/logwriter"
	"github.com/ddn0/peanut/pdo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch remote git repo data",
	RunE:  runFetch,
}

func fetch(ctx context.Context, item interface{}) error {
	dir := item.(string)

	lw := logwriter.NewColorWriter(filepath.Base(dir))
	defer lw.Flush()
	cmd := exec.Command("git", "fetch", "--all", "--prune")
	cmd.Dir = dir
	cmd.Stdout = lw
	cmd.Stderr = lw

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	if viper.GetBool("verbose") {
		fmt.Fprintln(lw, "fetching")
	}

	errs := make(chan error, 1)
	go func() {
		errs <- cmd.Run()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errs:
		return err
	}
}

func runFetch(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	cfg, err := readConf()
	if err != nil {
		return err
	}

	gc := git.NewClient(nil)

	seen := make(map[string]bool)
	var dirs []interface{}
	for _, dir := range cfg.RepoPaths() {
		wt, err := gc.WorkTree(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: error reading git work tree of %q: %s\n", dir, err)
			return err
		}
		if seen[wt.Repo] {
			continue
		}
		seen[wt.Repo] = true
		dirs = append(dirs, wt.Repo)
	}

	if err := pdo.DoAll(pdo.DoAllOpt{
		Func:          fetch,
		Items:         dirs,
		Timeout:       viper.GetDuration("timeout"),
		MaxConcurrent: viper.GetInt("max-concurrent"),
	}); err != nil {
		return err
	}
	return nil
}

func init() {
	c := fetchCmd
	//flags := c.Flags()

	RootCmd.AddCommand(c)
}
