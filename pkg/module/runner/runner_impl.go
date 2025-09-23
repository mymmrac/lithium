package runner

import (
	"context"
	"sync"

	"github.com/mymmrac/lithium/pkg/module/di"
)

func init() { //nolint:gochecknoinits
	di.Base().MustProvide(NewRunner)
}

type runner struct {
	sync.Mutex
	wg       sync.WaitGroup
	services []Service
	running  bool
	err      error
}

// NewRunner creates new runner.
func NewRunner(cancel context.CancelFunc) Runner {
	return &runner{
		services: []Service{&contextWatcher{cancel: cancel}},
	}
}

// Add service to runner.
func (a *runner) Add(ctx context.Context, service Service) {
	a.Lock()
	if !a.running {
		a.services = append(a.services, service)
	} else {
		a.wg.Add(1)
		go a.run(ctx, service)
	}
	a.Unlock()
}

// RunAndWait run services and waits for execution to finish.
func (a *runner) RunAndWait(ctx context.Context) error {
	a.Lock()
	if !a.running {
		a.running = true
		for _, service := range a.services {
			a.wg.Add(1)
			go a.run(ctx, service)
		}
	}
	a.Unlock()
	a.wg.Wait()
	return a.err
}

func (a *runner) run(ctx context.Context, service Service) {
	svcErr := service.Run(ctx)
	a.Lock()
	if a.err == nil {
		a.err = svcErr
		for _, otherSvc := range a.services {
			otherSvc.Stop()
		}
		a.services = nil
	}
	a.Unlock()
	a.wg.Done()
}
