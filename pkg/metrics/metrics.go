package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	namespace = "run_puppet"
	jobName   = "run-puppet"
)

// Metrics collects timings of a single run-puppet invocation.
// They are pushed to a Prometheus pushgateway once at the end of the run.
// All methods are safe to call on a nil *Metrics.
type Metrics struct {
	registry *prometheus.Registry

	runDuration    prometheus.Gauge
	cloneDuration  prometheus.Gauge
	puppetDuration prometheus.Gauge
	exitCode       prometheus.Gauge
	lastRun        prometheus.Gauge
}

func newGauge(name string, help string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{Namespace: namespace, Name: name, Help: help})
}

func New() *Metrics {
	m := &Metrics{
		registry:       prometheus.NewRegistry(),
		runDuration:    newGauge("run_duration_seconds", "Full runtime of run-puppet, including the random delay."),
		cloneDuration:  newGauge("git_clone_duration_seconds", "Time spent cloning the puppet repository."),
		puppetDuration: newGauge("puppet_apply_duration_seconds", "Time spent running puppet apply."),
		exitCode:       newGauge("exit_code", "Exit code of the run."),
		lastRun:        newGauge("last_run_timestamp_seconds", "Unix timestamp of the end of the last run."),
	}

	// clone and puppet durations are registered when observed, so skipped steps are not reported as 0
	m.registry.MustRegister(m.runDuration, m.exitCode, m.lastRun)

	return m
}

// setDuration registers g if needed and sets it to the seconds elapsed since start.
func (m *Metrics) setDuration(g prometheus.Gauge, start time.Time) {
	_ = m.registry.Register(g) // only fails with AlreadyRegisteredError for our fixed gauges
	g.Set(time.Since(start).Seconds())
}

func (m *Metrics) ObserveRun(start time.Time) {
	if m == nil {
		return
	}
	m.setDuration(m.runDuration, start)
}

func (m *Metrics) ObserveClone(start time.Time) {
	if m == nil {
		return
	}
	m.setDuration(m.cloneDuration, start)
}

func (m *Metrics) ObservePuppet(start time.Time) {
	if m == nil {
		return
	}
	m.setDuration(m.puppetDuration, start)
}

func (m *Metrics) SetExitCode(code int) {
	if m == nil {
		return
	}
	m.exitCode.Set(float64(code))
}

// Push replaces all metrics of this job/instance group on the pushgateway.
func (m *Metrics) Push(ctx context.Context, url string, instance string) error {
	if m == nil {
		return nil
	}
	m.lastRun.SetToCurrentTime()

	return push.New(url, jobName).
		Gatherer(m.registry).
		Grouping("instance", instance).
		PushContext(ctx)
}
