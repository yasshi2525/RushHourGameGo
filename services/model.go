package services

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"reflect"
	"time"

	"github.com/yasshi2525/RushHour/entities"
)

// Operation represents request for model
type Operation struct {
	Source string
	Target entities.ModelType
	ID     uint
	X      float64
	Y      float64
	Op     string
	OName  string
}

func (op *Operation) String() string {
	return fmt.Sprintf("Op(%s:%s) %s %v(%d) at (%.2f, %.2f)", op.Source, op.OName, op.Op, op.Target, op.ID, op.X, op.Y)
}

var modelChannel chan *Operation

var mkFuncs map[entities.ModelType]interface{}
var rmFuncs map[entities.ModelType]interface{}

// StartModelWatching setup watching model
func StartModelWatching() {

	modelChannel = make(chan *Operation, conf.Game.Service.Procedure.Queue)

	mkFuncs = make(map[entities.ModelType]interface{})
	mkFuncs[entities.PLAYER] = CreatePlayer
	mkFuncs[entities.RESIDENCE] = CreateResidence
	mkFuncs[entities.COMPANY] = CreateCompany
	mkFuncs[entities.RAILNODE] = CreateRailNode
	mkFuncs[entities.RAILEDGE] = ExtendRailNode
	mkFuncs[entities.STATION] = CreateStation
	mkFuncs[entities.RAILLINE] = CreateRailLine
	mkFuncs[entities.TRAIN] = CreateTrain

	rmFuncs = make(map[entities.ModelType]interface{})
	rmFuncs[entities.RESIDENCE] = RemoveResidence
	rmFuncs[entities.COMPANY] = RemoveCompany
	rmFuncs[entities.RAILNODE] = RemoveRailNode
	rmFuncs[entities.RAILEDGE] = RemoveRailEdge
	rmFuncs[entities.STATION] = RemoveStation
	rmFuncs[entities.RAILLINE] = RemoveRailLine
	rmFuncs[entities.TRAIN] = RemoveTrain

	go watchModel()
	log.Println("model watching was successfully started.")
}

// StopModelWatching closes channel
func StopModelWatching() {
	if modelChannel != nil {
		close(modelChannel)
		log.Println("model watching was successfully stopped.")
	}
}

func watchModel() {
	for msg := range modelChannel {
		start := time.Now()
		lock := processMsg(msg)
		WarnLongExec(start, lock, conf.Game.Service.Perf.Operation.D, fmt.Sprintf("operation(%v)", msg))
	}
	log.Println("model watching channel was closed.")
}

func processMsg(msg *Operation) time.Time {
	MuModel.Lock()
	defer MuModel.Unlock()
	lock := time.Now()

	owner, _ := PasswordSignIn(msg.OName, msg.OName)
	size := 1 << (conf.Game.Entity.MaxScale - conf.Game.Entity.MinScale)
	rnd := float64(size) * rand.Float64()

	switch msg.Op {
	case "create":
		rv := reflect.ValueOf(mkFuncs[msg.Target])
		switch msg.Target {
		case entities.PLAYER:
			level := entities.Normal
			if msg.OName == "admin" {
				level = entities.Admin
			}
			CreatePlayer(msg.OName, msg.OName, msg.OName, rand.Intn(360), level)
		case entities.RESIDENCE:
			fallthrough
		case entities.COMPANY:
			rv.Call([]reflect.Value{
				reflect.ValueOf(owner),
				reflect.ValueOf(msg.X),
				reflect.ValueOf(msg.Y),
			})
		case entities.RAILNODE:
			if rand.Float64() < 0.1 {
				rv.Call([]reflect.Value{
					reflect.ValueOf(owner),
					reflect.ValueOf(msg.X),
					reflect.ValueOf(msg.Y),
					reflect.ValueOf(10),
				})
			}
		case entities.RAILEDGE:
			if raw := randEntity(owner, entities.RAILNODE); raw != nil {
				rn := raw.(*entities.RailNode)
				theta := 2 * math.Pi * rand.Float64()

				p := &entities.Point{
					X: rn.X + math.Cos(theta)*rnd,
					Y: rn.Y + math.Sin(theta)*rnd,
				}

				for !p.IsIn(0, 0, float64(conf.Game.Entity.MaxScale)) {
					theta = 2 * math.Pi * rand.Float64()
					rnd := float64(size) * rand.Float64()

					p = &entities.Point{
						X: rn.X + math.Cos(theta)*rnd,
						Y: rn.Y + math.Sin(theta)*rnd,
					}
				}

				rv.Call([]reflect.Value{
					reflect.ValueOf(owner),
					reflect.ValueOf(raw),
					reflect.ValueOf(p.X),
					reflect.ValueOf(p.Y),
					reflect.ValueOf(10),
				})

				if oth := randEntity(owner, entities.RAILNODE); oth != nil {
					to := oth.(*entities.RailNode)
					if rn != to {
						var connected bool
						for _, e := range rn.OutEdges {
							if e.ToNode == to {
								connected = true
								break
							}
						}

						if !connected && rn.Point.Dist(&to.Point) < math.Pow(2, 8) {
							ConnectRailNode(owner, rn, to, 10)
						}
					}
				}
			}
			if raw := randEntity(owner, entities.RAILEDGE); raw != nil {
				re := raw.(*entities.RailEdge)
				d := re.ToNode.Point.Sub(&re.FromNode.Point)
				theta := math.Atan2(d.Y, d.Y) + (rand.Float64()-0.5)*math.Pi/32

				p := &entities.Point{
					X: re.ToNode.X + math.Cos(theta)*rnd,
					Y: re.ToNode.Y + math.Sin(theta)*rnd,
				}

				for !p.IsIn(0, 0, float64(conf.Game.Entity.MaxScale)) {
					rnd := float64(size) * rand.Float64()
					theta = math.Atan2(d.Y, d.Y) + rand.Float64() - 0.5

					p = &entities.Point{
						X: re.ToNode.X + math.Cos(theta)*rnd,
						Y: re.ToNode.Y + math.Sin(theta)*rnd,
					}
				}

				ExtendRailNode(owner, re.ToNode, p.X, p.Y, 10)
			}
		case entities.STATION:
			if rn := randEntity(owner, entities.RAILNODE); rn != nil {
				rv.Call([]reflect.Value{
					reflect.ValueOf(owner),
					reflect.ValueOf(rn),
					reflect.ValueOf("NoName")})
			}
		case entities.RAILLINE:
			rv.Call([]reflect.Value{
				reflect.ValueOf(owner),
				reflect.ValueOf("NoName"),
				reflect.ValueOf(rand.Intn(2) == 0),
				reflect.ValueOf(rand.Intn(2) == 0)})
		case entities.LINETASK:
			l := randEntity(owner, entities.RAILLINE)
			if l != nil {
				if p := randEntity(owner, entities.PLATFORM); p != nil {
					l, p := l.(*entities.RailLine), p.(*entities.Platform)
					StartRailLine(owner, l, p)
				}
				if re := randEntity(owner, entities.RAILEDGE); re != nil {
					re := re.(*entities.RailEdge)
					StartRailLineEdge(owner, l.(*entities.RailLine), re)
					InsertLineTaskRailEdge(owner, l.(*entities.RailLine), re)
				}
				if rand.Intn(2) == 0 {
					ComplementRailLine(owner, l.(*entities.RailLine))
				}
			}
		case entities.TRAIN:
			t, _ := CreateTrain(owner, "NoName")
			l := randEntity(owner, entities.RAILLINE)
			if l != nil {
				DeployTrain(owner, t, l.(*entities.RailLine))
			}
		}
	case "remove":
		rv := reflect.ValueOf(rmFuncs[msg.Target])
		if !rv.IsValid() {
			break
		}
		if msg.ID == 0 {
			var ok bool
			msg.ID, ok = randID(msg.Target, owner)
			if !ok {
				break
			}
		}
		rv.Call([]reflect.Value{
			reflect.ValueOf(owner),
			reflect.ValueOf(msg.ID)})
	}
	return lock
}

// randID return random id existing in repository
func randID(t entities.ModelType, owner *entities.Player) (uint, bool) {
	mapdata := Model.Values[t]
	for _, key := range mapdata.MapKeys() {
		if mapdata.MapIndex(key).Interface().(entities.Entity).B().Permits(owner) {
			return uint(key.Uint()), true
		}
	}
	return 0, false
}

// UpdateModel queues user request.
func UpdateModel(msg *Operation) {
	//log.Printf("updatemodel op = %+v", *msg)
	select {
	case modelChannel <- msg:
	default:
		log.Printf("out of queue %+v", *msg)
	}
}

func randEntity(o *entities.Player, res entities.ModelType) interface{} {
	var entity interface{}
	Model.ForEach(res, func(obj entities.Entity) {
		if obj.B().Permits(o) {
			entity = obj
		}
	})
	return entity
}
