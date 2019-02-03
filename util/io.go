package util

import (
	"io"
)

// defaultIO is a basic implementation of the IO interface.
type IOContainer struct {
	reader func() io.Reader
	output func() io.Writer
	closer func() io.Closer
	EncoderGroup
	DecoderGroup
}

func NewIOContainer(reader func() io.Reader, output func() io.Writer, closer func() io.Closer) *IOContainer {
	return &IOContainer{reader: reader, output: output, closer: closer, EncoderGroup: DefaultEncoders, DecoderGroup: DefaultDecoders}
}
