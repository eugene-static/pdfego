package object

import (
	"github.com/eugene-static/pdfego/internal/components/dictionary"
	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/components/stream"
	"github.com/eugene-static/pdfego/internal/components/stream/primitives"
)

type Object struct {
	number     parameter.Reference
	generation parameter.Integer
	dictionary *dictionary.Dictionary
	stream     *stream.Stream
}

func New() *Object {
	return &Object{
		dictionary: dictionary.New(),
		stream:     stream.New(),
	}
}

func NewSimple() *Object {
	return &Object{
		dictionary: dictionary.New(),
	}
}

func Read(dict any, str []byte) *Object {
	obj := &Object{
		dictionary: dictionary.New(),
	}

	obj.ReadDictionary(dict)

	if str != nil {
		obj.stream = stream.New()

		obj.ReadStream(str)
	}

	return obj
}

func Setup(parameters []*parameter.Parameter, stream *stream.Stream) *Object {
	return &Object{
		dictionary: dictionary.New(parameters...),
		stream:     stream,
	}
}

func (obj *Object) WriteToStream(dst *stream.Stream) {
	builder := dst.NewStreamWriter().
		Write(parameter.Integer(obj.number)).SP().
		Write(obj.generation).SP().
		Write(primitives.Object).LF().
		Write(primitives.DictionaryOpen).LF()

	for _, param := range *obj.dictionary {
		builder.Write(param).LF()
	}

	builder.
		Write(primitives.DictionaryClose).LF()

	if obj.stream != nil && obj.stream.Len() > 0 {
		builder.
			Write(primitives.Stream).LF().
			Write(obj.Stream()).LF().
			Write(primitives.EndStream).LF()
	}

	builder.
		Write(primitives.EndObject).LF().
		LF().Close()
}

func (obj *Object) Number() parameter.Reference {
	return obj.number
}

func (obj *Object) Reference() *parameter.Reference {
	return &obj.number
}

func (obj *Object) Dictionary() *dictionary.Dictionary {
	return obj.dictionary
}

func (obj *Object) Stream() *stream.Stream {
	return obj.stream
}

func (obj *Object) Reset() {
	obj.number = 0

	if obj.stream != nil {
		obj.stream.Reset()
	}

	if obj.dictionary != nil {
		clear(*obj.dictionary)
	}
}

func (obj *Object) SetNumber(number parameter.Reference) {
	obj.number = number
}

func (obj *Object) SetDictionary(dict *dictionary.Dictionary) {
	obj.dictionary = dict
}

func (obj *Object) ReadDictionary(dict any) {
	obj.dictionary.Encode(dict)
}

func (obj *Object) SetStream(stream *stream.Stream) {
	obj.stream = stream
}

func (obj *Object) ReadStream(bytes []byte) {
	obj.stream.Write(bytes)
}
