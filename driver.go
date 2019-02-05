package driver

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
)

type InitFunc func()
type HandlerFunc func() error

type Driver interface {
	Initializers() []InitFunc
	Handlers() []HandlerFunc
	Help() string
	Debug()
	Init(ctx context.Context, i ...InitFunc)
	Runnable(ctx context.Context) bool
	Run(ctx context.Context, h ...HandlerFunc) error
}

func Drive(ctx context.Context, d Driver) error {
	switch os.Args[1] {
	case "help", "--help", "-h", "h":
		fmt.Println(d.Help())
	case "debug", "--debug":
		d.Debug()
	}
	d.Init(ctx, d.Initializers()...)
	if !d.Runnable(ctx) {
		return errors.New("driver is currently not runnable")
	}
	if err := d.Run(ctx, d.Handlers()...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
