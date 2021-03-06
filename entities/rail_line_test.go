package entities

import (
	"testing"

	"github.com/yasshi2525/RushHour/auth"
	"github.com/yasshi2525/RushHour/config"
)

func TestRailLine(t *testing.T) {
	a, _ := auth.GetAuther(config.CnfAuth{Key: "----------------"})
	c := config.CnfEntity{MaxScale: 16}
	t.Run("NewRailLine", func(t *testing.T) {
		m := NewModel(c, a)
		o := m.NewPlayer()
		l := m.NewRailLine(o)

		TestCases{
			{"O", l.O, o},
			{"o.l", o.RailLines[l.Idx()], l},
			{"model", m.RailLines[l.Idx()], l},
		}.Assert(t)
	})

	t.Run("StartPlatform", func(t *testing.T) {
		t.Run("auto ext", func(t *testing.T) {
			m := NewModel(c, a)
			o := m.NewPlayer()
			from := m.NewRailNode(o, 0, 0)
			_, re := from.Extend(10, 0)

			from.Tracks[re.ID] = make(map[uint]bool)
			from.Tracks[re.ID][from.ID] = true

			st := m.NewStation(o)
			g := m.NewGate(st)
			p := m.NewPlatform(from, g)
			l := m.NewRailLine(o)
			l.AutoExt = true

			head, _ := l.StartPlatform(p)

			TestCaseLineTasks{
				{"n0", OnDeparture, p},
				{"n0->n1", OnMoving, re},
				{"n1->n0", OnStopping, re.Reverse},
			}.Assert(t, head)
		})
		t.Run("manual", func(t *testing.T) {
			m := NewModel(c, a)
			o := m.NewPlayer()
			rn := m.NewRailNode(o, 0, 0)
			st := m.NewStation(o)
			g := m.NewGate(st)
			p := m.NewPlatform(rn, g)
			l := m.NewRailLine(o)

			head, _ := l.StartPlatform(p)

			TestCaseLineTasks{
				{"dep", OnDeparture, p},
			}.Assert(t, head)
		})
		t.Run("auto pass", func(t *testing.T) {
			m := NewModel(c, a)
			o := m.NewPlayer()
			rn := m.NewRailNode(o, 0, 0)
			st := m.NewStation(o)
			g := m.NewGate(st)
			p := m.NewPlatform(rn, g)
			l := m.NewRailLine(o)
			l.AutoPass = true

			head, tail := l.StartPlatform(p)

			TestCases{
				{"head", head, (*LineTask)(nil)},
				{"tail", tail, (*LineTask)(nil)},
				{"l", len(l.Tasks), 0},
			}.Assert(t)
		})
	})

	t.Run("Complement", func(t *testing.T) {
		m := NewModel(c, a)
		o := m.NewPlayer()
		from := m.NewRailNode(o, 0, 0)
		to, re := from.Extend(10, 0)

		from.Tracks[re.ID] = make(map[uint]bool)
		from.Tracks[re.ID][from.ID] = true
		to.Tracks[re.Reverse.ID] = make(map[uint]bool)
		to.Tracks[re.Reverse.ID][from.ID] = true

		st := m.NewStation(o)
		g := m.NewGate(st)
		p := m.NewPlatform(from, g)
		l := m.NewRailLine(o)

		head, _ := l.StartPlatform(p)
		l.Complement()

		TestCaseLineTasks{
			{"n0", OnDeparture, p},
			{"n0->n1", OnMoving, re},
			{"n1->n0", OnStopping, re.Reverse},
		}.Assert(t, head)
	})

	t.Run("Delete", func(t *testing.T) {
		m := NewModel(c, a)
		o := m.NewPlayer()
		from := m.NewRailNode(o, 0, 0)
		_, re := from.Extend(10, 0)
		l := m.NewRailLine(o)
		l.AutoExt = true
		l.StartEdge(re)

		l.Delete()

		TestCases{
			{"lt", len(m.LineTasks), 0},
			{"model", len(m.RailLines), 0},
		}.Assert(t)
	})
}
