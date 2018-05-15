package main

import (
	"os"

	"github.com/codeamp/circuit/cmd"
	_ "github.com/codeamp/circuit/plugins/codeamp"
	_ "github.com/codeamp/circuit/plugins/dockerbuilder"
	_ "github.com/codeamp/circuit/plugins/githubstatus"
	_ "github.com/codeamp/circuit/plugins/gitsync"
	_ "github.com/codeamp/circuit/plugins/heartbeat"
	_ "github.com/codeamp/circuit/plugins/kubernetes"
	_ "github.com/codeamp/circuit/plugins/route53"
	log "github.com/codeamp/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	if _logLevel := os.Getenv("LOG_LEVEL"); _logLevel != "" {
		logLevel, err := logrus.ParseLevel(_logLevel)

		if err != nil {
			log.Fatal(err)
		}

		log.SetLogLevel(logLevel)
	}

	cmd.Execute()
}
