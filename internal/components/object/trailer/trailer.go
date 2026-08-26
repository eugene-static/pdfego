package trailer

import (
	"bytes"

	"github.com/eugene-static/pdfego/internal/components/dictionary"
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
	"github.com/eugene-static/pdfego/internal/errs"
)

type Trailer struct {
	Info parameter.Reference `pdf:"Info"`
	Root parameter.Reference `pdf:"Root"`
	Size parameter.Integer   `pdf:"Size"`
}

func New() *Trailer {
	return &Trailer{}
}

func ParseTrailer(body []byte) (*object.Object, error) {
	trailerIndex := bytes.LastIndex(body, []byte(primitives.Trailer))
	if trailerIndex == -1 {
		err := errs.ErrNotFound(primitives.Trailer.String())

		return nil, err
	}

	trailerBody := body[trailerIndex:]

	startXrefIndex := bytes.Index(trailerBody, []byte(primitives.StartXRef))
	if startXrefIndex == -1 {
		err := errs.ErrNotFound(primitives.StartXRef.String())

		return nil, err
	}

	dict, err := dictionary.Parse(trailerBody[:startXrefIndex])
	if err != nil {
		return nil, err
	}

	obj := object.NewSimple()

	obj.SetDictionary(dict)

	return obj, nil
}

func (tr *Trailer) SetInfo(info parameter.Reference) {
	tr.Info = info
}

func (tr *Trailer) SetRoot(root parameter.Reference) {
	tr.Root = root
}

func (tr *Trailer) SetSize(size parameter.Integer) {
	tr.Size = size
}

func (tr *Trailer) WriteToStream(str *stream.Stream, xrefIndex parameter.Integer) {
	dict := dictionary.New()

	dict.Encode(tr)

	str.NewStreamWriter().
		Write(primitives.Trailer).LF().
		Write(dict).LF().
		Write(primitives.StartXRef).LF().
		Write(xrefIndex).LF().
		Write(primitives.EOF).
		Close()
}
