package scheduledbranchreleaser

import "github.com/codeamp/transistor"

type ScheduledBranchReleaser struct {
	Events chan transistor.Event
}
