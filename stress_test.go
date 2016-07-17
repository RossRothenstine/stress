package stress_test

import (
	"log"
	"testing"
	"time"

	"github.com/RossRothenstine/stress"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

type emptyWorker struct {
}

func (_ emptyWorker) Work(context.Context) {
}

func TestAllocatesWorkersFromPool(t *testing.T) {
	called := false
	stresser := &stress.Stresser{
		New: func() stress.Worker {
			called = true
			return emptyWorker{}
		},
		Limit: rate.NewLimiter(10, 10),
	}

	go stresser.Start()
	defer stresser.Stop()

	timer := time.NewTimer(time.Second)
	<-timer.C

	if called == false {
		t.Error("expected true, got false")
	}
}

type workTracker struct {
	numCalls int
}

func (w *workTracker) Work(context.Context) {
	w.numCalls++
}

func TestLowerBound(t *testing.T) {
	worker := &workTracker{}
	stresser := &stress.Stresser{
		New: func() stress.Worker {
			return worker
		},
		Limit: rate.NewLimiter(20, 20),
	}

	go stresser.Start()
	defer stresser.Stop()

	timer := time.NewTimer(time.Second)
	<-timer.C

	if worker.numCalls < 10 {
		t.Errorf("incorrect number of calls, expected at least %d, got %d", 10, worker.numCalls)
	}
}

type slowWorker struct {
	delay time.Duration
}

func (w *slowWorker) Work(context.Context) {
	time.Sleep(w.delay)
}

func TestRampUp(t *testing.T) {
	log.Println("Start Ramp Up")
	calls := 0
	stresser := &stress.Stresser{
		New: func() stress.Worker {
			calls++
			return &slowWorker{
				delay: time.Second / 2,
			}
		},
		Limit: rate.NewLimiter(10, 10),
	}

	go stresser.Start()
	defer stresser.Stop()

	timer := time.NewTimer(5 * time.Second)
	<-timer.C

	if calls == 1 {
		t.Errorf("incorrect number of workers created, expected at least more than %d, got %d", 1, calls)
	}
}
