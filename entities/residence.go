package entities

import (
	"fmt"
	"math/rand"
)

// Residence generate Human in a period
type Residence struct {
	Base
	Persistence
	Point

	Capacity int `json:"capacity"`
	// Wait represents how msec after it generates Human
	Wait float64 `json:"wait"`
	Name string  `json:"name"`

	Targets map[uint]*Human `gorm:"-" json:"-"`
	out     map[uint]*Step
}

// NewResidence create new instance without setting parameters
func (m *Model) NewResidence(x float64, y float64) *Residence {
	r := &Residence{
		Base:        m.NewBase(RESIDENCE),
		Persistence: NewPersistence(),
		Point:       NewPoint(x, y),
		Capacity:    m.conf.Residence.Capacity,
		Wait:        m.conf.Residence.Interval.D.Seconds() * rand.Float64(),
	}
	r.Init(m)
	r.Resolve()
	r.Marshal()
	m.Add(r)

	r.GenOutSteps()
	return r
}

// B returns base information of this elements.
func (r *Residence) B() *Base {
	return &r.Base
}

// P returns time information for database.
func (r *Residence) P() *Persistence {
	return &r.Persistence
}

// Pos returns entities' position.
func (r *Residence) Pos() *Point {
	return &r.Point
}

// GenOutSteps generates Steps from this Residence.
func (r *Residence) GenOutSteps() {
	// R -> C
	for _, c := range r.M.Companies {
		r.M.NewStep(r, c)
	}
	// R -> G
	for _, g := range r.M.Gates {
		r.M.NewStep(r, g)
	}
}

// Init creates map.
func (r *Residence) Init(m *Model) {
	r.Base.Init(RESIDENCE, m)
	r.out = make(map[uint]*Step)
	r.Targets = make(map[uint]*Human)
}

// OutSteps returns where it can go to
func (r *Residence) OutSteps() map[uint]*Step {
	return r.out
}

// InSteps returns where it comes from
func (r *Residence) InSteps() map[uint]*Step {
	return nil
}

// Resolve set reference
func (r *Residence) Resolve(args ...Entity) {
	for _, raw := range args {
		switch obj := raw.(type) {
		case *Human:
			r.Targets[obj.ID] = obj
		default:
			panic(fmt.Errorf("invalid type: %T %+v", obj, obj))
		}
	}
	r.Marshal()
}

// Marshal do nothing (for implements Resolvable)
func (r *Residence) Marshal() {
	// do-nothing
}

// UnMarshal set reference from id.
func (r *Residence) UnMarshal() {

}

// BeforeDelete remove reference of related entity
func (r *Residence) BeforeDelete() {
}

// CheckDelete check remaining reference
func (r *Residence) CheckDelete() error {
	return nil
}

// Delete removes this entity with related ones.
func (r *Residence) Delete() {
	for _, h := range r.Targets {
		r.M.Delete(h)
	}
	for _, s := range r.out {
		r.M.Delete(s)
	}
	r.M.Delete(r)
}

func (r *Residence) String() string {
	r.Marshal()
	return fmt.Sprintf("%s(%d):i=0,o=%d,h=%d:%v:%s", r.Type().Short(),
		r.ID, len(r.out), len(r.Targets), r.Pos(), r.Name)
}
