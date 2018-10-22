package heartbeat_test

import (
	"time"

	log "github.com/codeamp/logger"
)

type MockedCron struct {
}

func (c MockedCron) NewCronJob(month, day, weekday, hour, minute, second int8, task func(time.Time)) {
	log.Debug("Mocked Cron Response - Firing Immediately")
	task(time.Now())
}
