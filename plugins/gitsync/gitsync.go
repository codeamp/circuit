package gitsync

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

type GitSync struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("gitsync", func() transistor.Plugin {
		return &GitSync{}
	})
}

func (x *GitSync) Description() string {
	return "Sync Git repositories and create new features"
}

func (x *GitSync) SampleConfig() string {
	return ` `
}

func (x *GitSync) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started GitSync")

	return nil
}

func (x *GitSync) Stop() {
	log.Println("Stopping GitSync")
}

func (x *GitSync) Subscribe() []string {
	return []string{
		"plugins.GitPing",
		"plugins.GitSync:update",
	}
}

func (x *GitSync) Process(e transistor.Event) error {
	log.Info("Process GitSync event: %s", e.Name)
	var err error

	gitSyncEvent := e.Payload.(plugins.GitSync)
	gitSyncEvent.Action = plugins.Status
	gitSyncEvent.State = plugins.Fetching
	gitSyncEvent.StateMessage = ""

	commits, err := plugins.GitCommits(gitSyncEvent.From, gitSyncEvent.Project, gitSyncEvent.Git)
	if err != nil {
		log.Println(err)
		gitSyncEvent.State = plugins.Failed
		gitSyncEvent.StateMessage = fmt.Sprintf("%v (Action: %v)", err.Error(), gitSyncEvent.State)
		event := e.NewEvent(gitSyncEvent, err)
		x.events <- event
		return err
	}

	for i := range commits {
		c := commits[i]
		x.events <- e.NewEvent(c, nil)
	}

	return nil
}
