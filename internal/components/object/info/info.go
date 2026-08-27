package info

import (
	"time"

	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/parameter"
)

// TODO: перейти на xml

const DateLayout = "D:20060102150405-07'00'"

type Info struct {
	Producer     parameter.Text `pdf:"Producer"`
	CreationDate parameter.Text `pdf:"CreationDate"`
}

func New() *Info {
	return &Info{
		CreationDate: parameter.Text(time.Now().Format(DateLayout)),
		Producer:     parameter.Text("pdfego"),
	}
}

func (info *Info) SetProducer(producer string) {
	info.Producer = parameter.Text(producer)
}

func (info *Info) SetCreationDate(creationDate time.Time) {
	info.CreationDate = parameter.Text(creationDate.Format(DateLayout))
}

func (info *Info) Info() *object.Object {
	obj := object.NewSimple()

	obj.ReadDictionary(info)

	return obj
}
