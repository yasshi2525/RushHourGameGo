package entities

import (
	"fmt"
	"math"
	"time"
)

// Standing is for judgement Human placement on same X, Y
type Standing uint

const (
	// OnGround represents Human still not arrive at Station or get off Train forcefully
	OnGround Standing = iota + 1
	// OnPlatform represents Human enter Station and wait for Train
	OnPlatform
	// OnTrain represents Human ride on Train
	OnTrain
)

// Human commute from Residence to Company by Train
type Human struct {
	Base
	Point
	Owner

	// Avaialble represents how many seconds Human is able to use for moving or staying.
	Available float64 `gorm:"not null" json:"available"`

	// Mobilty represents how many meters Human can move in a second.
	Mobility float64 `gorm:"not null" json:"mobility"`

	// Angle represents where Human looks for. The unit is radian.
	Angle float64 `gorm:"not null" json:"angle"`

	// Lifespan represents how many seconds Human can live.
	// Human will die after specific term with keeping stay
	// in order to save memory resources.
	Lifespan float64 `gorm:"not null" json:"lifespan"`

	// Progress is [0,1] value representing how much Human proceed current task.
	Progress float64 `gorm:"not null" json:"progress"`

	On Standing `gorm:"-" json:"-"`

	M          *Model     `gorm:"-" json:"-"`
	From       *Residence `gorm:"-" json:"-"`
	To         *Company   `gorm:"-" json:"-"`
	onPlatform *Platform
	onTrain    *Train
	out        map[uint]*Step

	FromID     uint `gorm:"not null" json:"rid"`
	ToID       uint `gorm:"not null" json:"cid"`
	PlatformID uint `                json:"pid,omitempty"`
	TrainID    uint `                json:"tid,omitempty"`
}

// NewHuman create instance
func (m *Model) NewHuman(o *Player, x float64, y float64) *Human {
	h := &Human{
		Base:  NewBase(m.GenID(HUMAN)),
		Owner: NewOwner(o),
		Point: NewPoint(x, y),
	}
	h.Init(m)
	h.Resolve()
	h.Marshal()
	m.Add(h)

	h.GenOutSteps()
	return h
}

// GenOutSteps Generate Step for Human.
// It's depend on where Human stay.
func (h *Human) GenOutSteps() {
	switch h.On {
	case OnGround:
		// h - C for destination
		h.M.NewStep(h, h.To)
		// h -> G
		for _, g := range h.M.Gates {
			h.M.NewStep(h, g)
		}
	case OnPlatform:
		// h - G, P on Human
		h.M.NewStep(h, h.onPlatform)
		h.M.NewStep(h, h.onPlatform.WithGate)
	case OnTrain:
		// do-nothing
	default:
		panic(fmt.Errorf("invalid type: %T %+v", h.On, h.On))
	}
}

// Idx returns unique id field.
func (h *Human) Idx() uint {
	return h.ID
}

// Type returns type of entitiy
func (h *Human) Type() ModelType {
	return HUMAN
}

// Init creates map.
func (h *Human) Init(m *Model) {
	h.M = m
	h.out = make(map[uint]*Step)
}

// Pos returns entities' position
func (h *Human) Pos() *Point {
	return &h.Point
}

// IsIn returns it should be view or not.
func (h *Human) IsIn(x float64, y float64, scale float64) bool {
	return h.Pos().IsIn(x, y, scale)
}

// OutSteps returns where it can go to
func (h *Human) OutSteps() map[uint]*Step {
	return h.out
}

// InSteps returns where it comes from
func (h *Human) InSteps() map[uint]*Step {
	return nil
}

// Marshal set id from reference
func (h *Human) Marshal() {
	h.FromID = h.From.ID
	h.ToID = h.To.ID
	if h.onPlatform != nil {
		h.PlatformID = h.onPlatform.ID
	}
	if h.onTrain != nil {
		h.TrainID = h.onTrain.ID
	}
}

func (h *Human) UnMarshal() {
	h.Resolve(
		h.M.Find(RESIDENCE, h.FromID),
		h.M.Find(COMPANY, h.ToID))
	// nullable fields
	if h.PlatformID != ZERO {
		h.Resolve(h.M.Find(PLATFORM, h.PlatformID))
	}
	if h.TrainID != ZERO {
		h.Resolve(h.M.Find(TRAIN, h.TrainID))
	}
}

// Resolve set reference
func (h *Human) Resolve(args ...interface{}) {
	for _, raw := range args {
		switch obj := raw.(type) {
		case *Residence:
			h.From = obj
			obj.Resolve(h)
		case *Company:
			h.To = obj
			obj.Resolve(h)
		case *Platform:
			h.onPlatform = obj
			obj.Resolve(h)
		case *Train:
			h.onTrain = obj
			obj.Resolve(h)
		default:
			panic(fmt.Errorf("invalid type: %T %+v", obj, obj))
		}
	}
	h.Marshal()
}

// BeforeDelete deltes refernce of related entity
func (h *Human) BeforeDelete() {
	delete(h.From.Targets, h.ID)
	delete(h.To.Targets, h.ID)
	if h.onPlatform != nil {
		h.onPlatform.Occupied--
		delete(h.onPlatform.Passengers, h.ID)
	}
	if h.onTrain != nil {
		h.onTrain.Occupied--
		delete(h.onTrain.Passengers, h.ID)
	}
}

func (h *Human) UnResolve(args ...interface{}) {
	for _, raw := range args {
		switch obj := raw.(type) {
		case *Platform:
			h.onPlatform = nil
			h.PlatformID = ZERO
		case *Train:
			h.onTrain = nil
			h.TrainID = ZERO
		default:
			panic(fmt.Errorf("invalid type: %T %+v", obj, obj))
		}
	}
	h.Marshal()
}

// Permits represents Player is permitted to control
func (h *Human) Permits(o *Player) bool {
	return true
}

// CheckDelete check remain relation.
func (h *Human) CheckDelete() error {
	return nil
}

func (h *Human) Delete() {
	h.M.Delete(h)
}

// TurnTo make Human turn head to dest.
func (h *Human) turnTo(dest Locationable) *Human {
	h.Angle = math.Atan2(dest.Pos().Y-h.Y, dest.Pos().X-h.X)
	return h
}

// Move make Human walk forward to dist.
// If Human exhaust Available time, then stop.
func (h *Human) move(dist float64) *Human {
	capacity := h.Available * h.Mobility

	if dist > capacity {
		dist = capacity
		h.Available = 0
	} else {
		h.Available -= dist / h.Mobility
	}

	h.X += dist * math.Cos(h.Angle)
	h.Y += dist * math.Sin(h.Angle)
	return h
}

// WalkTo make Human walk to dest point.
// If Human cannot reach it, proceed forward as possible.
func (h *Human) WalkTo(dest Locationable) *Human {
	h.turnTo(dest).move(h.Dist(dest))
	return h
}

func (h *Human) GetIn(t *Train) *Human {
	//TODO
	return h
}

func (h *Human) GetOffForce() *Human {
	//TODO
	return h
}

func (h *Human) GetOff(platform *Platform) *Human {
	//TODO
	return h
}

func (h *Human) Enter(from *Gate, to *Platform) *Human {
	//TODO
	return h
}

func (h *Human) Exit(from *Platform, to *Gate) *Human {
	//TODO
	return h
}

func (h *Human) ExitForce() *Human {
	//TODO
	return h
}

func (h *Human) ShouldGetIn(to *Train) bool {
	// TODO
	return false
}

func (h *Human) ShouldGetOff(from *Train) bool {
	// TODO
	return false
}

// OnPlatform return platform human stands
func (h *Human) OnPlatform() *Platform {
	return h.onPlatform
}

// SetOnPlatform changes self changed status for backup
func (h *Human) SetOnPlatform(v *Platform) {
	h.onPlatform = v
	v.Resolve(h)
	h.Change()
}

// OnTrain return next field
func (h *Human) OnTrain() *Train {
	return h.onTrain
}

// SetOnTrain changes self changed status for backup
func (h *Human) SetOnTrain(v *Train) {
	h.onTrain = v
	v.Resolve(h)
	h.Change()
}

func (h *Human) IsNew() bool {
	return h.Base.IsNew()
}

// IsChanged returns true when it is changed after Backup()
func (h *Human) IsChanged(after ...time.Time) bool {
	return h.Base.IsChanged(after...)
}

// Reset set status as not changed
func (h *Human) Reset() {
	h.Base.Reset()
}

// String represents status
func (h *Human) String() string {
	h.Marshal()
	pstr, tstr := "", ""
	if h.onPlatform != nil {
		pstr = fmt.Sprintf(",p=%d", h.onPlatform.ID)
	}
	if h.onTrain != nil {
		tstr = fmt.Sprintf(",t=%d", h.onTrain.ID)
	}

	return fmt.Sprintf("%s(%d):r=%d,c=%d%s%s,a=%.1f,l=%.1f,%%=%.2f:%v",
		h.Type().Short(),
		h.ID, h.FromID, h.ToID, pstr, tstr, h.Available, h.Lifespan, h.Progress, h.Pos())
}
