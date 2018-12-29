package services

import (
	"time"

	"github.com/revel/revel"
	"github.com/yasshi2525/RushHour/app/models"
)

var gamemaster *time.Ticker

func StartProcedure() {
	gamemaster = time.NewTicker(1 * time.Second)

	go proceed()
}

func proceed() {
	for range gamemaster.C {
		start := time.Now()

		// 経路探索中の場合、ゲームを進行しない
		models.MuRoute.Lock()

		models.MuStatic.Lock()

		models.MuAgent.Lock()

		time.Sleep(600 * time.Millisecond)

		models.MuAgent.Unlock()
		models.MuStatic.Unlock()

		models.MuRoute.Unlock()

		WarnLongExec(start, 2, "ゲーム進行", false)
	}
}

func StopProcedure() {
	if gamemaster != nil {
		revel.AppLog.Info("中止処理 開始")
		gamemaster.Stop()
		revel.AppLog.Info("中止処理 終了")
	}
}