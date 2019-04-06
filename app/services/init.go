package services

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
)

var db *gorm.DB

// Init prepares for starting game
func Init() {
	revel.AppLog.Info("start preparation for game")
	defer revel.AppLog.Info("end preparation for game")

	start := time.Now()
	InitLock()
	LoadConf()
	defer WarnLongExec(start, start, Const.Perf.Init.D, "initialization", true)
	InitRepository()
	db = connectDB()
	//db.LogMode(true)
	MigrateDB()
	Restore()
	StartRouting()
}

// Terminate finalizes after stopping game
func Terminate() {
	if db != nil {
		closeDB()
	}
}

// Start start game
func Start() {
	StartBackupTicker()
	StartModelWatching()
	StartProcedure()
}

// Stop stop game
func Stop() {
	CancelRouting()
	StopProcedure()
	StopModelWatching()
	StopBackupTicker()
	Backup()
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

	for i := 1; i <= 60; i++ {
		database, err = gorm.Open(driver, spec)
		if err != nil {
			revel.AppLog.Warnf("failed to connect database(%v). retry after 10 seconds.", err)
			time.Sleep(10 * time.Second)
		}
	}

	if err != nil {
		panic(fmt.Errorf("failed to connect database: %v", err))
	}

	revel.AppLog.Info("connect database successfully")
	return database
}

func closeDB() {
	if err := db.Close(); err != nil {
		panic(err)
	}
	revel.AppLog.Info("disconnect database successfully")
}
