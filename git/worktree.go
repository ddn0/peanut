package git

import (
	"bytes"
	"strings"
)

// A WorkTree represents the working directory in git.
type WorkTree struct {
	Commit     *Commit
	DirtyFiles []string
	Repo       string
	client     *Client
}

// WorkTree returns the WorkTree for the given directory.
func (a *Client) WorkTree(dir string) (*WorkTree, error) {
	repo, err := output(dir, a.gitPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	repoStr := string(bytes.TrimSpace(repo))

	dirty, err := output(repoStr, a.gitPath, "ls-files",
		"--exclude-standard",
		"--others",
		"--deleted",
		"--modified",
		"--directory",
		"--no-empty-directory",
		"-z")
	if err != nil {
		return nil, err
	}
	dirtyFiles := bytes.Split(dirty, []byte{'\x00'})

	commit, err := a.Head(repoStr)
	if err != nil {
		return nil, err
	}

	wt := &WorkTree{
		Commit: commit,
		Repo:   repoStr,
		client: a,
	}
	for _, name := range dirtyFiles {
		if len(name) == 0 {
			continue
		}
		wt.DirtyFiles = append(wt.DirtyFiles, string(name))
	}
	return wt, nil
}

// UnmergedBranches returns branches that are not merged in branch.
func (a *WorkTree) UnmergedBranches(branch string) ([]string, error) {
	branches, err := output(a.Repo, a.client.gitPath, "branch", "--no-merged", branch)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, bs := range bytes.Split(branches, []byte{'\n'}) {
		s := string(bytes.TrimSpace(bs))
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "*") {
			continue
		}
		ret = append(ret, s)
	}
	return ret, nil
}
