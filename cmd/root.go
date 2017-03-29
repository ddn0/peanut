package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	stderr  = os.Stderr
	stdout  = os.Stdout
)

var RootCmd = &cobra.Command{
	Use:           "peanut",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func configDir() string {
	if d, p := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); len(d) > 0 && len(p) > 0 {
		return filepath.Join(d, p, ".peanut")
	}
	return filepath.Join(os.Getenv("HOME"), ".peanut")
}

func init() {
	cobra.OnInitialize(initConfig)
	c := RootCmd
	flags := c.PersistentFlags()

	flags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.peanut/config.yaml)")
	flags.Bool("verbose", false, "Print more output")
	flags.String("pdir", filepath.Join(configDir(), "dir"), "Path to package directory file")
	flags.String("dockerConf", filepath.Join(os.Getenv("HOME"), ".docker", "config.json"), "Path to docker auth file")
	flags.Int("maxConcurrent", 8, "Maximum number of concurrent operations to attempt")
	flags.Duration("timeout", 5*time.Minute, "Timeout")
}

func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(configDir())
	viper.SetEnvPrefix("peanut") // prefix of environment variables
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()
}
