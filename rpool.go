package rpool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Task func()

type worker struct {
	task chan Task
	lastActiveTime time.Time
}

type RPool struct {
	workers []*worker
	lock    sync.Mutex
	liveScanInterval time.Duration
	workerLiveTime time.Duration
}
// 设置扫描处于空闲状态的协程的间隔时间
func WithOpLiveScanIt(interval time.Duration)Option{
	return func(p *RPool) {
		p.liveScanInterval = interval
	}
}

//设置空闲协程的存活时间
func WithOpWorkerLiveTime(liveTime time.Duration)Option{
	return func(p *RPool) {
		p.workerLiveTime = liveTime
	}
}

type Option func(p *RPool)
/*
创建一个协程池
 */
func NewRPool(options ...Option) *RPool {
	p:= &RPool{
		workers: make([]*worker, 0, 2000),
	}
	for _,opt := range options {
		opt(p)
	}
	if p.workerLiveTime <=0{
		p.workerLiveTime = 30*time.Second
	}
	if p.liveScanInterval <=0{
		p.liveScanInterval = 30 *time.Second
	}
	go p.clearInactiveWorkers()
	return p
}

func (this *RPool) Execute(task Task) {
	this.lock.Lock()
	n := len(this.workers) - 1
	var w *worker
	if n <= 0 {
		w = &worker{
			task: make(chan Task),
		}
		go func(w *worker, p *RPool) {
			for {
				t,ok := <-w.task
				if !ok{
					return
				}
				t()
				w.lastActiveTime = time.Now()
				p.lock.Lock()
				p.workers = append(p.workers, w)
				p.lock.Unlock()
			}
		}(w, this)
		this.lock.Unlock()
		w.task <- task
		return
	}

	w = this.workers[n]
	this.workers = this.workers[:n]
	this.lock.Unlock()
	w.task <- task
}

func (this *RPool) clearInactiveWorkers(){
	tick:=time.Tick(this.liveScanInterval)
	for range tick {
		this.lock.Lock()
		inactive:=make([]*worker,0, len(this.workers)/2)
		i:=0
		for i=0 ;i< len(this.workers);i++{
			w:=this.workers[i]
			if time.Since(w.lastActiveTime)> this.workerLiveTime{
				inactive = append(inactive,w)
			}else{
				break
			}
		}
		this.workers = this.workers[i:]
		this.lock.Unlock()
		for _,w := range inactive {
			close(w.task)
		}

	}
}

// 创建多个 协程池的集合，使用原子自增的方式提交任务到多个协程池，
// 减小锁的竞争几率
type RPS struct {
	pools []*RPool
	idx int64
	size int64
}

func NewRPS(size int,opts... Option) *RPS{
	if size<=0{
		size = runtime.NumCPU()
	}else{
		size = getAdapterSize(size)
	}
	rps:=&RPS{
		size: int64(size-1),
		pools: make([]*RPool,size),
	}
	for i:=0;i< size;i++{
		rps.pools[i] = NewRPool(opts...)
	}
	return rps
}

func (r *RPS)Execute(task Task){
	idx:=atomic.AddInt64(&r.idx,1) & r.size
	r.pools[idx].Execute(task)
}
// 获取size 向下的 2的整数次方的值
func getAdapterSize(size int)int{
	s:=0
	for size>1{
		s++
		size/=2
	}
	return 1<<s
}
