package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/washtubs/expire"
)

func main() {
	globals := flag.NewFlagSet("", flag.ExitOnError)
	globals.Bool("dummy", false, "DUMMY FLAG")
	globals.Parse(os.Args[1:])

	commandStr := globals.Arg(0)
	if commandStr == "" {
		log.Fatal("No command: TODO list commands")
	}
	cmdArgs := globals.Args()[1:]
	cmd := getCommand(commandStr)

	cmdFs := cmd.flags()
	cmdFs.Parse(cmdArgs)

	err := cmd.parse(cmdFs)
	if err != nil {
		panic(err.Error())
	}

	err = cmd.exec()
	if err != nil {
		switch err.(type) {
		case exitCodeError:
			if err.Error() != "" {
				fmt.Errorf("%s\n", err.Error())
			}
			os.Exit(err.(exitCodeError).code)
		default:
			panic(err.Error())
		}
	}

}

type arrayFlags struct {
	values *[]string
}

func (i *arrayFlags) String() string {
	// TODO
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i.values = append(*i.values, value)
	return nil
}

func AddDryRunFlags(fs *flag.FlagSet, config *expire.DryRunConfig) {
	fs.BoolVar(&config.IsDryRun, "n", false, "TODO")
}

func AddBatchRunFlags(fs *flag.FlagSet, config *expire.BatchRunConfig) {
	fs.BoolVar(&config.IsBatchRun, "b", false, "TODO")
}

func ParseTargets(fs *flag.FlagSet, config *expire.TargetConfig) {
	config.Targets = fs.Args()
	if len(fs.Args()) > 0 {
		config.Target = fs.Arg(0)
	}
}

type exitCodeError struct {
	code int
	err  error
}

func (e exitCodeError) Error() string {
	if e.err != nil {
		return e.err.Error()
	} else {
		return ""
	}
}

type Command struct {
	// Create the initial flagset
	flags func() *flag.FlagSet

	// Called after parsing to finish fully forming the Config object
	parse func(*flag.FlagSet) error

	// Executes the command
	exec func() error
}

func getInitCommand() Command {
	var (
		config *expire.InitConfig
	)
	config = &expire.InitConfig{}

	flags := func() *flag.FlagSet {
		fs := flag.NewFlagSet("init", flag.ExitOnError)
		AddDryRunFlags(fs, &config.DryRunConfig)
		return fs
	}
	parse := func(fs *flag.FlagSet) error {
		return nil
	}
	exec := func() error {
		return expire.Init(config)
	}
	return Command{
		flags,
		parse,
		exec,
	}
}

func getNewCommand() Command {
	var (
		config   *expire.NewConfig
		duration string
	)
	config = &expire.NewConfig{}

	flags := func() *flag.FlagSet {
		fs := flag.NewFlagSet("new", flag.ExitOnError)
		fs.StringVar(&duration, "duration", "", "TODO")
		fs.BoolVar(&config.ResetOnTouch, "reset-on-touch", false, "TODO")
		fs.BoolVar(&config.Init, "init", false, "TODO")
		fs.BoolVar(&config.NoShadow, "no-shadow", false, "TODO")
		AddDryRunFlags(fs, &config.DryRunConfig)
		AddBatchRunFlags(fs, &config.BatchRunConfig)
		return fs
	}
	parse := func(fs *flag.FlagSet) error {
		if duration != "" {
			var err error
			config.Duration, err = expire.ParseDurationString(duration)
			if err != nil {
				return err
			}
		}
		ParseTargets(fs, &config.TargetConfig)
		return nil
	}
	exec := func() error {
		return expire.New(config)
	}
	return Command{
		flags,
		parse,
		exec,
	}
}

func getTouchCommand() Command {
	var (
		config *expire.TouchConfig
	)
	config = &expire.TouchConfig{}

	flags := func() *flag.FlagSet {
		fs := flag.NewFlagSet("touch", flag.ExitOnError)
		AddDryRunFlags(fs, &config.DryRunConfig)
		AddBatchRunFlags(fs, &config.BatchRunConfig)
		return fs
	}
	parse := func(fs *flag.FlagSet) error {
		ParseTargets(fs, &config.TargetConfig)
		return nil
	}
	exec := func() error {
		return expire.Touch(config)
	}
	return Command{
		flags,
		parse,
		exec,
	}
}

func getRenewCommand() Command {
	var (
		config *expire.RenewConfig
	)
	config = &expire.RenewConfig{}

	flags := func() *flag.FlagSet {
		fs := flag.NewFlagSet("renew", flag.ExitOnError)
		AddDryRunFlags(fs, &config.DryRunConfig)
		AddBatchRunFlags(fs, &config.BatchRunConfig)
		return fs
	}
	parse := func(fs *flag.FlagSet) error {
		ParseTargets(fs, &config.TargetConfig)
		return nil
	}
	exec := func() error {
		return expire.Renew(config)
	}
	return Command{
		flags,
		parse,
		exec,
	}
}

func getCheckCommand() Command {
	var (
		config *expire.CheckConfig
	)
	config = &expire.CheckConfig{}

	flags := func() *flag.FlagSet {
		fs := flag.NewFlagSet("check", flag.ExitOnError)
		return fs
	}
	parse := func(fs *flag.FlagSet) error {
		ParseTargets(fs, &config.TargetConfig)
		return nil
	}
	exec := func() error {
		checkResp, err := expire.Check(config)
		return exitCodeError{
			code: int(checkResp),
			err:  err,
		}
	}
	return Command{
		flags,
		parse,
		exec,
	}
}

func getDeleteCommand() Command {
	var (
		config *expire.DeleteConfig
	)
	config = &expire.DeleteConfig{}

	flags := func() *flag.FlagSet {
		fs := flag.NewFlagSet("delete", flag.ExitOnError)
		fs.BoolVar(&config.DeInit, "de-init", false, "TODO")
		AddDryRunFlags(fs, &config.DryRunConfig)
		AddBatchRunFlags(fs, &config.BatchRunConfig)
		return fs
	}
	parse := func(fs *flag.FlagSet) error {
		ParseTargets(fs, &config.TargetConfig)
		return nil
	}
	exec := func() error {
		return expire.Delete(config)
	}
	return Command{
		flags,
		parse,
		exec,
	}
}

func getNextCommand() Command {
	var (
		format string
		config *expire.NextConfig
	)
	config = &expire.NextConfig{}

	flags := func() *flag.FlagSet {
		fs := flag.NewFlagSet("next", flag.ExitOnError)
		fs.BoolVar(&config.Delete, "delete", false, "TODO")
		fs.BoolVar(&config.Expired, "expired", false, "TODO")
		fs.BoolVar(&config.Exist, "exist", false, "TODO")
		fs.BoolVar(&config.NoExist, "no-exist", false, "TODO")
		fs.Var(&arrayFlags{&config.MatchGlob}, "match-glob", "TODO")
		fs.Var(&arrayFlags{&config.MatchRegex}, "match-regex", "TODO")
		fs.IntVar(&config.Limit, "limit", 0, "TODO")
		fs.StringVar(&format, "format", "", "TODO")
		return fs
	}
	parse := func(fs *flag.FlagSet) error {
		return nil
	}
	exec := func() error {
		if format == "" {
			format = "{{ .TargetContextual }} - {{ .ExpirationRelative }}"
		}
		format = format + "\n"
		t, err := template.New("next").Parse(format)
		if err != nil {
			return err
		}

		recs, err := expire.Next(config)
		if err != nil {
			return err
		}

		for _, rec := range recs {
			err := t.Execute(os.Stdout, rec)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return Command{
		flags,
		parse,
		exec,
	}
}

func getCommand(cmd string) Command {
	switch cmd {
	case "init":
		return getInitCommand()
	case "new":
		return getNewCommand()
	case "touch":
		return getTouchCommand()
	case "renew":
		return getRenewCommand()
	case "check":
		return getCheckCommand()
	case "delete":
		return getDeleteCommand()
	case "next":
		return getNextCommand()
	}
	panic("Unhandled command: " + cmd)
}
