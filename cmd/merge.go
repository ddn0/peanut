package cmd

import (
	"bytes"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

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
		if _, err := readGitCommand(dir, "rev-parse", "--verify", r); err != nil {
			continue
		}
		rs = append(rs, r)
	}

	if len(rs) == 0 {
		return nil
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
		if err := execGitCommand(dir, "branch", "-d", b); err != nil {
			return err
		}
	}
	return nil
}

func isAncestor(dir, ref1, ref2 string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", ref1, ref2)
	cmd.Dir = dir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		return false, err
	}
	s, ok := ee.Sys().(syscall.WaitStatus)
	if !ok {
		return false, err
	}
	if s.ExitStatus() == 1 {
		return false, nil
	}
	return false, err
}

// Return dir to root if curBranch has been merged in remote root
func returnMerged(dir string, curBranch, root string) error {
	if curBranch == root {
		return nil
	}

	upstream, err := readGitCommand(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", root+"@{upstream}")
	if err != nil {
		return err
	}

	merged, err := isAncestor(dir, curBranch, upstream)
	if err != nil {
		return err
	}
	if !merged {
		return nil
	}

	if err := execGitCommand(dir, "checkout", root); err != nil {
		return err
	}

	lw := logwriter.NewColorWriter(filepath.Base(dir))
	defer lw.Flush()

	lw.Printf("cd %s && git checkout %s\n", dir, root)
	return nil
}

func execGitCommand(dir string, args ...string) error {
	lw := logwriter.NewColorWriter(filepath.Base(dir))
	defer lw.Flush()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = lw
	cmd.Stderr = lw

	return cmd.Run()
}

func readGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	bs, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(bs)), nil
}

func execBranchCommand(out io.Writer, dir string, args ...string) ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", append([]string{"branch"}, args...)...)
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

func mergedBranches(out io.Writer, dir, root string) ([]string, error) {
	return execBranchCommand(out, dir, "--merged", root)
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

	returnRoot := viper.GetString("branch-for-return")
	roots := strings.Split(viper.GetString("branches-for-prune-local"), ",")
	for _, dir := range cfg.RepoPaths() {
		wt, err := gc.WorkTree(dir)
		if err != nil {
			return err
		}

		if len(wt.DirtyFiles) != 0 && !viper.GetBool("ignore-dirty") {
			continue
		}

		if viper.GetBool("return") {
			if err := returnMerged(dir, wt.Commit.Branch, returnRoot); err != nil {
				return err
			}
			wt, err = gc.WorkTree(dir)
			if err != nil {
				return err
			}
		}

		if viper.GetBool("prune-local") {
			if err := pruneLocal(dir, roots); err != nil {
				return err
			}
		}

		mc, err := wt.Commit.UpstreamMerge()
		if err != nil {
			continue
		}

		if !mc.CanFFMerge() {
			continue
		}

		if err := execGitCommand(mc.Current.Repo, "merge", "--ff-only"); err != nil {
			return err
		}

		if err := execGitCommand(mc.Current.Repo, "submodule", "update", "--init", "--recursive"); err != nil {
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
	flags.Bool("return", false, "If current branch has been merged in origin/{branch-for-return}, checkout {branch-for-return}")
	flags.String("branch-for-return", "master", "Branch to treat as root for return")
	flags.Bool("prune-local", false, "Remove local branches that have been merged in remote {branches-for-prune-local}")
	flags.String("branches-for-prune-local", "master,stable", "Comma-separated list of branches to treat as roots for prune-local")
}
