package heartbeat

import (
	"time"
	gocron "github.com/rk/go-cron"
)

type cron interface {
	NewCronJob(month, day, weekday, hour, minute, second int8, task func(time.Time))
}

type Cron struct{}

func (c Cron) NewCronJob(month, day, weekday, hour, minute, second int8, task func(time.Time)) {
	gocron.NewCronJob(month, day, weekday, hour, minute, second, task)
}
