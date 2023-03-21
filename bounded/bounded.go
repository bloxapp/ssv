package bounded

import (
	"os"
	"runtime"
	"strconv"
	"sync"
)

var outChanPool = sync.Pool{
	New: func() interface{} {
		return make(chan error, 1)
	},
}

type job struct {
	f   func() error
	out chan<- error
}

var in = make(chan job, 1024)

func init() {
	// Get the number of available CPUs.
	//
	// SSV_NUM_CPU is a useful override when running in environments
	// such as Kubernetes, where the runtime.NumCPU() is wrong.
	numCPU := runtime.NumCPU()
	if os.Getenv("SSV_NUM_CPU") != "" {
		numCPU, _ = strconv.Atoi(os.Getenv("SSV_NUM_CPU"))
		if numCPU < 1 {
			numCPU = 1
		}
	}

	// Set GOMAXPROCS to the number of available CPUs, but at least 10.
	goMaxProcs := numCPU
	if goMaxProcs < 10 {
		goMaxProcs = 10
	}
	runtime.GOMAXPROCS(goMaxProcs)

	// Create numCPU + 1 goroutines to do CGO calls.
	cgoroutines := numCPU + 1

	for i := 0; i < cgoroutines; i++ {
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			for j := range in {
				err := j.f()
				j.out <- err
			}
		}()
	}
}

// CGO runs the given function in a goroutine dedicated to CGO calls,
// and returns the error returned by the function.
//
// This helps bound the number of different goroutines that call CGO
// to a fixed number of goroutines with locked OS threads, thereby
// reducing the number of OS threads that CGO creates and destroyes.
func CGO(f func() error) error {
	out := outChanPool.Get().(chan error)
	defer outChanPool.Put(out)

	in <- job{f, out}
	return <-out
}
