package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nguereza-tony/corekit/logger"
)

// WorkerStatus represents the current status of a worker
type WorkerStatus struct {
	Name        string
	Status      string // "healthy", "warning", "unhealthy", "stopped"
	LastRun     *time.Time
	LastError   string
	RunsCount   int64
	ErrorsCount int64
}

// Worker interface that all workers must implement
type Worker interface {
	Start(ctx context.Context) error
	Stop() error
	Name() string
}

// ManagedWorker wraps a worker with monitoring
type ManagedWorker struct {
	Worker
	Status    WorkerStatus
	mu        sync.RWMutex
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	manager   *Manager
	startTime time.Time
	ctx       context.Context
}

// Manager manages all background workers
type Manager struct {
	workers   map[string]*ManagedWorker
	logger    *logger.Logger
	mu        sync.RWMutex
	startTime time.Time
}

// NewManager creates a new worker manager
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		workers:   make(map[string]*ManagedWorker),
		logger:    log,
		startTime: time.Now(),
	}
}

// Register adds a worker to the manager
func (m *Manager) Register(worker Worker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workers[worker.Name()] = &ManagedWorker{
		Worker:  worker,
		manager: m,
		Status: WorkerStatus{
			Name:   worker.Name(),
			Status: "stopped",
		},
	}
	m.logger.Infof("Worker registered: %s", worker.Name())
}

// StartAll starts all registered workers
func (m *Manager) StartAll(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, managedWorker := range m.workers {
		managedWorker.start(ctx)
	}
}

// start starts a single managed worker
func (mw *ManagedWorker) start(parentCtx context.Context) {
	mw.mu.Lock()
	mw.Status.Status = "starting"
	mw.mu.Unlock()

	ctx, cancel := context.WithCancel(parentCtx)
	mw.cancel = cancel

	mw.wg.Add(1)
	go func() {
		defer mw.wg.Done()
		mw.run(ctx)
	}()
}

// run is the main loop for a managed worker
func (mw *ManagedWorker) run(ctx context.Context) {
	mw.manager.logger.Infof("Worker %s started", mw.Name())
	mw.updateStatus("healthy", nil)

	for {
		select {
		case <-ctx.Done():
			mw.updateStatus("stopped", nil)
			mw.manager.logger.Infof("Worker %s stopped", mw.Name())
			return
		default:
			mw.executeRun(ctx)
		}
	}
}

// executeRun executes a single worker run and handles retry logic
func (mw *ManagedWorker) executeRun(ctx context.Context) {
	startTime := time.Now()

	err := mw.Worker.Start(ctx)

	if err != nil && err != context.Canceled {
		mw.recordError(err)
		mw.manager.logger.Errorf("Worker %s failed: %v", mw.Name(), err)

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			mw.manager.logger.Infof("Restarting worker %s...", mw.Name())
		}
	} else if err == nil {
		mw.recordRun(startTime)

		mw.manager.logger.Warnf("Worker %s exited, restarting...", mw.Name())
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}
	}
}

// updateStatus updates worker status
func (mw *ManagedWorker) updateStatus(status string, err error) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.Status.Status = status
	if err != nil {
		mw.Status.LastError = err.Error()
	} else {
		mw.Status.LastError = ""
	}
}

// recordRun records a successful worker run
func (mw *ManagedWorker) recordRun(startTime time.Time) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.Status.RunsCount++
	mw.Status.Status = "healthy"
	now := time.Now()
	mw.Status.LastRun = &now
}

// recordError records a worker error
func (mw *ManagedWorker) recordError(err error) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.Status.ErrorsCount++
	mw.Status.LastError = err.Error()

	if mw.Status.ErrorsCount > 3 {
		mw.Status.Status = "unhealthy"
	} else if mw.Status.ErrorsCount > 0 {
		mw.Status.Status = "warning"
	}
}

// RecordWorkerRun records a successful data processing cycle
func (m *Manager) RecordWorkerRun(workerName string) {
	m.mu.RLock()
	mw, exists := m.workers[workerName]
	m.mu.RUnlock()

	if exists {
		mw.mu.Lock()
		defer mw.mu.Unlock()
		mw.Status.RunsCount++
		mw.Status.Status = "healthy"
		now := time.Now()
		mw.Status.LastRun = &now

		IncrementWorkerRun(workerName)
		SetWorkerLastRun(workerName, now.Unix())
	}
}

// RecordWorkerError records an error for a worker
func (m *Manager) RecordWorkerError(workerName string, err error) {
	m.mu.RLock()
	mw, exists := m.workers[workerName]
	m.mu.RUnlock()

	if exists {
		mw.mu.Lock()
		defer mw.mu.Unlock()
		mw.Status.ErrorsCount++
		mw.Status.LastError = err.Error()
		if mw.Status.ErrorsCount > 3 {
			mw.Status.Status = "unhealthy"
		} else {
			mw.Status.Status = "warning"
		}

		IncrementWorkerError(workerName)
	}
}

// StopAll stops all workers gracefully
func (m *Manager) StopAll() {
	m.logger.Info("Stopping all workers...")
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, mw := range m.workers {
		if mw.cancel != nil {
			mw.cancel()
		}
		if err := mw.Worker.Stop(); err != nil {
			m.logger.Errorf("Failed to stop worker %s: %v", mw.Name(), err)
		}
		mw.wg.Wait()
	}
	m.logger.Info("All workers stopped")
}

// WaitForShutdown waits for interrupt signal and stops workers
func (m *Manager) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	m.logger.Info("Shutdown signal received")
	m.StopAll()
}

// GetWorkersStatus returns status of all workers
func (m *Manager) GetWorkersStatus() map[string]WorkerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]WorkerStatus)
	for name, mw := range m.workers {
		mw.mu.RLock()
		result[name] = mw.Status
		mw.mu.RUnlock()
	}
	return result
}

// GetWorkerStatus returns status of a specific worker
func (m *Manager) GetWorkerStatus(name string) *WorkerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if mw, exists := m.workers[name]; exists {
		mw.mu.RLock()
		defer mw.mu.RUnlock()
		status := mw.Status
		return &status
	}
	return nil
}

// GetUptime returns manager uptime
func (m *Manager) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// RestartWorker restarts a worker by name
func (m *Manager) RestartWorker(ctx context.Context, workerName string, accountID uuid.UUID, ipAddress, userAgent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mw, exists := m.workers[workerName]
	if !exists {
		return fmt.Errorf("worker %s not found", workerName)
	}

	if mw.cancel != nil {
		mw.cancel()
	}
	if err := mw.Stop(); err != nil {
		m.logger.Errorf("Failed to stop worker %s: %v", workerName, err)
	}

	time.Sleep(500 * time.Millisecond)

	mw.mu.Lock()
	mw.Status.Status = "starting"
	mw.Status.LastError = ""
	mw.Status.RunsCount = 0
	mw.Status.ErrorsCount = 0
	mw.Status.LastRun = nil
	mw.mu.Unlock()

	newCtx := context.Background()
	m.startManagedWorker(newCtx, mw)

	m.logger.Infof("Worker %s restarted successfully", workerName)
	return nil
}

// startManagedWorker starts a single worker with monitoring
func (m *Manager) startManagedWorker(ctx context.Context, mw *ManagedWorker) {
	mw.startTime = time.Now()
	mw.ctx, mw.cancel = context.WithCancel(ctx)

	go func() {
		m.logger.Infof("Starting worker: %s", mw.Name())
		mw.updateStatus("healthy", nil)

		for {
			select {
			case <-mw.ctx.Done():
				mw.updateStatus("stopped", nil)
				m.logger.Infof("Worker %s stopped", mw.Name())
				return
			default:
				startTime := time.Now()
				err := mw.Start(mw.ctx)

				if err != nil && err != context.Canceled {
					mw.recordError(err)
					m.logger.Errorf("Worker %s failed: %v", mw.Name(), err)

					select {
					case <-mw.ctx.Done():
						return
					case <-time.After(5 * time.Second):
						m.logger.Infof("Restarting worker %s...", mw.Name())
					}
				} else {
					mw.recordRun(startTime)

					if err == nil {
						m.logger.Warnf("Worker %s exited, restarting...", mw.Name())
						select {
						case <-mw.ctx.Done():
							return
						case <-time.After(1 * time.Second):
						}
					}
				}
			}
		}
	}()
}
