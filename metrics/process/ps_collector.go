package process

import (
	"fmt"
	"github.com/bloxapp/ssv/metrics"
	"runtime"
	"sort"
)

const (
	collectorID = "process"
	// metrics:
	completedGCCycles = "completed_gc_cycles"
	goVersion         = "go_version"
	cpusCount         = "cpus_count"
	goroutinesCount   = "goroutines_count"
	memoryStats       = "memory_stats"
)

// SetupProcessMetrics initialize collector for process metrics
func SetupProcessMetrics() {
	c := processCollector{}
	metrics.Register(&c)
}

// validatorsCollector implements metrics.Collector for validators information
type processCollector struct {
}

func (c *processCollector) ID() string {
	return collectorID
}

func (c *processCollector) Collect() ([]string, error) {
	var results []string

	results = append(results, fmt.Sprintf("%s{} %s", goVersion, runtime.Version()))
	results = append(results, fmt.Sprintf("%s{} %d", cpusCount, runtime.NumCPU()))
	results = append(results, fmt.Sprintf("%s{} %d", goroutinesCount, runtime.NumGoroutine()))
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	results = append(results, fmt.Sprintf("%s{alloc=\"%f\",sys=\"%f\"} 1", memoryStats,
		float64(ms.Alloc)/1000000.0, float64(ms.Sys)/1000000.0))
	results = append(results, fmt.Sprintf("%s{} %d", completedGCCycles, ms.NumGC))

	sort.Strings(results)

	return results, nil
}
