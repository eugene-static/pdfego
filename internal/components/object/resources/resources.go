package resources

import (
	"errors"
	"fmt"
	"sync"

	"github.com/eugene-static/pdfego/internal/components/dictionary"
	"github.com/eugene-static/pdfego/internal/components/object"
	"github.com/eugene-static/pdfego/internal/components/object/resources/font"
	"github.com/eugene-static/pdfego/internal/components/object/resources/image"
	"github.com/eugene-static/pdfego/internal/components/parameter"
	"github.com/eugene-static/pdfego/internal/errs"
)

const ObjectNumber parameter.Reference = 2

type Resources struct {
	mu     *sync.RWMutex
	fonts  map[string]*font.Font
	images map[string]*image.Image
	object map[string]*object.Object
}

func New() *Resources {
	return &Resources{
		mu:     &sync.RWMutex{},
		fonts:  make(map[string]*font.Font),
		images: make(map[string]*image.Image),
		object: make(map[string]*object.Object),
	}
}

func (r *Resources) SetFont(alias string, font *font.Font) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fonts[alias] = font
}

func (r *Resources) GetFont(alias string) (*font.Font, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.fonts) == 0 {
		err := errors.New("нет установленных шрифтов")

		return nil, err
	}

	fnt, ok := r.fonts[alias]
	if !ok {
		err := errs.ErrNotFound(fmt.Sprintf("шрифт с именем %s", alias))

		return nil, err
	}

	return fnt, nil
}

func (r *Resources) SetImage(alias string, img *image.Image) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.images[alias] = img
}

func (r *Resources) GetImage(alias string) (*image.Image, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	img, ok := r.images[alias]

	return img, ok
}

func (r *Resources) SetObject(alias string, obj *object.Object) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.object[alias] = obj
}

func (r *Resources) ForEachFont(fn func(k string, f *font.Font)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, v := range r.fonts {
		fn(k, v)
	}
}

func (r *Resources) ForEachImage(fn func(k string, img *image.Image)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, v := range r.images {
		fn(k, v)
	}
}

func (r *Resources) ForEachObject(fn func(k string, obj *object.Object)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, v := range r.object {
		fn(k, v)
	}
}

func (r *Resources) Resources(fonts, xObjects *dictionary.Dictionary) *object.Object {
	obj := object.NewSimple()

	_resources := &resources{
		Font:    fonts,
		XObject: xObjects,
	}

	obj.ReadDictionary(_resources)
	obj.SetNumber(ObjectNumber)

	return obj
}

type resources struct {
	Font    *dictionary.Dictionary `pdf:"Font"`
	XObject *dictionary.Dictionary `pdf:"XObject,omitempty"`
}
