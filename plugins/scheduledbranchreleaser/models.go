package scheduledbranchreleaser

import "github.com/codeamp/transistor"

type ScheduledBranchReleaser struct {
	events chan transistor.Event
}
