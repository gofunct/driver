package driver

import (
	"context"
	"github.com/pkg/errors"
)

type Driver interface {
	Runnable() bool
	Run(ctx context.Context) error
}

func Drive(ctx context.Context, d Driver) error {
	if !d.Runnable() {
		return errors.New("driver is currently not runnable")
	}
	if err := d.Run(ctx); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
