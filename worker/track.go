package worker

import (
	"fmt"
	"time"
)

// TrackRun is a helper function that records a worker run and handles panics
func TrackRun(worker Worker, manager *Manager, fn func()) {
	startTime := time.Now()

	defer func() {
		if r := recover(); r != nil {
			if manager != nil {
				manager.RecordWorkerError(worker.Name(), fmt.Errorf("panic: %v", r))
			}
		}
		if manager != nil {
			manager.RecordWorkerRun(worker.Name())
		}
	}()

	fn()

	if manager != nil {
		manager.logger.Debugf("Worker %s completed in %v", worker.Name(), time.Since(startTime))
	}
}
