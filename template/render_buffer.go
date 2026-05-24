package template

import (
	"github.com/eugene-static/pdf-craft/buffer"
	"github.com/eugene-static/pdf-craft/core"
	"github.com/eugene-static/pdf-craft/meter"
)

type V3 struct {
	core         *core.Core
	buf          *buffer.Buffer
	page         core.Page
	block        *BlockV3
	headerHeight meter.MM
	x0           meter.MM
	y0           meter.MM
}

func NewV3(core *core.Core) *V3 {
	return &V3{
		core: core,
		block: &BlockV3{
			slotV3: &SlotV3{},
		},
	}
}

type BlockV3 struct {
	profile byte
	core    *core.Core
	slotV3  *SlotV3
	opts    Options
}

type SlotV3 struct {
	core *core.Core
	node node
}

func (t *V3) Block() *BlockV3 {
	return t.block
}

func (b *BlockV3) Slot() *SlotV3 {
	return b.slotV3
}

type node interface {
}
