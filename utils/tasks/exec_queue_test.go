package tasks

import (
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecQueue(t *testing.T) {
	var i int64
	q := NewExecutionQueue(1 * time.Millisecond)

	go q.Start()

	go func() {
		count := 100
		for count > 0 {
			count--
			q.Queue(func() error {
				atomic.AddInt64(&i, 1)
				return nil
			})
			q.Queue(func() error {
				atomic.AddInt64(&i, -1)
				return nil
			})
		}
	}()

	q.Queue(func() error {
		atomic.AddInt64(&i, 1)
		return nil
	})
	q.Wait()
	require.Equal(t, int64(1), atomic.LoadInt64(&i))
	require.Equal(t, 0, len(q.(*executionQueue).getWaiting()))
	require.Equal(t, 0, len(q.Errors()))
}

func TestExecQueue_Stop(t *testing.T) {
	var i int64
	q := NewExecutionQueue(1 * time.Millisecond)

	go q.Start()

	q.Queue(func() error {
		atomic.AddInt64(&i, 1)
		return nil
	})
	require.Equal(t, 1, len(q.(*executionQueue).getWaiting()))
	time.Sleep(2 * time.Millisecond)
	require.Equal(t, 0, len(q.(*executionQueue).getWaiting()))

	require.False(t, q.(*executionQueue).isStopped())
	q.Stop()
	require.True(t, q.(*executionQueue).isStopped())
	q.Queue(func() error {
		atomic.AddInt64(&i, 1)
		return nil
	})
	time.Sleep(2 * time.Millisecond)
	// q was stopped, therefore the function should be kept in waiting
	require.Equal(t, 1, len(q.(*executionQueue).getWaiting()))
	require.Equal(t, int64(1), atomic.LoadInt64(&i))
}

func TestExecQueue_QueueDistinct(t *testing.T) {
	var i int64
	q := NewExecutionQueue(2 * time.Millisecond)

	inc := func() error {
		atomic.AddInt64(&i, 1)
		return nil
	}
	q.QueueDistinct(inc, "1")
	q.QueueDistinct(inc, "1")
	q.QueueDistinct(inc, "1")
	require.Equal(t, 1, len(q.(*executionQueue).getWaiting()))
	go q.Start()
	defer q.Stop()
	// waiting for function to execute
	time.Sleep(4 * time.Millisecond)
	require.Equal(t, 0, len(q.(*executionQueue).getWaiting()))
	q.QueueDistinct(inc, "1")
	q.QueueDistinct(inc, "1")
	q.QueueDistinct(inc, "1")
	require.Equal(t, 1, len(q.(*executionQueue).getWaiting()))

}

func TestExecQueue_Empty(t *testing.T) {
	q := NewExecutionQueue(1 * time.Millisecond)

	go q.Start()

	q.Wait()
	q.Stop()
	require.True(t, q.(*executionQueue).isStopped())
}
