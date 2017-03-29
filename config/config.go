package config

type Config struct {
	Repos []*Repo `json:"repos"`
}

type Repo struct {
	Path string `json:"path"`
}

func (a Config) RepoPaths() (ret []string) {
	for _, c := range a.Repos {
		ret = append(ret, c.Path)
	}
	return
}
