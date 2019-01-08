package transistor

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/codeamp/logger"
	workers "github.com/jrallison/go-workers"
	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
)

// Transistor runs codeflow and collects datt based on the given config
type Transistor struct {
	Config     Config
	Events     chan Event
	TestEvents chan Event
	Shutdown   chan struct{}
	Plugins    []*RunningPlugin
}

type Config struct {
	Server         string
	Password       string
	Database       string
	Namespace      string
	Pool           string
	Process        string
	Plugins        map[string]interface{}
	EnabledPlugins []string
	Queueing       bool
}

// NewTransistor returns an Transistor struct based off the given Config
func NewTransistor(config Config) (*Transistor, error) {
	if len(config.Plugins) == 0 {
		log.Fatal("No plugins found, did you provide valid config file?")
	}

	transistor := &Transistor{
		Config: config,
	}

	// channel shared between all plugin threads for accumulating events
	transistor.Events = make(chan Event, 10000)

	// channel shared between all plugin threads to trigger shutdown
	transistor.Shutdown = make(chan struct{})

	if err := transistor.LoadPlugins(); err != nil {
		log.Fatal(err)
	}

	return transistor, nil
}

// NewTestTransistor returns an Transistor struct based off the given Config
func NewTestTransistor(config Config) (*Transistor, error) {
	var err error
	var transistor *Transistor

	if transistor, err = NewTransistor(config); err != nil {
		log.FatalWithFields("Failed initializing transistor", log.Fields{
			"error": err,
		})
	}

	transistor.TestEvents = make(chan Event, 10000)
	transistor.Config.Queueing = false

	return transistor, nil
}

func (t *Transistor) LoadPlugins() error {
	var err error
	for name := range t.Config.Plugins {
		log.Warn("Adding plugin: ", name)
		if err = t.addPlugin(name); err != nil {
			return fmt.Errorf("Error parsing %s, %s", name, err)
		}
		if len(t.Config.EnabledPlugins) == 0 || SliceContains(name, t.Config.EnabledPlugins) {
			if err = t.enablePlugin(name); err != nil {
				return fmt.Errorf("Error parsing %s, %s", name, err)
			}
		}
	}

	return nil
}

// Returns t list of strings of the configured plugins.
func (t *Transistor) PluginNames() []string {
	var name []string
	for key, _ := range t.Config.Plugins {
		name = append(name, key)
	}
	return name
}

func (t *Transistor) addPlugin(name string) error {
	if len(t.PluginNames()) > 0 && !SliceContains(name, t.PluginNames()) {
		return nil
	}

	creator, ok := PluginRegistry[name]
	if !ok {
		return fmt.Errorf("Undefined but requested plugin: %s", name)
	}

	plugin := creator()

	if err := mapstructure.Decode(t.Config.Plugins[name], plugin); err != nil {
		log.Fatal(err)
	}

	work := func(message *workers.Msg) {
		e, _ := json.Marshal(message.Args())
		event := Event{}
		json.Unmarshal([]byte(e), &event)
		if err := MapPayload(event.PayloadModel, &event); err != nil {
			log.Fatal(fmt.Errorf("PayloadModel not found: %s. Did you add it to ApiRegistry?", event.PayloadModel))
		}

		defer func() {
			if r := recover(); r != nil {
				log.Error(fmt.Sprintf("%s: %s", r, debug.Stack()))
			}
		}()

		plugin.Process(event)
	}

	wc := t.Config.Plugins[name].(map[string]interface{})
	workersCount := 0
	if wc["workers"] != nil {
		workersCount = wc["workers"].(int)
	}

	wrc := t.Config.Plugins[name].(map[string]interface{})
	workerRetriesCount := 0
	if wrc["worker_retries"] != nil {
		workerRetriesCount = wrc["worker_retries"].(int)
	}

	rp := &RunningPlugin{
		Name:          name,
		Plugin:        plugin,
		Work:          work,
		Enabled:       false,
		Workers:       workersCount,
		WorkerRetries: workerRetriesCount,
	}

	t.Plugins = append(t.Plugins, rp)

	return nil
}

func (t *Transistor) enablePlugin(name string) error {
	if len(t.PluginNames()) > 0 && !SliceContains(name, t.PluginNames()) {
		return nil
	}

	for _, rp := range t.Plugins {
		if rp.Name == name {
			rp.Enabled = true
		}
	}

	return nil
}

// flusher monitors the events plugin channel and schedules them to correct queues
func (t *Transistor) flusher() {
	for {
		select {
		case <-t.Shutdown:
			log.Info("Hang on, flushing any cached metrics before shutdown")
			return
		case e := <-t.Events:
			ev_handled := false

			for _, plugin := range t.Plugins {
				if plugin.Workers > 0 {
					subscribedTo := plugin.Plugin.Subscribe()
					if SliceContains(e.Event(), subscribedTo) {
						ev_handled = true
						if t.Config.Queueing {
							log.DebugWithFields("Enqueue event", log.Fields{
								"event_name":  e.Event(),
								"plugin_name": plugin.Name,
							})

							options := workers.EnqueueOptions{Retry: false}
							if plugin.WorkerRetries > 0 {
								options.Retry = true
								options.RetryCount = plugin.WorkerRetries
							}

							workers.EnqueueWithOptions(plugin.Name, "Event", e, options)
						} else {
							log.Warn("GoFunc Process")
							go func() {
								plugin.Plugin.Process(e)
							}()
						}
					}
				}
			}

			if t.TestEvents != nil {
				t.TestEvents <- e
			} else if !ev_handled {
				log.WarnWithFields("Event not handled by any plugin", log.Fields{
					"event_name": e.Name,
				})
				// e.Dump()
			}
		}
	}
}

// Run runs the transistor daemon
func (t *Transistor) Run() error {
	var wg sync.WaitGroup

	if t.Config.Queueing {
		workers.Middleware = workers.NewMiddleware(
			&workers.MiddlewareRetry{},
		)

		workers.Logger = log.Instance()

		processName := t.Config.Process
		if processName == "" {
			processName = uuid.NewV4().String()
		}

		workers.Configure(map[string]string{
			"server":    t.Config.Server,
			"password":  t.Config.Password,
			"namespace": t.Config.Namespace,
			"database":  t.Config.Database,
			"pool":      t.Config.Pool,
			"process":   processName,
		})
	}

	defer t.stopPlugins()
	for _, plugin := range t.Plugins {
		if !plugin.Enabled {
			continue
		}

		// Start service of any Plugins
		switch p := plugin.Plugin.(type) {
		case Plugin:
			if plugin.Workers > 0 {
				log.Debug(fmt.Sprintf("Starting plugin: %s", plugin.Name))
				if err := p.Start(t.Events); err != nil {
					log.InfoWithFields("Service failed to start", log.Fields{
						"plugin_name": plugin.Name,
						"error":       err,
					})
					p.Stop()
					return err
				}
				plugin.Started = true

				if t.Config.Queueing {
					workers.Process(plugin.Name, plugin.Work, plugin.Workers)
				}
			} else {
				log.Warn(fmt.Sprintf("Plugin '%s' specified, but 0 workers requested", plugin.Name))
			}
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		t.flusher()
	}()

	if t.Config.Queueing {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workers.Run()
			t.Stop()
		}()

		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	workers.StatsServer(8080)
		//}()
	}

	wg.Wait()
	return nil
}

func (t *Transistor) stopPlugins() {
	for _, plugin := range t.Plugins {
		if !plugin.Enabled || plugin.Started == false {
			log.Warn("Skipping shutdown of plugin. Not enabled or not started: ", plugin.Name)
			continue
		}

		switch p := plugin.Plugin.(type) {
		case Plugin:
			log.Debug(fmt.Sprintf("Stopping Plugin: %s", plugin.Name))
			p.Stop()
			plugin.Started = false
		}
	}
}

// Shutdown the transistor daemon
func (t *Transistor) Stop() {
	log.Warn("Stopping 'shutdown' channel")
	close(t.Shutdown)
}

type testEventResponse struct {
	Event
	Error error
}

func (t *Transistor) GetTestEvent(name EventName, action Action, timeout time.Duration) (Event, error) {
	eventName := fmt.Sprintf("%s:%s", name, action)

	// timeout in the case that we don't get requested event
	timer := time.NewTimer(time.Second * timeout)
	defer timer.Stop()

	responseChan := make(chan (testEventResponse))

	go func() {
		for {
			select {
			case e := <-t.TestEvents:
				matched, err := regexp.MatchString(eventName, e.Event())
				if err != nil {
					log.ErrorWithFields("GetTestEvent regex match encountered an error", log.Fields{
						"regex":  name,
						"string": e.Name,
						"error":  err,
					})

					responseChan <- testEventResponse{Event{}, err}
					return
				}

				if matched {
					responseChan <- testEventResponse{e, nil}
					return
				}

				//log.Debug(fmt.Printf("TestEvent received but not matched. Found '%s', looking for '%s'", e.Event(), eventName))
			case <-timer.C:
				responseChan <- testEventResponse{Event{}, fmt.Errorf("Timer expired while waiting for test event (%s)", time.Second*timeout)}
				return
			default:
				time.Sleep(time.Millisecond * 50)
			}
		}
	}()

	result := <-responseChan
	return result.Event, result.Error
}
