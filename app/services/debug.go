package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/revel/revel"
)

var (
	numUser        = 10
	viewInterval   = 1 * time.Second
	updateInterval = 30 * time.Second
	backupInterval = 30 * time.Second
)

type opCallback func(source string, target string)

// Main immitates some user requests some actions.
// TODO remove it
func Main() {
	Restore()
	StartModelWatching()
	StartProcedure()

	for i := 0; i < numUser; i++ {
		source := fmt.Sprintf("user%d", i)
		tickOp(source, "dummy", viewInterval, func(src string, tar string) { ViewMap() })
		for _, target := range []string{"residence", "company"} {
			tickOp(source, target, updateInterval, func(src string, tar string) {
				revel.AppLog.Infof("%s create %s", src, tar)
				UpdateModel(&Operation{
					Source: src,
					Op:     "create",
					Target: tar,
					X:      rand.Float64() * 100,
					Y:      rand.Float64() * 100,
				})
			})
		}
		for _, target := range []string{"residence", "company"} {
			tickOp(source, target, updateInterval, func(src string, tar string) {
				revel.AppLog.Infof("%s remove %s", src, tar)
				UpdateModel(&Operation{
					Source: src,
					Op:     "remove",
					Target: tar,
				})
			})
		}
	}

	var backup = time.NewTicker(backupInterval)
	go func() {
		for range backup.C {
			Backup()
		}
	}()
}

func tickOp(source string, target string, interval time.Duration, callback opCallback) {
	go func() {
		sleep := rand.Intn(int(interval.Seconds()))
		time.Sleep(time.Duration(sleep) * time.Second)
		t := time.NewTicker(interval)
		for range t.C {
			callback(source, target)
		}
	}()
}

// WarnLongExec alerts long time consuming task.
func WarnLongExec(start time.Time, max float64, title string, verbose bool) {
	if consumed := time.Now().Sub(start).Seconds(); consumed > max {
		revel.AppLog.Warnf("%s に %.1f sec 消費", title, consumed)
	} else if verbose {
		revel.AppLog.Debugf("%s に %.1f sec 消費", title, consumed)
	}
}
