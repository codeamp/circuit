package main

import (
	"github.com/codeamp/circuit/cmd"
	_ "github.com/codeamp/circuit/plugins/codeamp"
	_ "github.com/codeamp/circuit/plugins/dockerbuilder"
	_ "github.com/codeamp/circuit/plugins/gitsync"
	_ "github.com/codeamp/circuit/plugins/heartbeat"
	_ "github.com/codeamp/circuit/plugins/kubernetes/deployments"
	_ "github.com/codeamp/circuit/plugins/kubernetes/loadbalancers"
)

func main() {
	cmd.Execute()
}
