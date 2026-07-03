package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Worker metrics
	WorkerRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "corekit_worker_runs_total",
			Help: "Total number of worker runs",
		},
		[]string{"worker"},
	)

	WorkerErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "corekit_worker_errors_total",
			Help: "Total number of worker errors",
		},
		[]string{"worker"},
	)

	WorkerLastRunTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "corekit_worker_last_run_timestamp",
			Help: "Timestamp of last worker run",
		},
		[]string{"worker"},
	)
)

// Worker metrics functions
func IncrementWorkerRun(workerName string) {
	WorkerRunsTotal.WithLabelValues(workerName).Inc()
}

func IncrementWorkerError(workerName string) {
	WorkerErrorsTotal.WithLabelValues(workerName).Inc()
}

func SetWorkerLastRun(workerName string, timestamp int64) {
	WorkerLastRunTimestamp.WithLabelValues(workerName).Set(float64(timestamp))
}
