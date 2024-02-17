# template skiplist
模板化的跳表 Golang template skiplist

模板语法需要go sdk 1.18或以上，如果代码下下来有编译错误要先升到1.18

Template of Golang is supported only for Go SDK 1.18 or above.

## 基本使用 / Quick start

```go
package main

import (
	"fmt"
	"github.com/aruyuna9531/skiplist"
)

// 定义一个能插进skiplist的结构体，需要满足接口ISkipListElement
type i struct {
	k int
	v int
}

func (ii *i) Key() int {
	return ii.k
}

func (ii *i) Value() int {
	return ii.v
}

func (ii *i) Less(i2 skiplist.ISkiplistElement[int, int]) bool {
	iii, ok := i2.(*i)
	if !ok {
		panic("cannot compare Less between different type, or check either self ptr or input ptr is nil")
	}
	return ii.v < iii.v
}

func main() {
	// 创建一个新跳表实例
	l := skiplist.NewSkipList[int, int]()
	l.Add(&i{k: 1, v: 1})
	// ...其他要加的元素
	
	// 获得第一个元素的值
	elem, _ := l.GetElementByRank(1)
	
	// 获得元素1在表里的位置
	rank, _ := l.GetRankByKey(1)
	
	// 获得表里排名第8-15的元素
	elems, _ := l.GetRange(8,15)
}

```

! 排名规则是最小的元素排名值是1，最大的是倒数第1（等于元素个数）。0不会出现在排名的值里，和数组下标概念不一样。

! rank of elements are 1 to length(elements)

! 不提供单独修改一个元素的值（.v）的接口，因为这个操作还要调整这个节点在链表里的位置。 如果上层应用需要提供update值的操作，请在上层update里使用```DeleteByKey()```再```Add()```。

! update the value of existing node is not supported. if your design needs the operation of "update value only", please use ```DeleteByKey()``` and ```Add()``` orderly.
```go
type appStruct struct{
	set *skiplist.SkipList[keyType, valueType]
}

func (appPointer *appStruct) Update(key keyType, value valueType) {
	// nil assert if necessary
	appPointer.set.DeleteByKey(key)
	// error solve if necessary
	appPointer.set.Add(&...)
	// error solve if necessary
}
```

! 基本上应该能使，不过不保证没bug。

如果有bug欢迎提issue