package rpool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type RoutinePool interface {
	Execute(task Task)
}

type Task func()

type worker struct {
	task           chan Task
	lastActiveTime time.Time
}

type RPool struct {
	workers          []*worker
	lock             sync.Mutex
	liveScanInterval time.Duration
	workerLiveTime   time.Duration
	//workers cache
	workersTempPool sync.Pool
	//worker cache
	//workerPool sync.Pool
}

// 设置扫描处于空闲状态的协程的间隔时间
func WithOpLiveScanIt(interval time.Duration) Option {
	return func(p *RPool) {
		p.liveScanInterval = interval
	}
}

//设置空闲协程的存活时间
func WithOpWorkerLiveTime(liveTime time.Duration) Option {
	return func(p *RPool) {
		p.workerLiveTime = liveTime
	}
}

type Option func(p *RPool)

/*
创建一个协程池
*/
func NewRPool(options ...Option) *RPool {
	p := &RPool{
		workers: make([]*worker, 0, 1000),
	}
	for _, opt := range options {
		opt(p)
	}
	if p.workerLiveTime <= 0 {
		p.workerLiveTime = 30 * time.Second
	}
	if p.liveScanInterval <= 0 {
		p.liveScanInterval = 30 * time.Second
	}
	go p.clearInactiveWorkers()
	return p
}

func (this *RPool) runWorker(w *worker) {
	for {
		t, ok := <-w.task
		if !ok {
			return
		}
		t()
		w.lastActiveTime = time.Now()
		this.lock.Lock()
		this.workers = append(this.workers, w)
		this.lock.Unlock()
	}
}

func (this *RPool) Execute(task Task) {
	this.lock.Lock()
	n := len(this.workers) - 1
	var w *worker
	if n <= 0 {
		w = &worker{
			task: make(chan Task),
		}
		go this.runWorker(w)
		this.lock.Unlock()
		w.task <- task
		return
	}

	w = this.workers[n]
	this.workers = this.workers[:n]
	this.lock.Unlock()
	w.task <- task
}

func (this *RPool) newTempWorkers() []*worker {
	wi := this.workersTempPool.Get()
	if wi == nil {
		return make([]*worker, 0, 5)
	}
	return wi.([]*worker)[:0]
}

func (this *RPool) clearInactiveWorkers() {
	tick := time.Tick(this.liveScanInterval)
	for range tick {
		this.lock.Lock()
		inactive := this.newTempWorkers()
		i := 0
		for i = 0; i < len(this.workers); i++ {
			w := this.workers[i]
			if time.Since(w.lastActiveTime) > this.workerLiveTime {
				inactive = append(inactive, w)
			} else { //当
				break
			}
		}
		this.workers = this.workers[i:]
		this.lock.Unlock()
		for _, w := range inactive {
			close(w.task)
		}
		this.workersTempPool.Put(inactive)
	}
}

// 创建多个 协程池的集合，使用原子自增的方式散列任务到多个协程池，
// 减小锁的竞争几率
type RPS struct {
	pools []RoutinePool
	idx   int64
	size  int64
}

func NewRPS(size int, creater func() RoutinePool) RoutinePool {
	if size <= 0 {
		size = runtime.NumCPU()
	}
	size = getAdapterSize(size)
	rps := &RPS{
		size:  int64(size - 1),
		pools: make([]RoutinePool, size),
	}
	for i := 0; i < size; i++ {
		rps.pools[i] = creater()
	}
	return rps
}

func (r *RPS) Execute(task Task) {
	idx := atomic.AddInt64(&r.idx, 1) & r.size
	r.pools[idx].Execute(task)
}

// 获取size 向下的 2的整数次方的值
func getAdapterSize(size int) int {
	s := 0
	for size > 1 {
		s++
		size /= 2
	}
	return 1 << s
}

// 1000
type pool2 struct {
	task chan Task
}

func newPool2() *pool2 {
	p := &pool2{
		task: make(chan Task, 1000),
	}
	go p.run()
	return p
}

func (this *pool2) Execute(task Task) {
	this.task <- task
}

func (this *pool2) run() {
	for {
		t := <-this.task
		go func(t Task) {
			//timer:=time.NewTimer(30*time.Second)
			for {
				t()
				select {
				case t = <-this.task:
					//timer.Reset(30*time.Second)
					//case <-timer.C:
					//	return
				}
			}
		}(t)
		runtime.Gosched()

	}
}
