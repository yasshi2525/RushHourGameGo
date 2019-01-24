package entities

import (
	"fmt"
)

const EPS float64 = 0.00001

// Train carries Human from Station to Station.
type Train struct {
	Base
	Persistence
	Point
	Shape

	Capacity int `json:"capacity"`
	// Mobility represents how many Human can get off at the same time.
	Mobility int     `json:"mobility"`
	Speed    float64 `json:"speed"`
	Progress float64 `json:"progress"`
	Name     string  `gorm:"not null" json:"name"`
	Occupied int     `gorm:"-"        json:"occupied"`

	OnRailEdge *RailEdge `gorm:"-" json:"-"`
	OnPlatform *Platform `gorm:"-" json:"-"`
	task       *LineTask
	Passengers map[uint]*Human `gorm:"-" json:"-"`

	TaskID     uint `         json:"ltid,omitempty"`
	RailEdgeID uint `gorm:"-" json:"reid,omitempty"`
	PlatformID uint `gorm:"-" json:"pid,omitempty"`
}

// NewTrain creates instance
func (m *Model) NewTrain(o *Player, name string) *Train {
	pos := Point{}
	t := &Train{
		Base:        m.NewBase(TRAIN, o),
		Persistence: NewPersistence(),
		Point:       pos,
		Shape:       NewShapeNode(&pos),
		Capacity:    Const.Train.Capacity,
		Mobility:    Const.Train.Mobility,
		Speed:       Const.Train.Speed,
		Name:        name,
	}
	t.Init(m)
	t.Marshal()
	m.Add(t)
	return t
}

// B returns base information of this elements.
func (t *Train) B() *Base {
	return &t.Base
}

// P returns time information for database.
func (t *Train) P() *Persistence {
	return &t.Persistence
}

// S returns entities' position.
func (t *Train) S() *Shape {
	return &t.Shape
}

func (t *Train) UnLoad() {
	for _, h := range t.Passengers {
		h.Point = *t.Pos().Rand(Const.Train.Randomize)
		h.onTrain = nil
		h.TrainID = ZERO
		t.Occupied--
	}
}

func (t *Train) Step(sec float64) {
	if t.task == nil {
		return
	}
	for sec > EPS {
		switch t.task.TaskType {
		case OnDeparture:
			// [TODO] make human get off
			t.SetTask(t.task.next)
		default:
			t.task.Step(&t.Progress, &sec)
			if t.Progress > 1-EPS {
				t.SetTask(t.task.next)
			}
		}
		//revel.AppLog.Debugf("t(%d) sec = %f prod = %f: %v", t.ID, sec, t.Progress, t)
	}
	t.X, t.Y = t.task.Loc(t.Progress).Flat()
}

// Idx returns unique id field.
func (t *Train) Idx() uint {
	return t.ID
}

// Type returns type of entitiy
func (t *Train) Type() ModelType {
	return TRAIN
}

// Init makes map
func (t *Train) Init(m *Model) {
	t.Base.Init(TRAIN, m)
	t.Passengers = make(map[uint]*Human)
}

func (t *Train) SetTask(lt *LineTask) {
	if t.task != nil {
		if lt == nil {
			t.task.RailLine.UnResolve(t)
		}
		t.task.UnResolve(t)
	}
	t.task = lt
	t.Progress = 0
	if lt != nil {
		t.TaskID = lt.ID
		lt.Resolve(t)
		t.Shape = lt.Shape
	} else {
		t.UnLoad()
		t.TaskID = ZERO
		t.Shape = Shape{}
	}

	t.Point = *t.Shape.Div(t.Progress)
	t.Change()
	t.Marshal()
}

// Resolve set ID from reference
func (t *Train) Resolve(args ...Entity) {
	for _, raw := range args {
		switch obj := raw.(type) {
		case *Player:
			t.O = obj
			obj.Resolve(t)
		case *LineTask:
			t.task = obj
			obj.Resolve(t)
		case *RailEdge:
			t.OnRailEdge = obj
		case *Platform:
			t.OnPlatform = obj
		case *Human:
			t.Passengers[obj.ID] = obj
			t.Occupied++
		default:
			panic(fmt.Errorf("invalid type: %T %+v", obj, obj))
		}
	}
	t.Marshal()
}

// Marshal set id from reference
func (t *Train) Marshal() {
	if t.O != nil {
		t.OwnerID = t.O.ID
	}
	if t.task != nil {
		t.TaskID = t.task.ID
	}
	if t.OnRailEdge != nil {
		t.RailEdgeID = t.OnRailEdge.ID
	} else {
		t.RailEdgeID = ZERO
	}
	if t.OnPlatform != nil {
		t.PlatformID = t.OnPlatform.ID
	} else {
		t.PlatformID = ZERO
	}
}

func (t *Train) UnMarshal() {
	t.Resolve(t.M.Find(PLAYER, t.OwnerID))
	// nullable fields
	if t.TaskID != ZERO {
		t.Resolve(t.M.Find(LINETASK, t.TaskID))
	}
}

func (t *Train) UnResolve(args ...interface{}) {
	for _, raw := range args {
		switch obj := raw.(type) {
		default:
			panic(fmt.Errorf("invalid type: %T %+v", obj, obj))
		}
	}
}

// CheckDelete check remain relation.
func (t *Train) CheckDelete() error {
	return nil
}

func (t *Train) BeforeDelete() {
	t.UnLoad()
	t.O.UnResolve(t)
}

func (t *Train) Delete(force bool) {
	t.M.Delete(t)
}

// Task return task field
func (t *Train) Task() *LineTask {
	return t.task
}

// String represents status
func (t *Train) String() string {
	t.Marshal()
	ostr := ""
	if t.O != nil {
		ostr = fmt.Sprintf(":%s", t.O.Short())
	}
	ltstr := ""
	if t.task != nil {
		ltstr = fmt.Sprintf(",lt=%d", t.task.ID)
	}
	posstr := ""
	if t.Pos() != nil {
		posstr = fmt.Sprintf(":%s", t.Pos())
	}
	return fmt.Sprintf("%s(%v):h=%d/%d%s,%%=%.2f%s%s:%s", t.Type().Short(),
		t.ID, len(t.Passengers), t.Capacity, ltstr, t.Progress, posstr, ostr, t.Name)
}
