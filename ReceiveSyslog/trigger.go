package ReceiveSyslog

import (
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

// Create a new logger
var log = logger.GetLogger("trigger-receiveSyslog")

type SyslogTrigger struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	server   *ServerTCP
}

// NewFactory create a new Trigger factory
func NewFactory(md *trigger.Metadata) trigger.Factory {
	return &SyslogFactory{metadata: md}
}

// SyslogFactory Syslog Trigger factory
type SyslogFactory struct {
	metadata *trigger.Metadata
}

// New Creates a new trigger instance for a given id
func (t *SyslogFactory) New(config *trigger.Config) trigger.Trigger {
	return &SyslogTrigger{metadata: t.metadata, config: config, server: nil}
}

// Metadata implements trigger.Trigger.Metadata
func (t *SyslogTrigger) Metadata() *trigger.Metadata {
	return t.metadata
}

func (t *SyslogTrigger) Initialize(ctx trigger.InitContext) error {
	if t.config.Settings == nil {
		return fmt.Errorf("no Settings found for trigger '%s'", t.config.Id)
	}

	port, ok := t.config.Settings["port"];
	if!ok {
		return fmt.Errorf("no Port found for trigger '%s' in settings", t.config.Id)
	}

	regexMap := make(map[string]*trigger.Handler)

	// Init handlers
	for _, handler := range ctx.GetHandlers() {

		regex :=handler.GetStringSetting("regex")

		log.Debugf("Registering handler [%s]", regex)

		regexMap[regex] = handler
	}

	log.Debugf("Configured on port %s", t.config.Settings["port"])
	t.server = &ServerTCP{ port: int(port.(float64)), m: regexMap }

	return nil
}

func (t *SyslogTrigger) Start() error {
	return t.server.Start()
}

// Stop implements util.Managed.Stop
func (t *SyslogTrigger) Stop() error {
	return t.server.Stop()
}
