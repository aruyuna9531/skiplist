package skiplist

type ISkiplistElement[KeyType comparable] interface {
	Key() KeyType // 在set里能索引到节点的key
	// Less 定义自身应该在sortset内部排在输入参数前面时返回true。排在后面返回false。
	// 如果相互Less返回了相同的布尔值，两者的相对位置将随机排列，定义Less时尽量不要得到互相Less比较返回相同值的情况，除非同分同等对待等特殊情形。（此时EqualElementExistence应置为true）
	Less(ISkiplistElement[KeyType]) bool
}

func SkiplistElementCompareLess[KeyType comparable](i, j ISkiplistElement[KeyType]) bool {
	return i.Less(j) && !j.Less(i)
}
func SkiplistElementCompareGreater[KeyType comparable](i, j ISkiplistElement[KeyType]) bool {
	return j.Less(i) && !i.Less(j)
}
func SkiplistElementCompareEqual[KeyType comparable](i, j ISkiplistElement[KeyType]) bool {
	return j.Less(i) == i.Less(j)
}
