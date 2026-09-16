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
	fonts  map[string]*font.Fonts
	images map[string]*image.Image
	object map[string]*object.Object
}

func New() *Resources {
	return &Resources{
		mu:     &sync.RWMutex{},
		fonts:  make(map[string]*font.Fonts),
		images: make(map[string]*image.Image),
		object: make(map[string]*object.Object),
	}
}

func (r *Resources) SetFont(alias string, _font *font.Font) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.fonts[alias]
	if !ok {
		r.fonts[alias] = new(font.Fonts)
	}

	r.fonts[alias].Add(_font)
}

func (r *Resources) GetFonts(alias string) (*font.Fonts, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.fonts) == 0 {
		err := errors.New("нет установленных шрифтов")

		return nil, err
	}

	fonts, ok := r.fonts[alias]
	if !ok {
		err := errs.ErrNotFound(fmt.Sprintf("шрифтов с именем %s", alias))

		return nil, err
	}

	return fonts, nil
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

func (r *Resources) ForEachFont(fn func(f *font.Font)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, fonts := range r.fonts {
		fonts.ForEach(fn)
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
