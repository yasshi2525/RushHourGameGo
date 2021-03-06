package entities

import (
	"log"
	"time"
)

// Base based on gorm.Model
type Base struct {
	ID      uint      `gorm:"primary_key" json:"id"`
	M       *Model    `gorm:"-"           json:"-"`
	T       ModelType `gorm:"-"           json:"-"`
	O       *Player   `gorm:"-"           json:"-"`
	OwnerID uint      `                   json:"oid"`
	// ChangedAt represents when model is changed. (UpdateAt is for gorm)
	ChangedAt time.Time `gorm:"-" json:"-"`
}

// NewBase create new Base
func (m *Model) NewBase(t ModelType, owner ...*Player) Base {
	var o *Player
	oid := ZERO
	if len(owner) > 0 && owner[0] != nil {
		o = owner[0]
		oid = o.ID
	}
	return Base{
		ID:        m.GenID(t),
		M:         m,
		T:         t,
		O:         o,
		OwnerID:   oid,
		ChangedAt: time.Now(),
	}
}

// Idx returns unique number of this model type.
func (b *Base) Idx() uint {
	return b.ID
}

// Type returns this entities' model type.
func (b *Base) Type() ModelType {
	return b.T
}

// Init set properties of it.
// Init must be invoked when it's created by reflection.
func (b *Base) Init(t ModelType, m *Model) {
	b.M = m
	b.T = t
	b.ChangedAt = time.Now()
}

// Permits always permits to Admin, Owner.
func (b *Base) Permits(target *Player) bool {
	switch target.Level {
	case Admin:
		return true
	case Normal:
		return b.O == target
	case Guest:
		return false
	default:
		log.Printf("invalid type %T: %+v", target.Level, target)
		return false
	}
}

// IsChanged returns true when it is changed after
func (b *Base) IsChanged(after time.Time) bool {
	return b.ChangedAt.Sub(after) > 0
}
