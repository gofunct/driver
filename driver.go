package driver

import (
	"bytes"
	"context"
	"errors"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/user"
	"strings"
)

type Driver interface {
	Run(ctx context.Context, obj interface{}) error
}

type ActFunc func(obj interface{}) (interface{}, error)
type HandlerFunc func(ctx context.Context, obj interface{}) (interface{}, error)
type WrapperFunc func(handlerFunc HandlerFunc) HandlerFunc

type Context struct {
	*bytes.Buffer
	Handlers  []WrapperFunc
	Finalizer func() error
}

func NewWrapperFunc(actions ...ActFunc) WrapperFunc {
	var err error
	return func(h HandlerFunc) HandlerFunc {
		return func(ctx context.Context, obj interface{}) (i interface{}, e error) {
			if ctx.Err() != nil {
				return obj, ctx.Err()
			}
			for _, a := range actions {
				obj, err = a(obj)
				if err != nil {
					return obj, err
				}
			}
			return obj, nil
		}
	}
}

func (c *Context) Run(ctx context.Context, obj interface{}) error {
	f := func() HandlerFunc {
		return func(_ context.Context, _ interface{}) (i interface{}, e error) {
			if ctx.Err() != nil {
				return obj, ctx.Err()
			}
			if obj == nil {
				return obj, errors.New("dynamic object must not be nil")
			}
			return obj, nil
		}
	}
	for _, h := range c.Handlers {
		f() = h(f())
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	switch x := obj.(type) {
	case error:
		return x
	case nil:
		return errors.New("error: nil object at end of run")
	}
	return c.Finalizer()
}

type Flagger struct {
	Name        string
	RequireRoot bool
	Annotations map[string]string
	Bind        func(fn func(viper.FlagValue))
	Store       io.Reader
	EnvPrefix   string
}

func (f *Flagger) VisitAll(fn func(viper.FlagValue)) {
	f.Bind(fn)
}

func (c *Context) Configure(flagger *Flagger) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	if usr.Name != "root" {
		return errors.New("root user is required to run")
	}
	viper.AutomaticEnv()
	if flagger.Store != nil {
		if err := viper.ReadConfig(flagger.Store); err != nil {
			return err
		}
	}
	if flagger.Annotations != nil {
		for k, v := range flagger.Annotations {
			_ = os.Setenv(k, v)
			viper.Set(k, v)
		}
	}
	if flagger.EnvPrefix != "" {
		viper.SetEnvPrefix(flagger.EnvPrefix)
	}

	if err := viper.BindFlagValues(flagger); err != nil {
		return err
	}
	return nil
}

func (c *Context) Require(s string, def string) {
	if viper.Get(s) == nil {
		if v, ok := os.LookupEnv(s); ok == false || v == "" {
			viper.SetDefault(s, def)
			_ = os.Setenv(s, strings.ToUpper(def))
		}
	}
}
