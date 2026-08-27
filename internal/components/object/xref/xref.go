package xref

import (
	"bufio"
	"bytes"
	"strconv"

	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	bw "github.com/eugene-static/pdfego/internal/components/stream/bytes"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/internal/errs"
)

const (
	maxGeneration   uint = 65535
	minRecordsCount      = 3
)

type XRef struct {
	index               uint64
	initialObjectNumber parameter.Integer
	recordsCount        parameter.Integer
	records             []Record
}

func New() *XRef {
	bytesOffsets := make([]Record, minRecordsCount, 100)
	bytesOffsets[0] = nullXRefRecord()

	return &XRef{
		records:      bytesOffsets,
		recordsCount: minRecordsCount,
	}
}

func ParseXref(body []byte, trailerIndex int) (*XRef, error) {
	startXrefIndex := bytes.LastIndex(body, []byte(primitives.StartXRef))
	if startXrefIndex == -1 {
		err := errs.ErrNotFound(primitives.StartXRef.String())

		return nil, err
	}

	eofIndex := bytes.LastIndex(body, []byte(primitives.EOF.String()))
	if eofIndex == -1 {
		err := errs.ErrNotFound(primitives.EOF.String())

		return nil, err
	}

	xrefIndex, err := strconv.ParseUint(string(bytes.TrimSpace(body[startXrefIndex:eofIndex])), 10, 64)
	if err != nil {
		return nil, err
	}

	xrefReader := bytes.NewReader(body[xrefIndex:trailerIndex])
	xrefScanner := bufio.NewScanner(xrefReader)

	lineIndex := 0

	var (
		initialObjectNumber uint
		recordsCount        int
		records             []Record
	)

	for xrefScanner.Scan() {
		line := xrefScanner.Bytes()

		if lineIndex == 0 || len(line) == 0 {
			lineIndex++

			continue
		}

		switch lineIndex {
		case 1:
			subSection := bytes.Fields(line)
			if len(subSection) != 2 {
				err = errs.ErrInvalidFormat("XRef subsection")

				return nil, err
			}

			initialObjectNumber, err = bw.ParseUint(subSection[0])
			if err != nil {
				return nil, err
			}

			recordsCount, err = bw.ParseInt(subSection[1])
			if err != nil {
				return nil, err
			}

			records = make([]Record, 0, recordsCount)
		default:
			record := bytes.Fields(line)
			if len(record) != 3 {
				err = errs.ErrInvalidFormat("XRef record")

				return nil, err
			}

			offset, err := bw.ParseUint(record[0])
			if err != nil {
				return nil, err
			}

			generation, err := bw.ParseUint(record[1])
			if err != nil {
				return nil, err
			}

			inUseEntry := record[2][0]

			records = append(records, Record{
				BytesOffset: offset,
				Generation:  generation,
				InUseEntry:  inUseEntry,
			})

		}

		lineIndex++
	}

	return &XRef{
		index:               xrefIndex,
		initialObjectNumber: parameter.Integer(initialObjectNumber),
		recordsCount:        parameter.Integer(recordsCount),
		records:             records,
	}, nil
}

func (xref *XRef) Index() parameter.Integer {
	return parameter.Integer(xref.index)
}

func (xref *XRef) RecordsCount() parameter.Integer {
	return xref.recordsCount
}

func (xref *XRef) UpdateBytesOffset(offset uint64) {
	xref.index = offset
}

func (xref *XRef) ObjectBytesOffset(objectNumber parameter.Reference) int {
	if int(objectNumber) > len(xref.records) {
		return -1
	}

	return int(xref.records[objectNumber].BytesOffset)
}

func (xref *XRef) NewRecord(byteOffset uint) parameter.Reference {
	objNum := len(xref.records)

	xref.records = append(xref.records, Record{
		BytesOffset: byteOffset,
		Generation:  0,
		InUseEntry:  primitives.RecordIsUsed,
	})

	xref.recordsCount++

	return parameter.Reference(objNum)
}

func (xref *XRef) UpdateRecord(index parameter.Reference, byteOffset uint) error {
	if int(index) > int(xref.recordsCount) || index < 0 {
		err := errs.ErrNotFound("XRef record")

		return err
	}

	xref.records[index].BytesOffset = byteOffset
	xref.records[index].InUseEntry = primitives.RecordIsUsed

	return nil
}

func (xref *XRef) WriteToStream(dst *stream.Stream) {
	sw := dst.NewStreamWriter().
		Write(primitives.XRef).LF().
		Write(xref.initialObjectNumber).SP().
		Write(xref.recordsCount).LF()

	for i := range xref.records {
		sw.Write(xref.records[i])
	}

	sw.Close()
}

func (xref *XRef) Reset() {
	xref.index = 0
	xref.initialObjectNumber = 0
	xref.recordsCount = minRecordsCount
	xref.records = xref.records[:minRecordsCount]
}

type Record struct {
	BytesOffset uint
	Generation  uint
	InUseEntry  byte
}

func nullXRefRecord() Record {
	return Record{
		BytesOffset: 0,
		Generation:  maxGeneration,
		InUseEntry:  primitives.RecordNotUsed,
	}
}

func (r Record) Append(dst []byte) []byte {
	dst = bw.NewWriter(dst).
		WriteUintFixed(r.BytesOffset, 10).SP().
		WriteUintFixed(r.Generation, 5).SP().
		WriteByte(r.InUseEntry).CRLF().
		Bytes()

	return dst
}
