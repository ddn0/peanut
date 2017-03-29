package cmd

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ddn0/peanut/config"
	"github.com/spf13/viper"
)

func readConf() (*config.Config, error) {
	dirFile := viper.GetString("dir")
	bs, err := ioutil.ReadFile(dirFile)
	if err != nil {
		return &config.Config{}, nil
	}

	var cfg config.Config
	if err := json.Unmarshal(bs, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
