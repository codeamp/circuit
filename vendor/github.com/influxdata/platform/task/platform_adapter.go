package task

import (
	"context"
	"errors"
	"time"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/task/backend"
	"github.com/influxdata/platform/task/options"
)

type RunController interface {
	CancelRun(ctx context.Context, taskID, runID platform.ID) error
	//TODO: add retry run to this.
}

// PlatformAdapter wraps a task.Store into the platform.TaskService interface.
func PlatformAdapter(s backend.Store, r backend.LogReader, rc RunController) platform.TaskService {
	return pAdapter{s: s, r: r}
}

type pAdapter struct {
	s  backend.Store
	rc RunController
	r  backend.LogReader
}

var _ platform.TaskService = pAdapter{}

func (p pAdapter) FindTaskByID(ctx context.Context, id platform.ID) (*platform.Task, error) {
	t, m, err := p.s.FindTaskByIDWithMeta(ctx, id)
	if err != nil {
		return nil, err
	}

	// The store interface specifies that a returned task is nil if the operation succeeded without a match.
	if t == nil {
		return nil, nil
	}

	return toPlatformTask(*t, m)
}

func (p pAdapter) FindTasks(ctx context.Context, filter platform.TaskFilter) ([]*platform.Task, int, error) {
	const pageSize = 100 // According to the platform.TaskService.FindTasks API.

	params := backend.TaskSearchParams{PageSize: pageSize}
	if filter.Organization != nil {
		params.Org = *filter.Organization
	}
	if filter.User != nil {
		params.User = *filter.User
	}
	if filter.After != nil {
		params.After = *filter.After
	}
	ts, err := p.s.ListTasks(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	pts := make([]*platform.Task, len(ts))
	for i, t := range ts {
		pts[i], err = toPlatformTask(t.Task, &t.Meta)
		if err != nil {
			return nil, 0, err
		}
	}

	totalResults := len(pts) // TODO(mr): don't lie about the total results. Update ListTasks signature?
	return pts, totalResults, nil
}

func (p pAdapter) CreateTask(ctx context.Context, t *platform.Task) error {
	opts, err := options.FromScript(t.Flux)
	if err != nil {
		return err
	}

	// TODO(mr): decide whether we allow user to configure scheduleAfter. https://github.com/influxdata/platform/issues/595
	scheduleAfter := time.Now().Unix()
	req := backend.CreateTaskRequest{Org: t.Organization, User: t.Owner.ID, Script: t.Flux, ScheduleAfter: scheduleAfter}
	id, err := p.s.CreateTask(ctx, req)
	if err != nil {
		return err
	}
	t.ID = id
	t.Every = opts.Every.String()
	t.Cron = opts.Cron

	return nil
}

func (p pAdapter) UpdateTask(ctx context.Context, id platform.ID, upd platform.TaskUpdate) (*platform.Task, error) {
	if upd.Flux == nil && upd.Status == nil {
		return nil, errors.New("cannot update task without content")
	}

	req := backend.UpdateTaskRequest{ID: id}
	if upd.Flux != nil {
		req.Script = *upd.Flux
	}
	if upd.Status != nil {
		req.Status = backend.TaskStatus(*upd.Status)
	}
	res, err := p.s.UpdateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	opts, err := options.FromScript(res.NewTask.Script)
	if err != nil {
		return nil, err
	}

	task := &platform.Task{
		ID:     id,
		Name:   opts.Name,
		Status: res.NewMeta.Status,
		Owner:  platform.User{},
		Flux:   res.NewTask.Script,
		Every:  opts.Every.String(),
		Cron:   opts.Cron,
	}

	t, err := p.s.FindTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	task.Owner.ID = t.User
	task.Organization = t.Org

	return task, nil
}

func (p pAdapter) DeleteTask(ctx context.Context, id platform.ID) error {
	_, err := p.s.DeleteTask(ctx, id)
	// TODO(mr): Store.DeleteTask returns false, nil if ID didn't match; do we want to handle that case?
	return err
}

func (p pAdapter) FindLogs(ctx context.Context, filter platform.LogFilter) ([]*platform.Log, int, error) {
	logs, err := p.r.ListLogs(ctx, filter)
	logPointers := make([]*platform.Log, len(logs))
	for i := range logs {
		logPointers[i] = &logs[i]
	}
	return logPointers, len(logs), err
}

func (p pAdapter) FindRuns(ctx context.Context, filter platform.RunFilter) ([]*platform.Run, int, error) {
	runs, err := p.r.ListRuns(ctx, filter)
	return runs, len(runs), err
}

func (p pAdapter) FindRunByID(ctx context.Context, taskID, id platform.ID) (*platform.Run, error) {
	task, err := p.s.FindTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return p.r.FindRunByID(ctx, task.Org, id)
}

func (p pAdapter) RetryRun(ctx context.Context, taskID, id platform.ID, requestedAt int64) error {
	task, err := p.s.FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	run, err := p.r.FindRunByID(ctx, task.Org, id)
	if err != nil {
		return err
	}
	if run.Status == backend.RunStarted.String() {
		return backend.ErrRunNotFinished
	}

	scheduledTime, err := time.Parse(time.RFC3339, run.ScheduledFor)
	if err != nil {
		return err
	}
	t := scheduledTime.UTC().Unix()

	return p.s.ManuallyRunTimeRange(ctx, run.TaskID, t, t, requestedAt)
}

func (p pAdapter) CancelRun(ctx context.Context, taskID, runID platform.ID) error {
	return p.rc.CancelRun(ctx, taskID, runID)
}

func toPlatformTask(t backend.StoreTask, m *backend.StoreTaskMeta) (*platform.Task, error) {
	opts, err := options.FromScript(t.Script)
	if err != nil {
		return nil, err
	}

	pt := &platform.Task{
		ID:           t.ID,
		Organization: t.Org,
		Name:         t.Name,
		Owner: platform.User{
			ID:   t.User,
			Name: "", // TODO(mr): how to get owner name?
		},
		Flux: t.Script,
		Cron: opts.Cron,
	}
	if opts.Every != 0 {
		pt.Every = opts.Every.String()
	}
	if m != nil {
		pt.Status = string(m.Status)
	}
	return pt, nil
}
