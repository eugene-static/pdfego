package bytes

import (
	"bytes"
	"strconv"

	"github.com/eugene-static/pdfego/internal/components/stream"
)

const intSize = 32 << (^uint(0) >> 63)

type Writer struct {
	dst []byte
}

func NewWriter(dst []byte) *Writer {
	return &Writer{dst}
}

func (bw *Writer) Write(appender stream.Appender) *Writer {
	bw.dst = appender.Append(bw.dst)

	return bw
}

func (bw *Writer) WriteByte(b byte) *Writer {
	bw.dst = append(bw.dst, b)

	return bw
}

func (bw *Writer) WriteInt(value int) *Writer {
	bw.dst = AppendInt(bw.dst, int64(value))

	return bw
}

func (bw *Writer) WriteUint(value uint) *Writer {
	bw.dst = AppendUint(bw.dst, uint64(value))

	return bw
}

func (bw *Writer) WriteUintFixed(value uint, size int) *Writer {
	buf := make([]byte, size)

	for i := size - 1; i >= 0; i-- {
		buf[i] = byte('0' + value%10)
		value /= 10
	}

	bw.dst = append(bw.dst, buf...)

	return bw
}

func (bw *Writer) WriteFloat(value float64) *Writer {
	bw.dst = AppendFloat(bw.dst, value)

	return bw
}

func (bw *Writer) SP() *Writer {
	bw.dst = append(bw.dst, ' ')

	return bw
}

func (bw *Writer) LF() *Writer {
	bw.dst = append(bw.dst, '\n')

	return bw
}

func (bw *Writer) CRLF() *Writer {
	bw.dst = append(bw.dst, '\r', '\n')

	return bw
}

func (bw *Writer) Bytes() []byte {
	return bw.dst
}

func AppendInt(dst []byte, v int64) []byte {
	return strconv.AppendInt(dst, v, 10)
}

func AppendUint(dst []byte, v uint64) []byte {
	return strconv.AppendUint(dst, v, 10)
}

func AppendFloat(dst []byte, v float64) []byte {
	return strconv.AppendFloat(dst, v, 'f', 2, 64)
}

func AppendBool(dst []byte, b bool) []byte {
	if b {
		return append(dst, "true"...)
	}
	return append(dst, "false"...)
}

func AppendSpace(dst []byte) []byte {
	return append(dst, ' ')
}

func TrimSuffix(dst []byte, suffix []byte) []byte {
	return bytes.TrimSuffix(dst, suffix)
}

func TrimSpace(dst []byte) []byte {
	return bytes.TrimSpace(dst)
}

// Да, это калька со strconv. Но она нужна для парсинга слайса байт.
func ParseInt(s []byte) (v int, err error) {
	sLen := len(s)
	if intSize == 32 && (0 < sLen && sLen < 10) ||
		intSize == 64 && (0 < sLen && sLen < 19) {
		// Fast path for small integers that fit int type.
		s0 := s
		if s[0] == '-' || s[0] == '+' {
			s = s[1:]

			if len(s) < 1 {
				return 0, strconv.ErrSyntax
			}
		}

		n := 0

		for _, ch := range s {
			ch -= '0'

			if ch > 9 {
				return 0, strconv.ErrSyntax
			}

			n = n*10 + int(ch)
		}

		if s0[0] == '-' {
			n = -n
		}

		return n, nil
	}

	// Slow path for invalid, big, or underscored integers.
	i64, err := strconv.ParseInt(string(s), 10, 0)

	return int(i64), err
}

func ParseUint(s []byte) (v uint, err error) {
	sLen := len(s)
	if intSize == 32 && (0 < sLen && sLen < 10) ||
		intSize == 64 && (0 < sLen && sLen < 19) {

		n := 0

		for _, ch := range s {
			ch -= '0'

			if ch > 9 {
				return 0, strconv.ErrSyntax
			}

			n = n*10 + int(ch)
		}

		return uint(n), nil
	}

	u64, err := strconv.ParseUint(string(s), 10, 0)

	return uint(u64), err
}

func ParseFloat(s []byte) (v float64, err error) {
	return strconv.ParseFloat(string(s), 64)
}

func FormatInt(v int) string {
	return strconv.FormatInt(int64(v), 10)
}
