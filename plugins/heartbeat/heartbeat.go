package heartbeat

import (
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/rk/go-cron"
)

type Heartbeat struct {
	events chan transistor.Event
}

func init() {
	log.Error("init from heartbeat")
	transistor.RegisterPlugin("heartbeat", func() transistor.Plugin {
		return &Heartbeat{}
	})
}

func (x *Heartbeat) Start(e chan transistor.Event) error {
	x.events = e

	cron.NewCronJob(cron.ANY, cron.ANY, cron.ANY, cron.ANY, cron.ANY, 0, func(time.Time) {
		event := transistor.NewEvent(plugins.GetEventName("heartbeat"), transistor.GetAction("status"), plugins.HeartBeat{Tick: "minute"})
		x.events <- event
	})

	cron.NewCronJob(cron.ANY, cron.ANY, cron.ANY, cron.ANY, 0, 0, func(time.Time) {
		event := transistor.NewEvent(plugins.GetEventName("heartbeat"), transistor.GetAction("status"), plugins.HeartBeat{Tick: "hour"})
		x.events <- event
	})

	log.Info("Started Heartbeat")
	return nil
}

func (x *Heartbeat) Stop() {
	log.Info("Stopping Heartbeat")
}

func (x *Heartbeat) Subscribe() []string {
	return []string{}
}

func (x *Heartbeat) Process(e transistor.Event) error {
	return nil
}
