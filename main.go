package main

import (
	"github.com/codeamp/circuit/cmd"
	_ "github.com/codeamp/circuit/plugins/codeamp"
	_ "github.com/codeamp/circuit/plugins/git_sync"
	_ "github.com/codeamp/circuit/plugins/heartbeat"
)

func main() {
	cmd.Execute()
}
