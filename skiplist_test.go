package skiplist

import (
	"math/rand"
	"testing"
	"time"
)

type i struct {
	k int
	v int
}

func (ii *i) Key() int {
	return ii.k
}

func (ii *i) Less(i2 ISkiplistElement[int]) bool {
	iii, ok := i2.(*i)
	if !ok {
		panic("cannot compare Less between different type, or check either self ptr or input ptr is nil")
	}
	return ii.v < iii.v
}

func TestSortSet(t *testing.T) {
	rand.Seed(time.Now().UnixMilli())
	l := NewSkipList[int]()
	for ii := 0; ii < 1000; ii++ {
		l.Add(&i{k: ii, v: ii})
	}
	//l.DeleteByKey(485) // 这是一个在最高索引层都存在的节点
	l.DeleteByKey(566) // 一个次最高索引层的节点（最高没有）
	l.Print()
}
