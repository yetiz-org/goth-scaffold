package daemons

import (
	"os"
	"runtime"

	"github.com/pbnjay/memory"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
)

type SchedulePerformanceMeasure struct {
	kkdaemon.DefaultSchedulerDaemon
}

func (d *SchedulePerformanceMeasure) Start() {
}

func (d *SchedulePerformanceMeasure) Stop(sig os.Signal) {
}

func (d *SchedulePerformanceMeasure) When() kkdaemon.CronSyntax {
	return "* * * * *"
}

func (d *SchedulePerformanceMeasure) Loop() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	performanceData := PerformanceData{
		RoutineCount:  runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		TotalMemory:   memory.TotalMemory() / 1024 / 1024,
		FreeMemory:    memory.FreeMemory() / 1024 / 1024,
		Alloc:         m.Alloc / 1024 / 1024,
		Sys:           m.Sys / 1024 / 1024,
		HeapAlloc:     m.HeapAlloc / 1024 / 1024,
		HeapSys:       m.HeapSys / 1024 / 1024,
		HeapInuse:     m.HeapInuse / 1024 / 1024,
		HeapIdle:      m.HeapIdle / 1024 / 1024,
		HeapReleased:  m.HeapReleased / 1024 / 1024,
		StackInuse:    m.StackInuse / 1024 / 1024,
		NumGC:         m.NumGC,
		NextGC:        m.NextGC / 1024 / 1024,
		PauseTotalMs:  m.PauseTotalNs / 1000000,
		LastGC:        m.LastGC,
		GCCPUFraction: m.GCCPUFraction,
		ActiveCount:   m.Mallocs - m.Frees,
	}

	kklogger.InfoJ("daemons:SchedulePerformanceMeasure.Loop#performance_measure", performanceData)
	return nil
}

type PerformanceData struct {
	RoutineCount  int     `json:"routine_count"`
	NumCPU        int     `json:"num_cpu"`
	TotalMemory   uint64  `json:"total_memory"`
	FreeMemory    uint64  `json:"free_memory"`
	Alloc         uint64  `json:"alloc"`
	Sys           uint64  `json:"sys"`
	HeapAlloc     uint64  `json:"heap_alloc"`
	HeapSys       uint64  `json:"heap_sys"`
	HeapInuse     uint64  `json:"heap_inuse"`
	HeapIdle      uint64  `json:"heap_idle"`
	HeapReleased  uint64  `json:"heap_released"`
	StackInuse    uint64  `json:"stack_inuse"`
	NumGC         uint32  `json:"num_gc"`
	NextGC        uint64  `json:"next_gc"`
	PauseTotalMs  uint64  `json:"pause_total_ms"`
	LastGC        uint64  `json:"last_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
	ActiveCount   uint64  `json:"active_count"`
}
