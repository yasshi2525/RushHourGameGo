package services

import (
	"fmt"
	"time"

	"github.com/revel/revel"

	"github.com/yasshi2525/RushHour/app/entities"
	"github.com/yasshi2525/RushHour/app/services/route"
)

// CreateRailLine create RailLine
func CreateRailLine(o *entities.Player, name string, ext bool) (*entities.RailLine, error) {
	l := Model.NewRailLine(o)
	l.Name = name
	l.AutoExt = ext

	return l, nil
}

// StartRailLine start RailLine at Station
func StartRailLine(
	o *entities.Player,
	l *entities.RailLine,
	p *entities.Platform) error {
	if err := CheckAuth(o, l); err != nil {
		return err
	}
	if err := CheckAuth(o, p); err != nil {
		return err
	}
	if len(l.Tasks) > 0 {
		return fmt.Errorf("already registered %v", l.Tasks)
	}
	if rn := p.OnRailNode; len(rn.OutEdge) > 0 {
		model, _ := route.SearchRail(o, Const.Routing.Worker)
		n := model[rn.ID].Nodes[entities.RAILNODE][rn.ID]
		return StartRailLineEdge(o, l, Model.RailEdges[n.ViaEdge.ID])
	}
	Model.NewLineTaskDept(l, p)
	return nil
}

func StartRailLineEdge(o *entities.Player,
	l *entities.RailLine,
	re *entities.RailEdge) error {
	if err := CheckAuth(o, l); err != nil {
		return err
	}
	if err := CheckAuth(o, re); err != nil {
		return err
	}
	if len(l.Tasks) > 0 {
		return fmt.Errorf("already registered %v", l.Tasks)
	}
	var lt *entities.LineTask
	if p := re.FromNode.OverPlatform; p != nil {
		lt = Model.NewLineTaskDept(l, p)
	}
	lt = Model.NewLineTask(lt, re, false)
	if p := re.ToNode.OverPlatform; p != nil {
		lt = Model.NewLineTaskDept(l, p, lt)
	}
	lt = Model.NewLineTask(lt, re.Reverse, false)
	RingRailLine(o, l)
	return nil
}

// InsertLineTaskRailEdge corrects RailLine for new RailEdge
// RailEdge.From must be original RailNode.
// RailEdge.To   must be      new RailPoint.
//
// Before (a) ---------------> (b) -> (c)
// After  (a) -> (X) -> (a) -> (b) -> (c)
//
// RailEdge : (a) -> (X)
func InsertLineTaskRailEdge(o *entities.Player, l *entities.RailLine, re *entities.RailEdge, pass bool) error {
	if err := CheckAuth(o, re); err != nil {
		return err
	}

	// extract tasks which direct origin
	// find (a) -> (b)
	bases := []*entities.LineTask{}

	for _, lt := range re.FromNode.InTasks {
		if lt.RailLine == l {
			bases = append(bases, lt)
		}
	}

	for _, base := range bases {
		next := base.Next() // = (b) -> (c)

		inter, _ := AttachLineTask(o, base, re, pass)         // = (a) -> (X)
		inter, _ = AttachLineTask(o, inter, re.Reverse, pass) // = (X) -> (a)

		// when (X) is station and is stopped, append "dept" task after it
		if inter.TaskType == entities.OnStopping && next != nil && next.TaskType != entities.OnDeparture {
			inter = Model.NewLineTaskDept(inter.RailLine, inter.Dest, inter)

		}
		inter.SetNext(next) // (a) -> (b) -> (c)

		// recalculate transport cost if RailLine loops
		if inter.RailLine.IsRing() {
			delStepRailLine(inter.RailLine)
			genStepRailLine(inter.RailLine)
		}
	}
	return nil
}

func InsertLineTaskStation(o *entities.Player, st *entities.Station, pass bool) error {
	if err := CheckAuth(o, st); err != nil {
		return err
	}

	// find LineTask such as dept from new station point
	for _, lt := range st.Platform.OnRailNode.OutTasks {
		// set dest  from edge.from.overPlatform
		lt.Resolve(lt.Moving)
	}

	// find LineTask such as dest to new station point
	// cache once bacause it will be appended after that
	bases := []*entities.LineTask{}
	for _, lt := range st.Platform.OnRailNode.InTasks {
		bases = append(bases, lt)
	}

	for _, lt := range bases {
		if pass {
			// change move -> pass
			lt.TaskType = entities.OnPassing
		} else {
			// change move -> stop
			lt.TaskType = entities.OnStopping
			// insert dest
			next := lt.Next()
			inter := Model.NewLineTaskDept(lt.RailLine, st.Platform, lt)
			inter.SetNext(next)

		}
		// set dest
		lt.Resolve(lt.Moving) // register dest from edge.to.overPlatform
	}
	return nil
}

// AttachLineTask attaches new RailEdge
// Need to update Step after call
func AttachLineTask(o *entities.Player, tail *entities.LineTask, newer *entities.RailEdge, pass bool) (*entities.LineTask, error) {
	if err := CheckAuth(o, tail); err != nil {
		return nil, err
	}
	if err := CheckAuth(o, newer); err != nil {
		return nil, err
	}
	if tail.ToNode() != newer.FromNode {
		return nil, fmt.Errorf("unconnected RailEdge. %v != %v", tail.ToNode(), newer.FromNode)
	}

	// when task is "stop", append task "departure"
	if tail.TaskType == entities.OnStopping {
		tail = Model.NewLineTaskDept(tail.RailLine, tail.Dest, tail)
	}

	tail = Model.NewLineTask(tail, newer, pass)

	return tail, nil
}

// RingRailLine connects tail and head
func RingRailLine(o *entities.Player, l *entities.RailLine) (bool, error) {
	if err := CheckAuth(o, l); err != nil {
		return false, err
	}
	// Check RainLine is not ringing
	if l.CanRing() {
		head, tail := l.Borders()
		tail.SetNext(head)
		genStepRailLine(l)
		return true, nil
	}
	return false, nil
}

func CompleteRailLine(o *entities.Player, l *entities.RailLine) (bool, error) {
	if err := CheckAuth(o, l); err != nil {
		return false, err
	}
	if len(l.Tasks) == 0 || l.IsRing() {
		return false, nil
	}
	head, tail := l.Borders()
	route, _ := route.SearchRail(l.Own, Const.Routing.Worker)
	n := route[head.FromNode().ID].Nodes[entities.RAILNODE][tail.ToNode().ID]
	e := n.ViaEdge
	for e != nil {
		if tail.TaskType == entities.OnStopping {
			tail = Model.NewLineTaskDept(l, tail.Dest, tail)
		}
		tail = Model.NewLineTask(tail, Model.RailEdges[e.ID], false)
		e = e.ToNode.ViaEdge
	}
	// [DEBUG]
	lineValidation(l)
	return true, nil
}

// delStepRailLine discards old step
func delStepRailLine(l *entities.RailLine) {
	for _, s := range l.Steps {
		Model.Delete(s)
	}
}

// genStepRailLine generate Step P <-> P
func genStepRailLine(l *entities.RailLine) {
	tracks := route.SearchRailLine(l, Const.Routing.Worker)
	for _, tr := range tracks {
		tr.ExportStep(Model)
	}
}

// [DEBUG]
func lineValidation(l *entities.RailLine) {
	var headCnt, tailCnt, loopSize int
	var deadloop, smallloop bool

	for _, lt := range l.Tasks {
		if lt.Before() == nil {
			headCnt++
		}
		if lt.Next() == nil {
			tailCnt++
		}
	}

	if headCnt > 1 {
		revel.AppLog.Errorf("[DEBUG] MULTI HEAD DETECTED!")
	}

	if tailCnt > 1 {
		revel.AppLog.Errorf("[DEBUG] MULTI TAIL DETECTED!")
	}

	var top *entities.LineTask
	for _, top = range l.Tasks {
		break
	}

	if top != nil {
		lt := top.Next()
		for lt != nil && lt != top {
			lt = lt.Next()
			if loopSize > len(l.Tasks) {
				revel.AppLog.Errorf("[DEBUG] DEAD LOOP DETECTED: lt(%d)", lt.ID)
				deadloop = true
				break
			}
			loopSize++
		}
		if lt == top && loopSize < len(l.Tasks)-1 {
			revel.AppLog.Errorf("[DEBUG] SMALL LOOP DETECTED: lt(%d)", lt.ID)
			smallloop = true
		}
	}

	if headCnt > 1 || tailCnt > 1 || deadloop || smallloop {
		dumpRailLine(l)
		time.Sleep(2 * time.Second)
		panic("error detected")
	}
}

func dumpRailLine(l *entities.RailLine) {
	for _, lt := range l.Tasks {
		revel.AppLog.Errorf("[DEBUG] %v", lt)
	}
}
