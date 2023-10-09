package pool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultExpire = 10 //秒

var (
	ErrorInvalidCap    = errors.New("pool cap can not < 0")
	ErrorInvalidExpire = errors.New("pool expire can not < 0")
	ErrorHasClose      = errors.New("pool has been release")
)

type sig struct {
}

type Pool struct {
	//容量
	cap int32
	//正在运行的worker数量
	running int32
	//空闲worker
	workers []*Worker
	//过期时间，空闲超过这个时间，回收
	expire time.Duration
	//释放资源
	release      chan sig
	lock         sync.Mutex
	once         sync.Once
	workerCache  sync.Pool
	cond         *sync.Cond
	PanicHandler func()
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewTimePool(cap int, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrorInvalidCap
	}
	if expire <= 0 {
		return nil, ErrorInvalidExpire
	}

	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
	}
	p.workerCache.New = func() interface{} {
		return &Worker{
			pool: p,
			task: make(chan func(), 1),
		}
	}
	p.cond = sync.NewCond(&p.lock)
	go p.expireWorker()
	return p, nil
}

func (p *Pool) expireWorker() {
	//定期清理空闲worker
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClose() {
			break
		}
		//循环空闲的worker，如果空闲时间大于expire进行清理
		p.lock.Lock()
		idleWorkers := p.workers
		expireIndex := -1
		if len(idleWorkers) > 0 {
			for i := len(idleWorkers) - 1; i >= 0; i-- {
				if time.Now().Sub(idleWorkers[i].lastTime) >= p.expire {
					expireIndex = i
					break
				}
			}
			//如果i位置的worker超时了，那么前面的都超时
			if expireIndex != -1 {
				for i, w := range idleWorkers {
					if i > expireIndex {
						break
					}
					//结束任务
					w.task <- nil
				}
				p.workers = idleWorkers[expireIndex+1:]
			}
		}
		p.lock.Unlock()
		fmt.Printf("清理完成，running:%d，workers:%v \n", len(p.workers), p.workers)
	}
	ticker.Stop()
}

//提交任务
func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return ErrorHasClose
	}
	//获取一个worker，执行任务
	w := p.GetWorker()
	w.task <- task
	return nil
}

func (p *Pool) GetWorker() *Worker {
	//如果有空闲的直接获取
	p.lock.Lock()
	defer p.lock.Unlock()

	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		w := idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		return w
	}

	//没有空闲的，新建
	if p.running < p.cap {
		c := p.workerCache.Get()
		var w *Worker
		if c == nil {
			w = &Worker{pool: p, task: make(chan func(), 1)}
		} else {
			w = c.(*Worker)
		}
		w.run()
		return w
	}

	//如果worker数量大于pool容量，阻塞等待，worker释放
	return p.waitIdleWorker()
}

func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.cond.Wait()

	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		if p.running < p.cap {
			c := p.workerCache.Get()
			var w *Worker
			if c == nil {
				w = &Worker{pool: p, task: make(chan func(), 1)}
			} else {
				w = c.(*Worker)
			}
			w.run()
			return w
		}
		return p.waitIdleWorker()
	}
	w := idleWorkers[n]
	idleWorkers[n] = nil
	p.workers = idleWorkers[:n]

	return w
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.cond.Signal()
	p.lock.Unlock()
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) Release() {
	p.once.Do(func() {
		p.lock.Lock()
		defer p.lock.Unlock()
		for i, w := range p.workers {
			w.task = nil
			w.pool = nil
			p.workers[i] = nil
		}
		p.workers = nil
		p.release <- sig{}
	})
}

func (p *Pool) IsClose() bool {
	return len(p.release) > 0
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = <-p.release
	go p.expireWorker()
	return true
}
