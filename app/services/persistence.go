package services

import (
	"fmt"
	"reflect"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"github.com/yasshi2525/RushHour/app/entities"
)

var (
	db      *gorm.DB
	logMode = false
)

type eachCallback func(v reflect.Value)

// InitPersistence prepares database connection and migrate
func InitPersistence() {
	db = connectDB()
	configureDB(db)
	migrateDB(db)
}

func connectDB() *gorm.DB {
	var (
		database     *gorm.DB
		driver, spec string
		found        bool
		err          error
	)
	if driver, found = revel.Config.String("db.driver"); !found {
		panic("db.drvier is not defined")
	}
	if spec, found = revel.Config.String("db.spec"); !found {
		panic("db.spec is not defined")
	}

	if database, err = gorm.Open(driver, spec); err != nil {
		panic("failed to connect database")
	}

	revel.AppLog.Info("connect database successfully")
	return database
}

func configureDB(database *gorm.DB) *gorm.DB {
	database.LogMode(logMode)
	db.SingularTable(true)
	return database
}

func migrateDB(database *gorm.DB) *gorm.DB {
	db.AutoMigrate(StaticInstances...)

	// Player has private resources
	for _, t := range []interface{}{
		&entities.RailNode{},
		&entities.RailEdge{},
		&entities.Station{},
		&entities.Platform{},
		&entities.Gate{},
		&entities.LineTask{},
		&entities.Line{},
		&entities.Train{},
	} {
		db.Model(t).AddForeignKey("owner_id", "player(id)", "RESTRICT", "RESTRICT")
	}

	// RailEdge connects RailNode
	db.Model(&entities.RailEdge{}).AddForeignKey("from_id", "rail_node(id)", "CASCADE", "RESTRICT")
	db.Model(&entities.RailEdge{}).AddForeignKey("to_id", "rail_node(id)", "CASCADE", "RESTRICT")

	// Station composes Platforms and Gates
	db.Model(&entities.Platform{}).AddForeignKey("on_id", "rail_node(id)", "RESTRICT", "RESTRICT")
	db.Model(&entities.Platform{}).AddForeignKey("in_id", "station(id)", "CASCADE", "RESTRICT")
	db.Model(&entities.Gate{}).AddForeignKey("in_id", "station(id)", "CASCADE", "RESTRICT")

	// Line composes LineTasks
	db.Model(&entities.LineTask{}).AddForeignKey("line_id", "line(id)", "CASCADE", "RESTRICT")
	// LineTask is chainable
	db.Model(&entities.LineTask{}).AddForeignKey("next_id", "line_task(id)", "SET NULL", "RESTRICT")

	// Train runs on a chain of Line
	db.Model(&entities.Train{}).AddForeignKey("task_id", "line_task(id)", "RESTRICT", "RESTRICT")

	// Human departs from Residence and destinates to Company
	db.Model(&entities.Human{}).AddForeignKey("from_id", "residence(id)", "RESTRICT", "RESTRICT")
	db.Model(&entities.Human{}).AddForeignKey("to_id", "company(id)", "RESTRICT", "RESTRICT")
	// Human is sometimes on Platform or on Train
	db.Model(&entities.Human{}).AddForeignKey("on_platform_id", "platform(id)", "RESTRICT", "RESTRICT")
	db.Model(&entities.Human{}).AddForeignKey("on_train_id", "train(id)", "RESTRICT", "RESTRICT")

	return db
}

// TerminatePersistence defines the end task before application shutdown
func TerminatePersistence() {
	closeDB()
}

func closeDB() {
	if err := db.Close(); err != nil {
		revel.AppLog.Error("failed to close the database", "error", err)
	}
	revel.AppLog.Info("disconnect database successfully")
}

type resultMax struct {
	MaxID uint64
}

// Restore get model from database
func Restore() {
	revel.AppLog.Info("DBリストア 開始")
	defer revel.AppLog.Info("DBリストア 終了")

	start := time.Now()
	defer WarnLongExec(start, 5, "DBリストア", true)

	MuStatic.Lock()
	defer MuStatic.Unlock()
	MuDynamic.Lock()
	defer MuDynamic.Unlock()

	setNextID()
	fetchStatic()
	resolveStatic()
	generateDynamics()
}

func setNextID() {
	for _, resource := range StaticTypes {
		sql := fmt.Sprintf("SELECT max(id) as max_id FROM %s", resource)
		var result resultMax

		if err := db.Raw(sql).Scan(&result).Error; err != nil {
			panic(fmt.Sprintf("cannot get max id of %s, %v", resource, err))
		}

		result.MaxID++
		NextID.Static[resource] = &result.MaxID
		revel.AppLog.Debugf("NextID.Static[%s] = %d", resource, *NextID.Static[resource])
	}

	for _, resource := range DynamicTypes {
		var i uint64 = 1
		NextID.Dynamic[resource] = &i
		revel.AppLog.Debugf("NextID.Dynamic[%s] = %d", resource, *NextID.Dynamic[resource])
	}
}

func fetchStatic() {
	for idx, resource := range StaticTypes {
		// select文組み立て
		table := fmt.Sprintf("%s", resource)
		rows, err := db.Table(table).Where("deleted_at is null").Rows()
		if err != nil {
			panic(fmt.Sprintf("failed to fetch: %s", err))
		}
		for rows.Next() {
			// 対応する Struct を作成
			obj := reflect.New(reflect.TypeOf(StaticInstances[idx]).Elem())
			objptr := reflect.Indirect(obj).Addr().Interface()

			if err := db.ScanRows(rows, objptr); err != nil {
				panic(err)
			}

			// Static に登録
			objid := reflect.ValueOf(objptr).Elem().FieldByName("ID")
			reflect.ValueOf(Static).Field(idx).SetMapIndex(objid, obj)
		}
	}
}

func resolveStatic() {
	revel.AppLog.Debug("resolveStatic")
	foreachStatic(func(raw reflect.Value) {
		obj := raw.Elem()
		rt := reflect.TypeOf(obj.Interface())

		// OwnerID -> Owner
		if _, ok := rt.FieldByName("Ownable"); ok {
			ownable := obj.FieldByName("Ownable")
			optr, oid := ownable.FieldByName("Owner"), ownable.FieldByName("OwnerID")
			owner := Static.Players[uint(oid.Uint())]
			optr.Set(reflect.ValueOf(owner))
		}
	})
}

func generateDynamics() {
	for _, r := range Static.Residences {
		// R -> C
		for _, c := range Static.Companies {
			createStep(&r.Junction, &c.Junction, 1.0)
		}
		// R -> G
		for _, g := range Static.Gates {
			createStep(&r.Junction, &g.Junction, 1.0)
		}
	}
	for _, c := range Static.Companies {
		// G -> C
		for _, g := range Static.Gates {
			createStep(&g.Junction, &c.Junction, 1.0)
		}
	}
}

func foreachStatic(callback eachCallback) {
	rt, rv := reflect.TypeOf(Static), reflect.ValueOf(Static)
	for i := 0; i < rt.NumField(); i++ {
		if f := rv.Field(i); f.Kind() == reflect.Map {
			for _, e := range f.MapKeys() {
				callback(f.MapIndex(e))
			}
		} else {
			revel.AppLog.Warnf("%s is not map", f.Kind().String())
		}
	}
}

// Backup set model to database
func Backup() {
	start := time.Now()
	revel.AppLog.Info("バックアップ 開始")
	defer WarnLongExec(start, 2, "バックアップ", true)
	defer revel.AppLog.Info("バックアップ 終了")

	MuStatic.RLock()
	defer MuStatic.RUnlock()

	MuDynamic.RLock()
	defer MuDynamic.RUnlock()

	tx := db.Begin()

	// resolve mutable refer
	// Object -> ObjectID
	for _, val := range Static.LineTasks {
		val.ResolveRef()
	}
	for _, val := range Static.Trains {
		val.ResolveRef()
	}
	for _, val := range Static.Humans {
		val.ResolveRef()
	}

	/*
		//全部やるときは↓
		foreachStatic(func(val reflect.Value) {
			if v, ok := val.Interface().(entities.Resolvable); ok {
				v.ResolveRef()
			} else {
				revel.AppLog.Warnf("%s is not resolvable", val.String())
			}
		})
	*/

	// upsert
	foreachStatic(func(val reflect.Value) {
		tx.Save(reflect.Indirect(val).Addr().Interface())
	})

	// remove old resources
	for _, resource := range StaticTypes {
		for _, id := range WillRemove[resource] {
			sql := fmt.Sprintf("UPDATE %s SET updated_at = ?, deleted_at = ? WHERE id = ?", resource)
			tx.Exec(sql, time.Now(), time.Now(), id)
		}
	}

	tx.Commit()
}
