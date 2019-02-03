package context

import (
	"context"
	"encoding/json"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
)

type ContextFunc func(context.Context, *Context)

type Context struct {
	In      io.Reader
	Out     io.Writer
	Object  interface{}
	Runners []ContextFunc
}

func (c *Context) Require(s string, def string) {
	if viper.Get(s) == nil {
		if v, ok := os.LookupEnv(s); ok == false || v == "" {
			viper.SetDefault(s, def)
			_ = os.Setenv(s, strings.ToUpper(def))
		}
	}
}


func NewContext(in io.Reader, out io.Writer, object interface{}, runners []ContextFunc) *Context {
	return &Context{In: in, Out: out, Object: object, Runners: runners}
}

func Execute(ctx context.Context, c *Context) error {
	ctx = context.WithValue(ctx, uuid.NewV4().String(), prettyJson(viper.AllSettings()))
	if ctx.Err() != nil {
		return ctx.Err()
	}
	for _, f := range c.Runners {
		f(ctx, c)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func prettyJson(v interface{}) []byte {
	output, _ := json.MarshalIndent(v, "", "  ")
	return output
}
