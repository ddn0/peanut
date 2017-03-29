package cmd

import (
	"io/ioutil"

	"github.com/ddn0/peanut/config"
	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
)

func readConf() (*config.Config, error) {
	dirFile := viper.GetString("dir")
	bs, err := ioutil.ReadFile(dirFile)
	if err != nil {
		return &config.Config{}, nil
	}

	var cfg config.Config
	if err := yaml.Unmarshal(bs, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func writeConf(cfg *config.Config) error {
	dirFile := viper.GetString("dir")
	bs, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dirFile, bs, 0666)
}
