package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/yasshi2525/RushHour/entities"
	"github.com/yasshi2525/RushHour/services"
)

type deptResponse struct {
	RailNode *entities.DelegateRailNode `json:"rn"`
}

// Depart returns result of rail node creation
// @Description result of rail node creation
// @Tags deptResponse
// @Summary depart
// @Accept json
// @Produce json
// @Param x body number true "x coordinate"
// @Param y body number true "y coordinate"
// @Param scale body number true "width,height(100%)=2^scale"
// @Success 200 {object} deptResponse "created rail node"
// @Failure 400 {object} errInfo "reason of fail"
// @Failure 401 {object} errInfo "invalid jwt"
// @Failure 503 {object} errInfo "under maintenance"
// @Router /rail_nodes [post]
func Depart(c *gin.Context) {
	o := c.MustGet(keyOwner).(*entities.Player)
	params := pointRequest{}
	if err := c.ShouldBind(&params); err != nil {
		c.Set(keyErr, err)
	} else {
		if rn, err := services.CreateRailNode(o, params.X, params.Y, params.Scale); err != nil {
			c.Set(keyErr, err)
		} else {
			c.Set(keyOk, &deptResponse{RailNode: rn})
		}
	}
}

type extendRequest struct {
	RailNode uint `form:"rnid" json:"rnid" validate:"required,numeric"`
}

type extendResponse struct {
	RailNode *entities.DelegateRailNode `json:"rn"`
	In       *entities.DelegateRailEdge `json:"e1"`
	Out      *entities.DelegateRailEdge `json:"e2"`
}

// Extend returns result of rail node extension
// @Description result of rail node extension
// @Tags extendResponse
// @Summary extend
// @Accept json
// @Produce json
// @Param x body number true "x coordinate"
// @Param y body number true "y coordinate"
// @Param scale body number true "width,height(100%)=2^scale"
// @Param rnid body integer true "tail rail node id"
// @Success 200 {object} extendResponse "extend rail node"
// @Failure 400 {object} errInfo "reason of fail"
// @Failure 401 {object} errInfo "invalid jwt"
// @Failure 503 {object} errInfo "under maintenance"
// @Router /rail_nodes/extend [post]
func Extend(c *gin.Context) {
	o := c.MustGet(keyOwner).(*entities.Player)
	p := pointRequest{}
	if err := c.ShouldBindBodyWith(&p, binding.JSON); err != nil {
		c.Set(keyErr, err)
	} else {
		ex := extendRequest{}
		if err := c.ShouldBindBodyWith(&ex, binding.JSON); err != nil {
			c.Set(keyErr, err)
		} else if rn, err := validateEntity(entities.RAILNODE, ex.RailNode); err != nil {
			c.Set(keyErr, err)
		} else {
			if to, re, err := services.ExtendRailNode(o, rn.(*entities.RailNode), p.X, p.Y, p.Scale); err != nil {
				c.Set(keyErr, err)
			} else {
				c.Set(keyOk, &extendResponse{to, re, re.Reverse})
			}
		}
	}
}

type connectRequest struct {
	From uint `form:"from" json:"from" validate:"required,numeric"`
	To   uint `form:"to" json:"to" validate:"required,numeric"`
}

type connectResponse struct {
	In  *entities.DelegateRailEdge `json:"e1"`
	Out *entities.DelegateRailEdge `json:"e2"`
}

// Connect returns result of rail connection
// @Description result of rail node connection
// @Tags connectResponse
// @Summary connect
// @Accept json
// @Produce json
// @Param x body number true "x coordinate"
// @Param y body number true "y coordinate"
// @Param scale body number true "width,height(100%)=2^scale"
// @Param from body integer true "from rail node id"
// @Param to body integer true "to rail node id"
// @Success 200 {object} connectResponse "connect rail node"
// @Failure 400 {object} errInfo "reason of fail"
// @Failure 401 {object} errInfo "invalid jwt"
// @Failure 503 {object} errInfo "under maintenance"
// @Router /rail_nodes/connect [post]
func Connect(c *gin.Context) {
	o := c.MustGet(keyOwner).(*entities.Player)
	sc := scaleRequest{}
	if err := c.ShouldBindBodyWith(&sc, binding.JSON); err != nil {
		c.Set(keyErr, err)
	} else {
		ex := connectRequest{}
		if err := c.ShouldBindBodyWith(&ex, binding.JSON); err != nil {
			c.Set(keyErr, err)
		} else if from, err := validateEntity(entities.RAILNODE, ex.From); err != nil {
			c.Set(keyErr, err)
		} else if to, err := validateEntity(entities.RAILNODE, ex.To); err != nil {
			c.Set(keyErr, err)
		} else {
			if re, err := services.ConnectRailNode(o, from.(*entities.RailNode), to.(*entities.RailNode), sc.Scale); err != nil {
				c.Set(keyErr, err)
			} else {
				c.Set(keyOk, &connectResponse{re, re.Reverse})
			}
		}
	}
}

type removeRailNodeRequest struct {
	RailNode uint `form:"id" json:"id" validate:"required,numeric"`
}

type removeRailNodeResponse struct {
	RailNode uint `json:"id"`
}

// RemoveRailNode returns result of rail deletion
// @Description result of rail node deletion
// @Tags removeRailNodeResponse
// @Summary remove rail node
// @Accept json
// @Produce json
// @Param id body integer true "rail node id"
// @Success 200 {object} removeRailNodeResponse "connect rail node"
// @Success 400 {object} errInfo "reason of fail"
// @Failure 401 {object} errInfo "invalid jwt"
// @Failure 503 {object} errInfo "under maintenance"
// @Router /rail_nodes [delete]
func RemoveRailNode(c *gin.Context) {
	o := c.MustGet(keyOwner).(*entities.Player)
	params := removeRailNodeRequest{}
	if err := c.ShouldBind(&params); err != nil {
		c.Set(keyErr, err)
	} else if err := services.RemoveRailNode(o, params.RailNode); err != nil {
		c.Set(keyErr, err)
	} else {
		c.Set(keyOk, &removeRailNodeResponse{params.RailNode})
	}
}
