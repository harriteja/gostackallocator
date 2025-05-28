package internal

import (
	"time"

	"go.uber.org/zap"
)

// AnalysisMetrics holds metrics for analysis operations
type AnalysisMetrics struct {
	FilesAnalyzed    int64
	IssuesFound      int64
	AnalysisDuration time.Duration
	StartTime        time.Time
	logger           *zap.Logger
}

// NewAnalysisMetrics creates a new metrics instance
func NewAnalysisMetrics(logger *zap.Logger) *AnalysisMetrics {
	if logger == nil {
		logger = GetLogger()
	}

	return &AnalysisMetrics{
		StartTime: time.Now(),
		logger:    logger,
	}
}

// IncrementFilesAnalyzed increments the files analyzed counter
func (m *AnalysisMetrics) IncrementFilesAnalyzed() {
	m.FilesAnalyzed++
	m.logger.Debug("Files analyzed incremented", zap.Int64("count", m.FilesAnalyzed))
}

// IncrementIssuesFound increments the issues found counter
func (m *AnalysisMetrics) IncrementIssuesFound() {
	m.IssuesFound++
	m.logger.Debug("Issues found incremented", zap.Int64("count", m.IssuesFound))
}

// RecordAnalysisDuration records the analysis duration
func (m *AnalysisMetrics) RecordAnalysisDuration(duration time.Duration) {
	m.AnalysisDuration = duration
	m.logger.Debug("Analysis duration recorded", zap.Duration("duration", duration))
}

// Finish finalizes the metrics and logs summary
func (m *AnalysisMetrics) Finish() {
	totalDuration := time.Since(m.StartTime)
	m.logger.Info("Analysis completed",
		zap.Int64("files_analyzed", m.FilesAnalyzed),
		zap.Int64("issues_found", m.IssuesFound),
		zap.Duration("total_duration", totalDuration),
		zap.Duration("analysis_duration", m.AnalysisDuration),
	)
}

// GetSummary returns a summary of the metrics
func (m *AnalysisMetrics) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"files_analyzed":    m.FilesAnalyzed,
		"issues_found":      m.IssuesFound,
		"analysis_duration": m.AnalysisDuration.String(),
		"total_duration":    time.Since(m.StartTime).String(),
	}
}

// Timer is a simple timer for measuring durations
type Timer struct {
	start  time.Time
	name   string
	logger *zap.Logger
}

// NewTimer creates a new timer
func NewTimer(name string, logger *zap.Logger) *Timer {
	if logger == nil {
		logger = GetLogger()
	}

	return &Timer{
		start:  time.Now(),
		name:   name,
		logger: logger,
	}
}

// Stop stops the timer and logs the duration
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.logger.Debug("Timer stopped",
		zap.String("name", t.name),
		zap.Duration("duration", duration),
	)
	return duration
}

// Elapsed returns the elapsed time without stopping the timer
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}
