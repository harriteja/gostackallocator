package adapter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// MetricsAdapter implements the MetricsClient interface using Prometheus
type MetricsAdapter struct {
	filesAnalyzed    prometheus.Counter
	issuesFound      prometheus.Counter
	analysisDuration prometheus.Histogram
	logger           *zap.Logger
}

// NewMetricsAdapter creates a new metrics adapter
func NewMetricsAdapter(logger *zap.Logger) *MetricsAdapter {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &MetricsAdapter{
		filesAnalyzed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "stackalloc_files_analyzed_total",
			Help: "The total number of files analyzed by stackalloc",
		}),
		issuesFound: promauto.NewCounter(prometheus.CounterOpts{
			Name: "stackalloc_issues_found_total",
			Help: "The total number of allocation issues found by stackalloc",
		}),
		analysisDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "stackalloc_analysis_duration_seconds",
			Help:    "Time spent analyzing files in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		logger: logger,
	}
}

// IncrementFilesAnalyzed increments the files analyzed counter
func (m *MetricsAdapter) IncrementFilesAnalyzed() {
	m.filesAnalyzed.Inc()
	m.logger.Debug("Incremented files analyzed counter")
}

// IncrementIssuesFound increments the issues found counter
func (m *MetricsAdapter) IncrementIssuesFound() {
	m.issuesFound.Inc()
	m.logger.Debug("Incremented issues found counter")
}

// RecordAnalysisDuration records the time spent analyzing
func (m *MetricsAdapter) RecordAnalysisDuration(duration float64) {
	m.analysisDuration.Observe(duration)
	m.logger.Debug("Recorded analysis duration", zap.Float64("duration_seconds", duration))
}

// NoOpMetricsAdapter provides a no-op implementation for when metrics are disabled
type NoOpMetricsAdapter struct{}

// NewNoOpMetricsAdapter creates a new no-op metrics adapter
func NewNoOpMetricsAdapter() *NoOpMetricsAdapter {
	return &NoOpMetricsAdapter{}
}

// IncrementFilesAnalyzed does nothing
func (m *NoOpMetricsAdapter) IncrementFilesAnalyzed() {}

// IncrementIssuesFound does nothing
func (m *NoOpMetricsAdapter) IncrementIssuesFound() {}

// RecordAnalysisDuration does nothing
func (m *NoOpMetricsAdapter) RecordAnalysisDuration(duration float64) {}
