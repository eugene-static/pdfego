package stream

import (
	"bytes"
	"io"
)

type Appender interface {
	Append(dst []byte) []byte
}

type Stream struct {
	stream *bytes.Buffer
}

func New() *Stream {
	return &Stream{
		stream: bytes.NewBuffer(make([]byte, 0, 1<<16)),
	}
}

func (str *Stream) Write(bytes []byte) (n int, err error) {
	str.stream.Grow(len(bytes))
	return str.stream.Write(bytes)
}

func (str *Stream) WriteTo(dst io.Writer) (int64, error) {
	n, err := dst.Write(str.stream.Bytes())

	return int64(n), err
}

func (str *Stream) WriteString(s string) {
	str.stream.WriteString(s)
}

func (str *Stream) AvailableBuffer() []byte {
	return str.stream.AvailableBuffer()
}

func (str *Stream) Len() int {
	return str.stream.Len()
}

func (str *Stream) Bytes() []byte {
	return str.stream.Bytes()
}

func (str *Stream) Reset() {
	str.stream.Reset()
}

func (str *Stream) Append(dst []byte) []byte {
	dst = append(dst, str.stream.Bytes()...)

	return dst
}

type Writer struct {
	stream *Stream
	buffer []byte
}

func (str *Stream) NewStreamWriter() *Writer {
	return &Writer{
		stream: str,
		buffer: str.AvailableBuffer(),
	}
}

func (sw *Writer) Write(appender Appender) *Writer {
	sw.buffer = appender.Append(sw.buffer)

	return sw
}

func (sw *Writer) SP() *Writer {
	sw.buffer = append(sw.buffer, ' ')

	return sw
}

func (sw *Writer) LF() *Writer {
	sw.buffer = append(sw.buffer, '\n')

	return sw
}

func (sw *Writer) Close() {
	sw.stream.Write(sw.buffer)
}
