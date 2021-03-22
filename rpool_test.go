package rpool

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func BenchmarkNewRPool(b *testing.B) {
	p:=NewRPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			p.Execute(func() {

			})
		}
	})

}


func BenchmarkNewRPool2(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			go func() {}()
		}
	})

}

func BenchmarkNewRPool3(b *testing.B) {

	p:=NewRPS(runtime.NumCPU())
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			p.Execute(func() {

			})
		}
	})

}


func BenchmarkTime(b *testing.B) {
	st:=time.Now()
	for i:=0 ;i<b.N;i++{
		time.Since(st)
	}

}

func Test_getSize(t *testing.T) {
	fmt.Println(getAdapterSize(1))
	fmt.Println(getAdapterSize(2))
	fmt.Println(getAdapterSize(3))
	fmt.Println(getAdapterSize(32))
}
