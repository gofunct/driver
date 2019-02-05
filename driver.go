package driver

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"reflect"
	"strings"
	"github.com/spf13/viper"
	"github.com/spf13/pflag"
)

var CommandLine = os.Args
var CalledAs = CommandLine[0]
var Args = CommandLine[1:]

var check = func(s string, list []string) error {
	for _, v := range list {
		if v == s {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("could not find path for---> %s in %s", s, list))
}
type Init func(ctx context.Context, d *Driver)
type Runner func(ctx context.Context, d *Driver) error
type Closer func(ctx context.Context, d *Driver)

type InitFunc func() Init
type RunnerFunc func() Runner
type CloserFunc func() Closer

type Handler struct {
	*pflag.FlagSet
	Initializers []InitFunc
	Runners []RunnerFunc
	Closers []CloserFunc
	*bytes.Buffer
}

func NewHandler(name string, data []byte, initializers []InitFunc, runners []RunnerFunc, closers []CloserFunc) *Handler {
	return &Handler{FlagSet: pflag.NewFlagSet(name, pflag.ExitOnError), Initializers: initializers, Runners: runners, Closers: closers, Buffer: bytes.NewBuffer(data)}
}
type Driver map[string]Handler

func(r Driver) Handle(key string, h Handler) {
	r[key] = h
}

func (r Driver) Drive(ctx context.Context, key string, d *Driver) error {
	var keys = []string{}
	for k, _ := range r {
		keys = append(keys, k)
	}
	if err := check(key, keys);err != nil {
		return err
	}
	for k, f := range r {
		if key == k {
			for _, i := range f.Initializers {
				newi := i()
				newi(ctx, d)
			}
			for _, ru := range f.Runners {
				newru := ru()
				if err := newru(ctx, d); err != nil {
					return err
				}
				for _, closer := range f.Closers {
					newcloser := closer()
					newcloser(ctx, d)
				}
			}
		}
	}
	return nil
}

func (d *Driver) RequireString(s string) string {
	for _, arg := range Args {
		switch {
		case strings.Contains(arg, "="):
			new := strings.Split(arg, "=")
			switch new[0] {
			case s:
				return new[1]
			case "--"+s:
				return new[1]
			case "-"+s:
				return new[1]

			}
		case strings.Contains(arg, ":"):
			new := strings.Split(arg, ":")
			switch new[0] {
			case s:
				return new[1]
			case "--"+s:
				return new[1]
			case "-"+s:
				return new[1]
			}
		}
	}
	if cfg := viper.GetString(s); cfg != "" {
		return cfg
	}
	if env, ok := os.LookupEnv(s); env != "" && ok {
		return env
	}
	if ans := d.Prompt(fmt.Sprintf("(REQUIRED STRING)----> %s", s)); ans != "" {
		return ans
	}
	panic(errors.New("Failed to set required variable"))

	return ""
}

func (d *Driver) Prompt(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return text
}


func (d *Driver) RequireBool(s string) bool {
	for _, arg := range Args {
		switch {
		case strings.Contains(arg, "="):
			new := strings.Split(arg, "=")
			switch new[0] {
			case s:
				switch new[1] {
				case "true", "t", "y", "yes", "Y", "Yes", "1":
					return true
				case "false", "f", "n", "no", "N", "No", "False":
					return false
				}
			case "--" + s:
				switch new[1] {
				case "true", "t", "y", "yes", "Y", "Yes", "1":
					return true
				case "false", "f", "n", "no", "N", "No", "False":
					return false
				}
			case "-" + s:
				switch new[1] {
				case "true", "t", "y", "yes", "Y", "Yes", "1":
					return true
				case "false", "f", "n", "no", "N", "No", "False":
					return false
				}
			}
		case strings.Contains(arg, ":"):
			new := strings.Split(arg, ":")
			switch new[0] {
			case s:
				switch new[1] {
				case "true", "t", "y", "yes", "Y", "Yes", "1":
					return true
				case "false", "f", "n", "no", "N", "No", "False":
					return false
				}
			case "--" + s:
				switch new[1] {
				case "true", "t", "y", "yes", "Y", "Yes", "1":
					return true
				case "false", "f", "n", "no", "N", "No", "False":
					return false
				}
			case "-" + s:
				switch new[1] {
				case "true", "t", "y", "yes", "Y", "Yes", "1":
					return true
				case "false", "f", "n", "no", "N", "No", "False":
					return false
				}
			}
		case reflect.DeepEqual(arg, s):
			return true
		}
	}
	if viper.IsSet(s) {
		return viper.GetBool(s)
	}
	if env, ok := os.LookupEnv(s); env != "" && ok {
		switch env {
		case "true", "t", "y", "yes", "Y", "Yes", "1":
			return true
		case "false", "f", "n", "no", "N", "No", "False":
			return false
		}
	}
	if ans := d.Prompt(fmt.Sprintf("(REQUIRED BOOL y/n t/f)----> %s", s)); ans != "" {
		switch ans {
		case "true", "t", "y", "yes", "Y", "Yes", "1":
			return true
		case "false", "f", "n", "no", "N", "No", "False":
			return false
		}
	}
	panic(errors.New("Failed to set required variable"))
	return false
}

func (d Driver) Debug() {
	r := []string{}
	for k, _ := range d {
		r = append(r, k)
	}
	viper.Set("Drivers", r)
	viper.Debug()
	for k, v := range r {
		fmt.Println(fmt.Sprintf("%s Function: %s", k, v))
	}
}