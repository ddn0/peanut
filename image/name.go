package image

import (
	"path"
	"strings"

	docker "github.com/ddn0/go-dockerclient"
)

// Construct an image name from a host registry, repo dir, repo name and tag
func ImageName(host, repo, tag string) string {
	var parts []string
	if len(host) != 0 {
		parts = append(parts, host)
		parts = append(parts, "/")
	}

	parts = append(parts, repo)
	if len(tag) != 0 {
		parts = append(parts, ":")
		parts = append(parts, tag)
	}

	return strings.Join(parts, "")
}

// Break up a docker image name like localhost.domain:1111/repo/name:tag into
// (host, repo/name, tag)
func ParseImageName(image string) (host string, repo string, tag string) {
	repo, tag = docker.ParseRepositoryTag(image)
	i := strings.Index(repo, "/")
	i2 := strings.Index(repo[i+1:], "/")
	if i2 >= 0 {
		host = repo[0:i]
		repo = repo[i+1:]
	}
	return
}

// Break up docker repo name into (dir, name)
func ParseRepo(repo string) (dir string, name string) {
	dir = path.Dir(repo)
	name = path.Base(repo)
	return
}

// Return basename of image name
func Basename(image string) (n string) {
	_, repo, _ := ParseImageName(image)
	_, n = ParseRepo(repo)
	return
}
