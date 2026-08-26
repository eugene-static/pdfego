package compressor

import (
	"bytes"
	"compress/zlib"
)

type Compressor struct {
	enabled bool
	buffer  *bytes.Buffer
	writer  *zlib.Writer
}

func New() *Compressor {
	buf := bytes.NewBuffer(make([]byte, 0, 1<<16))

	return &Compressor{
		buffer: buf,
		writer: zlib.NewWriter(buf),
	}
}

func (comp *Compressor) Enabled() bool {
	return comp.enabled
}

func (comp *Compressor) Enable(enable bool) {
	comp.enabled = enable
}

func (comp *Compressor) Compress(buf []byte) ([]byte, error) {
	if !comp.enabled {
		return buf, nil
	}

	return comp.compress(buf)
}

func (comp *Compressor) ForceCompress(buf []byte) ([]byte, error) {
	return comp.compress(buf)
}

func (comp *Compressor) compress(buf []byte) ([]byte, error) {
	comp.buffer.Reset()
	comp.writer.Reset(comp.buffer)

	_, err := comp.writer.Write(buf)
	if err != nil {
		return nil, err
	}

	err = comp.writer.Close()
	if err != nil {
		return nil, err
	}

	compressedLen := comp.buffer.Len()

	if compressedLen <= len(buf) {
		copy(buf[:compressedLen], comp.buffer.Bytes())

		return buf[:compressedLen], nil
	}

	buf = make([]byte, comp.buffer.Len())
	copy(buf, comp.buffer.Bytes())

	return buf, nil
}
