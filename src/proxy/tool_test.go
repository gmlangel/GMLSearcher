package proxy

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

//单元测试
func TestMD5(t *testing.T) {
	str := MakeMD5("test")
	fmt.Println(str)
}

//性能测试
func BenchmarkMD5(b *testing.B) {
	str := MakeMD5("test")
	fmt.Println(str)
}

type bbb struct{ x, y int }

func TestCopy(t *testing.T) {
	// //指针类型的切片与copy比对
	// arr := []*bbb{
	// 	&bbb{0, 1}, &bbb{1, 1}, &bbb{2, 1}, &bbb{3, 1}, &bbb{4, 1}}
	// arr2 := make([]*bbb, 1)
	// copy(arr2, arr[2:3])
	// fmt.Println(fmt.Sprintf("copy后的元素地址===%p,%p", &arr[2], &arr2[0]))
	// fmt.Println(fmt.Sprintf("copy后的元素的值===%p,%p", arr[2], arr2[0]))

	// arr3 := arr[2:3]
	// fmt.Println(fmt.Sprintf("切片后的元素地址===%p,%p", &arr[2], &arr3[0]))
	// fmt.Println(fmt.Sprintf("切片后的元素的值===%p,%p", arr[2], arr3[0]))

	a := []bbb{bbb{0, 1}, bbb{2, 3}}
	b := a
	fmt.Println(a, b)
	a[0] = bbb{13, 1}
	fmt.Println(a, b)

	// //值类型的切片与copy比对
	// arr := []int{0, 1, 2, 3, 4, 5}
	// arr2 := make([]int, 1)
	// copy(arr2, arr[2:3])
	// fmt.Println(fmt.Sprintf("copy后的元素地址===%p,%p。数组的起始地址为%p", &arr[2], &arr2[0], &arr))

	// arr3 := arr[2:3]
	// fmt.Println(fmt.Sprintf("切片后的元素地址===%p,%p。数组的起始地址为%p", &arr[2], &arr3[0], &arr))

	// fmt.Println(arr2[0])
	// arr[2].x = 13
	// fmt.Println(arr2[0])

	// // arr := []bbb{
	// // 	bbb{0, 1}, bbb{1, 1}, bbb{2, 1}, bbb{3, 1}, bbb{4, 1}}
	// // arr2 := make([]bbb, 1)
	// // copy(arr2, arr[2:3])
	// // fmt.Println(fmt.Sprintf("%p,%p", &arr[2], &arr2[0]))
	// // arr[2].x = 13vb
	// // fmt.Println(arr2[0])

	// arr := []int{0, 1, 2, 3, 4}
	// arr2 := []int{0}
	// fmt.Println(fmt.Sprintf("%p,%p,%p", &arr, &arr[0], &arr2[0]))
}

var a string
var once sync.Once

func setup() {
	a = "hello, world2"
	time.Sleep(time.Minute * 1)
}

func doprint(n int) {
	println("进入第", n, '次')
	once.Do(setup)
	println("执行第", n, "次打印", a)
}

func TestTwoprint(t *testing.T) {
	go doprint(1)
	println("执行中间函数")
	go doprint(2)
	time.Sleep(time.Minute * 5)
}
