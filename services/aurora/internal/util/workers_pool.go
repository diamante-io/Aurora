package util

import (
	"sync"
)

// WorkersPool provides a simple synchronization between workers doing some work.
// First add jobs by using AddWork(), then add a worker function using
// `SetWorker`. Finally, start as many workers as you want using `Start`.
type WorkersPool struct {
	work       []interface{}
	workerFunc func(workerID int, work interface{})
	mutex      sync.Mutex
}

func (w *WorkersPool) AddWork(job interface{}) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.work = append(w.work, job)
}

func (w *WorkersPool) WorkSize() int {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return len(w.work)
}

// getWork returns a work. The second return value is false if there's no more work.
func (w *WorkersPool) getWork() (interface{}, bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.work) == 0 {
		return nil, false
	}

	work := w.work[0]
	w.work = w.work[1:]
	return work, true
}

// SetWorker defines function that will be run by workers. workerFunc accepts two
// parameters:
//   * first provides ID of the worker [0, workers-1],
//   * seconds provides data that worker should work on.
func (w *WorkersPool) SetWorker(workerFunc func(int, interface{})) {
	w.workerFunc = workerFunc
}

// Start starts worker to execute workerFunc
func (w *WorkersPool) Start(workers int) {
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			for {
				job, ok := w.getWork()
				if !ok {
					wg.Done()
					return
				}

				w.workerFunc(workerID, job)
			}
		}(i)
	}
	wg.Wait()
}
