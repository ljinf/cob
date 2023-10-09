package pool

import (
	"github.com/ljinfu/cob/log"
	"time"
)

type Worker struct {
	pool *Pool
	//任务
	task chan func()
	//执行任务的时间
	lastTime time.Time
}

func (w *Worker) run() {
	w.pool.incRunning()
	go w.running()
}

func (w *Worker) running() {
	defer func() {
		w.pool.decRunning()
		w.pool.workerCache.Put(w)
		if err := recover(); err != nil {
			//捕获任务发生的panic
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler()
			} else {
				log.Default().Error(err)
			}
		}
		w.pool.cond.Signal()
	}()
	for t := range w.task {
		if t == nil {
			w.pool.workerCache.Put(w)
			return
		}
		t()
		//任务执行完，归还
		w.pool.PutWorker(w)
	}
}
