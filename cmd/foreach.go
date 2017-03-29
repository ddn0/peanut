package cmd

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/ddn0/peanut/logwriter"
	"github.com/ddn0/peanut/pdo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var foreachCmd = &cobra.Command{
	Use:   "foreach [args] [--] <command>",
	Short: "execute a command in each directory",
	RunE:  runForeach,
}

func spawn(ctx context.Context, item interface{}, args []string) error {
	dir := item.(string)

	lw := logwriter.NewColorWriter(filepath.Base(dir))
	defer lw.Flush()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = lw
	cmd.Stderr = lw

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

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

func runForeach(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	if len(args) == 0 {
		return nil
	}

	cfg, err := readConf()
	if err != nil {
		return err
	}

	var dirs []interface{}
	for _, dir := range cfg.RepoPaths() {
		dirs = append(dirs, dir)
	}

	if err := pdo.DoAll(pdo.DoAllOpt{
		Func: func(ctx context.Context, item interface{}) error {
			return spawn(ctx, item, args)
		},
		Items:         dirs,
		Timeout:       viper.GetDuration("timeout"),
		MaxConcurrent: viper.GetInt("max-concurrent"),
	}); err != nil {
		return err
	}
	return nil

	return writeConf(cfg)
}

func init() {
	c := foreachCmd
	//flags := c.Flags()

	RootCmd.AddCommand(c)
}
