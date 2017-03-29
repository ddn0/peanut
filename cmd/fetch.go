package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

func mergedBranches(out io.Writer, dir, root string) ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", "branch", "--merged", root)
	cmd.Dir = dir
	cmd.Stdout = &buf
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var branches []string
	for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
		bs := bytes.TrimSpace(line)
		if len(bs) == 0 {
			continue
		}
		if bs[0] == '*' {
			continue
		}
		branches = append(branches, string(bs))
	}
	return branches, nil
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
		MaxConcurrent: viper.GetInt("maxConcurrent"),
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
