package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddn0/peanut/git"
	"github.com/ddn0/peanut/logwriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge outstanding commits with working directories",
	RunE:  runMerge,
}

func pruneLocal(dir string, roots []string) error {
	lw := logwriter.NewColorWriter(filepath.Base(dir))
	defer lw.Flush()

	var rs []string
	for _, r := range roots {
		cmd := exec.Command("git", "rev-parse", "--verify", r)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			// Branch does not exist
			continue
		}
		rs = append(rs, r)
	}

	branches := make(map[string]int)
	for _, r := range rs {
		bs, err := mergedBranches(lw, dir, r)
		if err != nil {
			return err
		}
		for _, b := range bs {
			branches[b] += 1
		}
	}
	// Do not process roots
	for _, r := range rs {
		delete(branches, r)
	}

	for b, c := range branches {
		// Is branch merged in all roots?
		if c != len(rs) {
			continue
		}
		cmd := exec.Command("git", "branch", "-d", b)
		cmd.Dir = dir
		cmd.Stdout = lw
		cmd.Stderr = lw
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func mergeFF(dir string) error {
	lw := logwriter.NewColorWriter(filepath.Base(dir))
	defer lw.Flush()
	cmd := exec.Command("git", "merge", "--ff-only")
	cmd.Dir = dir
	cmd.Stdout = lw
	cmd.Stderr = lw
	fmt.Fprintf(lw, "merging %s\n", dir)

	return cmd.Run()
}

func submoduleUpdate(dir string) error {
	lw := logwriter.NewColorWriter(filepath.Base(dir))
	defer lw.Flush()
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Dir = dir
	cmd.Stdout = lw
	cmd.Stderr = lw

	return cmd.Run()
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

func runMerge(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	cfg, err := readConf()
	if err != nil {
		return err
	}

	gc := git.NewClient(nil)
	roots := strings.Split(viper.GetString("mainBranches"), ",")
	seen := make(map[string]bool)
	for _, dir := range cfg.RepoPaths() {
		wt, err := gc.WorkTree(dir)
		if err != nil {
			return err
		}
		if seen[wt.Repo] {
			continue
		}

		seen[wt.Repo] = true

		if viper.GetBool("prune-local") {
			if err := pruneLocal(dir, roots); err != nil {
				return err
			}
		}

		if len(wt.DirtyFiles) != 0 && !viper.GetBool("ignore-dirty") {
			continue
		}

		mc, err := wt.Commit.UpstreamMerge()
		if err != nil {
			continue
		}

		if !mc.CanFFMerge() {
			continue
		}

		if err := mergeFF(mc.Current.Repo); err != nil {
			return err
		}

		if err := submoduleUpdate(mc.Current.Repo); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	c := mergeCmd
	flags := c.Flags()

	RootCmd.AddCommand(c)
	flags.Bool("ignore-dirty", false, "Ignore dirty working directory when merging")
	flags.String("main-branches", "master,stable", "comma-separated list of branches to treat as roots for prune-local")
	flags.Bool("prune-local", false, "Remove local branches that have been merged with origin/{main-branches}")
}
