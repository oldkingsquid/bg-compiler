package processor

import (
	"bytes"
	"io"
	"strings"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/oldkingsquid/bg-compiler/flags"
)

type LogWriter struct {
	// input. Once output hits max bytes for either output then
	// the reader should be closed.
	input io.ReadCloser

	buf bytes.Buffer
}

func NewLogWriter(reader io.ReadCloser) *LogWriter {
	return &LogWriter{
		input: reader,
		buf:   *bytes.NewBuffer([]byte{}),
	}
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	bw, err := lw.buf.Write(p)
	if lw.buf.Len() >= flags.MaxReadOutputBytes() {
		lw.input.Close()
	}

	return bw, err
}

// Output returns the buffer value after writing limited
// by the max number of bytes returned.
func (lw LogWriter) Output() []byte {
	if lw.buf.Len() < flags.MaxReadOutputBytes() {
		return lw.buf.Bytes()
	}
	return lw.buf.Bytes()[0:flags.MaxReadOutputBytes()]
}

func (lw LogWriter) String() string {
	return string(lw.Output())
}

// ReadLogOutputs reads log outputs from the docker container and closes the
// reader once the max number of bytes has been read.
func ReadLogOutputs(r io.ReadCloser) (*LogWriter, *LogWriter, error) {
	stdOut, stdErr := NewLogWriter(r), NewLogWriter(r)
	_, err := stdcopy.StdCopy(stdOut, stdErr, r)
	if err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") ||
			strings.Contains(err.Error(), "read on closed response body") {
			return stdOut, stdErr, nil
		}
		return nil, nil, err
	}

	return stdOut, stdErr, nil
}
