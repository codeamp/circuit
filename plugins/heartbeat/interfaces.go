package heartbeat

import (
	"time"

	gocron "github.com/rk/go-cron"
)

type Croner interface {
	NewCronJob(month, day, weekday, hour, minute, second int8, task func(time.Time))
}

type LegitimateCron struct{}

func (c LegitimateCron) NewCronJob(month, day, weekday, hour, minute, second int8, task func(time.Time)) {
	gocron.NewCronJob(month, day, weekday, hour, minute, second, task)
}
