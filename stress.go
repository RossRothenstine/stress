package stress

import (
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

// A Worker does a unit of work towards the stress test.
type Worker interface {
	// Work should perform the unit of work for this worker.
	// The context provided is shared amongst all workers, and is derived
	// from the Stresser's context before calling Stresser.Start().
	Work(context.Context)
}

type Stresser struct {
	// Context is a shared state amongst all workers.
	Context context.Context

	// New allocates a new worker.
	New func() Worker

	Limit *rate.Limiter

	workers []Worker
	close   chan struct{}
}

func (s *Stresser) Start() {
	s.close = make(chan struct{})
	if s.Context == nil {
		s.Context = context.TODO()
	}
	// Always allocate at least one worker
	w := s.New()
	s.workers = []Worker{w}

	for {
		select {
		case <-s.close:
			return
		default:
			s.doWork()
		}
	}
}

func (s *Stresser) doWork() {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	events := 0
	for {
		select {
		case <-timer.C:
			if events < int(s.Limit.Limit()) {
				w := s.New()
				s.workers = append(s.workers, w)
			}
			return
		default:
			wg := sync.WaitGroup{}
			wg.Add(len(s.workers))
			// Run a single unit of work on all workers
			for _, w := range s.workers {
				go func() {
					defer wg.Done()
					ok := s.Limit.Allow()
					if !ok {
						return
					}
					w.Work(s.Context)
					events++
				}()
			}
			wg.Wait()

		}
	}
}

func (s *Stresser) Stop() {
	s.close <- struct{}{}
}
