package entities

import (
	"testing"

	"github.com/yasshi2525/RushHour/auth"
	"github.com/yasshi2525/RushHour/config"
)

func TestGate(t *testing.T) {
	a, _ := auth.GetAuther(config.CnfAuth{Key: "----------------"})
	t.Run("NewGate", func(t *testing.T) {
		m := NewModel(config.CnfEntity{}, a)
		m.NewCompany(0, 0)
		m.NewResidence(0, 0)
		o := m.NewPlayer()
		st := m.NewStation(o)
		g := m.NewGate(st)

		TestCases{
			{"O", g.O, o},
			{"st", g.InStation, st},
			{"stID", g.StationID, st.ID},
			{"st.g", st.Gate, g},
			{"st.gID", st.GateID, g.ID},
			{"in", len(g.InSteps()), 1},
			{"out", len(g.OutSteps()), 1},
			{"model", m.Gates[g.Idx()], g},
			{"s", len(m.Steps), 3},
		}.Assert(t)
	})
	t.Run("Delete", func(t *testing.T) {
		m := NewModel(config.CnfEntity{}, a)
		m.NewCompany(0, 0)
		m.NewResidence(0, 0)
		o := m.NewPlayer()
		st := m.NewStation(o)
		g := m.NewGate(st)

		g.CheckDelete()
		g.Delete()

		TestCases{
			{"o", len(o.Gates), 0},
			{"s", len(m.Steps), 1},
			{"model", len(m.Gates), 0},
		}.Assert(t)
	})
}
