package rpool

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func BenchmarkNewRPool(b *testing.B) {
	p := NewRPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Execute(func() {

			})
		}
	})

}

func BenchmarkNewRPool2(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			go func() {

			}()
		}
	})
}

func BenchmarkNewRPool3(b *testing.B) {

	p := NewRPS(runtime.NumCPU(), func() RoutinePool {
		return NewRPool()
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Execute(func() {

			})
		}
	})

}

func Test_getSize(t *testing.T) {
	fmt.Println(getAdapterSize(1))
	fmt.Println(getAdapterSize(2))
	fmt.Println(getAdapterSize(3))
	fmt.Println(getAdapterSize(32))
	fmt.Println(getAdapterSize(12))
}

func TestNewRPool(t *testing.T) {
	k := 0
	p := NewRPS(runtime.NumCPU(), func() RoutinePool {
		return NewRPool(WithOpLiveScanIt(1*time.Second), WithOpWorkerLiveTime(1*time.Second))
	})
	fmt.Println(k, runtime.NumGoroutine())
	//p:=NewRPool()
	for i := 0; i < 100000; i++ {
		p.Execute(func() {
			k++
		})
	}

	time.Sleep(1 * time.Second)
	fmt.Println(k, runtime.NumGoroutine())
	time.Sleep(5 * time.Second)
	fmt.Println(k, runtime.NumGoroutine())
}

func TestW(t *testing.T) {


}

func BenchmarkNewRPS(b *testing.B) {
	p := NewRPS(8, func() RoutinePool {
		return newPool2()
	})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Execute(func() {

			})
		}
	})
}

