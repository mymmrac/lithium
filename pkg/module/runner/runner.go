package runner

import "context"

// Runner represents something that can run services
type Runner interface {
	// Add service to runner
	Add(ctx context.Context, service Service)
	// RunAndWait run services and waits for execution to finish
	RunAndWait(ctx context.Context) error
}

// Service represents something that can be run and stop
type Service interface {
	// Run service
	Run(ctx context.Context) error
	// Stop service
	Stop()
}

// RunAndWait run services and waits for execution to finish (convenience function)
func RunAndWait(ctx context.Context, run Runner) error {
	return run.RunAndWait(ctx)
}

// AddServiceInvoker returns a function that adds a service to the runner (convenience function)
func AddServiceInvoker[S Service]() func(ctx context.Context, runner Runner, service S) {
	return func(ctx context.Context, runner Runner, service S) { runner.Add(ctx, service) }
}
