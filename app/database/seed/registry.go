package seed

import (
	"sort"
	"sync"
	"time"

	kklogger "github.com/yetiz-org/goth-kklogger"
)

type Seed interface {
	Name() string
	Order() int
	Run() error
}

var (
	mu       sync.Mutex
	registry []Seed
)

// Register registers a seed implementation
func Register(s Seed) {
	mu.Lock()
	defer mu.Unlock()
	registry = append(registry, s)
}

// RunAll runs all registered seeds in defined order
func RunAll() error {
	mu.Lock()
	seeds := make([]Seed, len(registry))
	copy(seeds, registry)
	mu.Unlock()

	// sort by order then name for determinism
	sort.Slice(seeds, func(i, j int) bool {
		if seeds[i].Order() == seeds[j].Order() {
			return seeds[i].Name() < seeds[j].Name()
		}
		return seeds[i].Order() < seeds[j].Order()
	})

	for _, s := range seeds {
		start := time.Now()
		kklogger.InfoJ("seed:RunAll#exec!start", "running seed: "+s.Name())
		if err := s.Run(); err != nil {
			kklogger.ErrorJ("seed:RunAll#exec!seed", "seed "+s.Name()+" failed: "+err.Error())
			return err
		}
		kklogger.InfoJ("seed:RunAll#exec!done", "seed "+s.Name()+" completed in "+time.Since(start).String())
	}
	return nil
}
