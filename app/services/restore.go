package services

import (
	"fmt"
	"reflect"
	"time"

	"github.com/yasshi2525/RushHour/app/services/auth"
	"github.com/yasshi2525/RushHour/app/services/route"

	"github.com/revel/revel"
	"github.com/yasshi2525/RushHour/app/entities"
)

// ZERO means nil
const ZERO = 0

// Restore get model from database
func Restore(withLock bool) {
	revel.AppLog.Info("start restore from database")
	defer revel.AppLog.Info("end restore from database")
	start := time.Now()
	if withLock {
		MuModel.Lock()
		defer MuModel.Unlock()
	}
	lock := time.Now()
	defer WarnLongExec(start, lock, Const.Perf.Restore.D, "restore", true)

	setNextID()
	fetchStatic()
	resolveStatic()
	for _, l := range Model.RailLines {
		lineValidation(l) // [DEBUG]
	}
	genDynamics()
}

// setNextID set max id as NextID from database for Restore()
func setNextID() {
	for _, key := range entities.TypeList {
		if !key.IsDB() {
			continue
		}
		var maxID struct {
			V uint64
		}
		sql := fmt.Sprintf("SELECT max(id) as v FROM %s", key.Table())
		if err := db.Raw(sql).Scan(&maxID).Error; err == nil {
			Model.NextIDs[key] = &maxID.V
		} else {
			panic(err)
		}
	}
}

// fetchStatic selects records for Restore()
func fetchStatic() {
	var cnt int
	for _, key := range entities.TypeList {
		if !key.IsDB() {
			continue
		}
		// select文組み立て
		if rows, err := db.Table(key.Table()).Where("deleted_at is null").Rows(); err == nil {
			for rows.Next() {
				// 対応する Struct を作成
				obj := key.Obj(Model).(entities.Persistable)
				if err := db.ScanRows(rows, obj); err == nil {
					obj.P().Reset()

					// Model に登録
					Model.Values[key].SetMapIndex(reflect.ValueOf(obj.B().Idx()), reflect.ValueOf(obj))
					cnt++
				} else {
					panic(err)
				}
			}
		} else {
			panic(err)
		}
	}
	revel.AppLog.Infof("restored %d entities", cnt)
}

// resolveStatic set pointer from id for Restore()
func resolveStatic() {
	for _, key := range entities.TypeList {
		if !key.IsDB() {
			continue
		}
		Model.ForEach(key, func(obj entities.Entity) {
			obj.(entities.Migratable).UnMarshal()
			Model.RootCluster.Add(obj)
		})
	}
}

// genDynamics create Dynamic instances
func genDynamics() {
	for _, o := range Model.Players {
		Model.Tokens[o.Token] = o
		hash := auth.Digest(auth.Decrypt(o.LoginID))
		Model.Logins[o.Auth][hash] = o
		route.RefreshTracks(o, Const.Routing.Worker)
	}
	for _, r := range Model.Residences {
		r.GenOutSteps()
	}
	for _, g := range Model.Gates {
		g.GenOutSteps()
	}
	for _, p := range Model.Platforms {
		p.GenOutSteps()
	}
	for _, l := range Model.RailLines {
		route.RefreshTransports(l, Const.Routing.Worker)
	}
	for _, h := range Model.Humans {
		h.GenOutSteps()
	}
}
