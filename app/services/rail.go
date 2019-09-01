package services

import (
	"fmt"

	"github.com/yasshi2525/RushHour/app/entities"
	"github.com/yasshi2525/RushHour/app/services/route"
)

// CreateRailNode create RailNode
func CreateRailNode(o *entities.Player, x float64, y float64, scale float64) (*entities.DelegateRailNode, error) {
	if err := CheckArea(x, y); err != nil {
		return nil, err
	}
	rn := Model.NewRailNode(o, x, y)
	AddOpLog("CreateRailNode", o, rn)

	if ch := Model.RootCluster.FindChunk(rn, scale); ch != nil {
		return ch.RailNode, nil
	}
	return nil, fmt.Errorf("invalid scale=%f", scale)
}

// RemoveRailNode remove RailNode
func RemoveRailNode(o *entities.Player, id uint) error {
	if rn, err := Model.DeleteIf(o, entities.RAILNODE, id); err != nil {
		return err
	} else {
		rn := rn.(*entities.RailNode)
		if o.ReRouting {
			route.RefreshTracks(o, Const.Routing.Worker)
		}
		for _, l := range o.RailLines {
			if l.ReRouting {
				route.RefreshTransports(l, Const.Routing.Worker)
			}
		}
		AddOpLog("RemoveRailNode", o, rn)
		return nil
	}
}

// ExtendRailNode extends Rail
func ExtendRailNode(o *entities.Player, from *entities.RailNode,
	x float64, y float64, scale float64) (*entities.DelegateRailNode, *entities.DelegateRailEdge, error) {
	if err := CheckAuth(o, from); err != nil {
		return nil, nil, err
	}
	if err := CheckArea(x, y); err != nil {
		return nil, nil, err
	}
	to, e1 := from.Extend(x, y)
	route.RefreshTracks(o, Const.Routing.Worker)
	for _, l := range o.RailLines {
		if l.ReRouting {
			route.RefreshTransports(l, Const.Routing.Worker)
		}
	}
	AddOpLog("ExtendRailNode", o, from, to, e1, e1.Reverse)

	fch := Model.RootCluster.FindChunk(from, scale)
	tch := Model.RootCluster.FindChunk(to, scale)
	if fch == nil || tch == nil {
		return nil, nil, fmt.Errorf("invalid scale=%f", scale)
	}
	return tch.RailNode, fch.OutRailEdges[tch.ID], nil
}

// RemoveRailEdge remove RailEdge
func RemoveRailEdge(o *entities.Player, id uint) error {
	if re, err := Model.DeleteIf(o, entities.RAILEDGE, id); err != nil {
		return err
	} else {
		re := re.(*entities.RailEdge)
		if o.ReRouting {
			route.RefreshTracks(o, Const.Routing.Worker)
		}
		for _, l := range o.RailLines {
			if l.ReRouting {
				route.RefreshTransports(l, Const.Routing.Worker)
			}
		}
		AddOpLog("RemoveRailEdge", o, re)
		return nil
	}
}
