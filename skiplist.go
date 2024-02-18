package skiplist

import (
	"bytes"
	"fmt"
	"log"
)

type SameKeyOperateType = int32

const (
	SameKeyRejected  SameKeyOperateType = 0
	SameKeyOverwrite SameKeyOperateType = 1
)

const (
	baseLayer      = 1
	createIndexGap = 5
	sameKeyOperate = SameKeyRejected // add了相同key的元素时处理方式
)

type SkipListNode[KeyType comparable] struct {
	v        ISkiplistElement[KeyType]
	frontPtr *SkipListNode[KeyType]
	backPtr  *SkipListNode[KeyType]
	downPtr  *SkipListNode[KeyType]
	upPtr    *SkipListNode[KeyType]

	nodeInLayer int32
	span        int32 // 与它的front之间的数据节点数（包括自身对应的底层节点，不包括front对应的）
	linkNodes   int32 // 与front之间的下层索引节点数（同上）
	isHead      bool
	isTail      bool
}

//func (sln *SkipListNode[KeyType]) Key() KeyType {
//	return sln.v.Key()
//}
//func (sln *SkipListNode[KeyType]) Less(other ISkiplistElement[KeyType]) bool {
//	return sln.v.Less(other)
//}

//func (sln *SkipListNode[KeyType]) GetV() ISkiplistElement[KeyType] {
//	return sln.v
//}

func (sln *SkipListNode[KeyType]) insertFront(node *SkipListNode[KeyType]) {
	if sln == nil || node == nil {
		return
	}
	node.backPtr = sln
	node.frontPtr = sln.frontPtr
	if sln.frontPtr != nil {
		sln.frontPtr.backPtr = node
	}
	sln.frontPtr = node
	p := node.findNearestUpperBack()
	if p != nil {
		p.linkNodes++
	}
	for ; p != nil; p = p.findNearestUpperBack() {
		p.span++
	}
}

// pushUpperLayer 在本节点的上一层创建索引节点
func (sln *SkipListNode[KeyType]) pushUpperLayer() {
	if sln == nil || sln.upPtr != nil {
		return
	}
	frontTo := sln
	frontStep := int32(0)
	frontLink := int32(0)
	for frontTo != nil && frontTo.upPtr == nil {
		frontStep += frontTo.span
		frontTo = frontTo.frontPtr
		frontLink++
	}
	if frontTo == nil {
		// 设计上不能出现这种情况，如果该层是新索引层，需要在调用此函数前事先在该层建立一个指向首元素的索引节点（即headPtr）
		panic("SkipListNode::pushUpperLayer error: cannot push to an empty layer")
	} else {
		frontTo = frontTo.upPtr
	}
	backTo := sln
	backStep := int32(0)
	backLink := int32(0)
	for backTo != nil && backTo.upPtr == nil {
		backTo = backTo.backPtr
		backStep += backTo.span
		backLink++
	}
	if backTo == nil {
		// 同上
		panic("SkipListNode::pushUpperLayer error: cannot push to an empty layer")
	} else {
		backTo = backTo.upPtr
	}
	node := &SkipListNode[KeyType]{
		v:           sln.v,
		frontPtr:    frontTo,
		backPtr:     backTo,
		downPtr:     sln,
		upPtr:       nil,
		nodeInLayer: sln.nodeInLayer + 1,
		span:        frontStep,
		linkNodes:   frontLink,
	}
	frontTo.backPtr = node
	backTo.frontPtr = node
	sln.upPtr = node
	// 更新上层后方节点斯潘（因为它中间插了一个）
	backTo.span = backStep
	backTo.linkNodes = backLink
	// 新增节点的左后方索引link+1（如果存在）
	ub := node.findNearestUpperBack()
	if ub != nil {
		ub.linkNodes++
	}
}

func (sln *SkipListNode[KeyType]) findNearestUpperBack() *SkipListNode[KeyType] {
	if sln == nil {
		return nil
	}
	backTo := sln
	for backTo != nil && backTo.upPtr == nil {
		backTo = backTo.backPtr
	}
	if backTo == nil {
		return nil
	}
	return backTo.upPtr
}
func (sln *SkipListNode[KeyType]) findNearestUpperFront() *SkipListNode[KeyType] {
	if sln == nil {
		return nil
	}
	frontTo := sln
	for frontTo != nil && frontTo.upPtr == nil {
		frontTo = frontTo.frontPtr
	}
	if frontTo == nil {
		return nil
	}
	return frontTo.upPtr
}

// compareNode -1 比入参小 / 0 与入参等值 / 1 比入参大 TODO 是否限制同层比较
func (sln *SkipListNode[KeyType]) compareNode(other *SkipListNode[KeyType]) int {
	if sln.isHead && !other.isHead {
		return -1
	}
	if !sln.isHead && other.isHead {
		return 1
	}
	if sln.isTail && !other.isTail {
		return 1
	}
	if !sln.isTail && other.isTail {
		return -1
	}
	if SkiplistElementCompareGreater(sln.v, other.v) {
		return 1
	}
	if SkiplistElementCompareLess(sln.v, other.v) {
		return -1
	}
	return 0
}

func (sln *SkipListNode[KeyType]) compareElem(k ISkiplistElement[KeyType]) int {
	if sln.isHead {
		return -1
	}
	if sln.isTail {
		return 1
	}
	if sln.v.Less(k) {
		return -1
	}
	if k.Less(sln.v) {
		return 1
	}
	return 0
}

type SkipList[KeyType comparable] struct {
	layersHead map[int32]*SkipListNode[KeyType]
	layersTail map[int32]*SkipListNode[KeyType]
	dict       map[KeyType]*SkipListNode[KeyType] // kv检索图
}

func NewSkipList[KeyType comparable]() *SkipList[KeyType] {
	return &SkipList[KeyType]{
		layersHead: make(map[int32]*SkipListNode[KeyType]),
		layersTail: make(map[int32]*SkipListNode[KeyType]),
		dict:       make(map[KeyType]*SkipListNode[KeyType]),
	}
}
func (sl *SkipList[KeyType]) GetLayersCount() int32 {
	return int32(len(sl.layersHead))
}

func (sl *SkipList[KeyType]) GetElementsCount() int32 {
	if sl == nil {
		return 0
	}
	return int32(len(sl.dict))
}

func (sl *SkipList[KeyType]) Add(e ISkiplistElement[KeyType]) (ret error) {
	if sl == nil {
		return fmt.Errorf("Sortset::Add error: self pointer = nil")
	}
	if e == nil {
		return fmt.Errorf("Sortset::Add error: input param e = nil")
	}
	if _, keyExist := sl.dict[e.Key()]; keyExist {
		switch sameKeyOperate {
		case SameKeyRejected:
			return fmt.Errorf("Sortset::Add error: key of input param deprecated (%v)", e.Key())
		case SameKeyOverwrite:
			err := sl.DeleteByKey(e.Key())
			if err != nil {
				return err
			}
		}
	}
	node := &SkipListNode[KeyType]{
		v:           e,
		frontPtr:    nil,
		backPtr:     nil,
		downPtr:     nil,
		upPtr:       nil,
		nodeInLayer: baseLayer, // 第1层一定要加个新节点的 就它了
		span:        1,
	}
	defer func() {
		if ret == nil {
			sl.dict[e.Key()] = node
		}
	}()
	if sl.GetElementsCount() == 0 {
		sl.layersHead[baseLayer] = &SkipListNode[KeyType]{
			frontPtr:    node,
			nodeInLayer: baseLayer,
			span:        1,
			isHead:      true,
		}
		sl.layersTail[baseLayer] = &SkipListNode[KeyType]{
			backPtr:     node,
			nodeInLayer: baseLayer,
			span:        0,
			isTail:      true,
		}
		node.backPtr = sl.layersHead[baseLayer]
		node.frontPtr = sl.layersTail[baseLayer]
	} else {
		for findPtr := sl.layersHead[sl.GetLayersCount()]; findPtr != nil; findPtr = findPtr.frontPtr {
			if findPtr.compareElem(e) == 0 {
				// 如果是overwrite模式那之前相同值元素会被删掉
				return fmt.Errorf("SkipList::Add error: cannot add node whose value equal to an existing node. (set EqualElementExistence = true if necessary)")
			}
			if findPtr.compareElem(e) < 0 {
				next := findPtr.frontPtr
				if next == nil || next.compareElem(e) > 0 {
					if findPtr.nodeInLayer == baseLayer {
						// 就插在这里了
						findPtr.insertFront(node)
						np := node
						needLoop := true
						for needLoop {
							if _, ex := sl.layersHead[np.nodeInLayer+1]; ex {
								p := np.findNearestUpperBack()
								if p.linkNodes <= createIndexGap {
									needLoop = false
									continue
								}
								tmpP := p.downPtr
								for i := 0; i < (createIndexGap+1)/2; i++ {
									tmpP = tmpP.frontPtr
									if tmpP.upPtr != nil {
										panic("p.up != nil when loop not finished")
									}
								}
								tmpP.pushUpperLayer()
								np = tmpP.upPtr
							} else {
								// 还没有上一层的情况
								p := sl.layersHead[np.nodeInLayer]
								tmpP := p
								needPush := true
								for i := 0; i < createIndexGap; i++ {
									tmpP = tmpP.frontPtr
									if tmpP.isTail {
										// 不需要上冒
										needPush = false
										break
									}
								}
								if !needPush {
									needLoop = false
									continue
								}

								for i := 0; i < (createIndexGap+1)/2; i++ {
									p = p.frontPtr
									if p == nil {
										panic("p = nil when loop is not finished")
									}
								}
								hNode := &SkipListNode[KeyType]{
									downPtr:     sl.layersHead[np.nodeInLayer],
									nodeInLayer: np.nodeInLayer + 1,
									isHead:      true,
								}
								tNode := &SkipListNode[KeyType]{
									downPtr:     sl.layersTail[np.nodeInLayer],
									upPtr:       nil,
									nodeInLayer: np.nodeInLayer + 1,
									isTail:      true,
								}
								hNode.frontPtr = tNode
								tNode.backPtr = hNode
								hNode.downPtr.upPtr = hNode
								tNode.downPtr.upPtr = tNode
								sl.layersHead[np.nodeInLayer+1] = hNode
								sl.layersTail[np.nodeInLayer+1] = tNode
								log.Printf("push new index layer at %d", np.nodeInLayer+1)
								p.pushUpperLayer()
								np = p.upPtr
							}
						}
						return nil
					} else {
						findPtr = findPtr.downPtr
					}
				}
			}
		}
	}
	return nil
}

func (sl *SkipList[KeyType]) delete(e *SkipListNode[KeyType]) (ret error) {
	h := sl.layersHead[sl.GetLayersCount()]
	for p := h; p != nil && !p.isTail; {
		if p.compareNode(e) == 0 {
			oldP := p
			p = p.downPtr
			if oldP.nodeInLayer != baseLayer {
				f := oldP.frontPtr
				b := oldP.backPtr
				b.span = b.span + oldP.span
				b.linkNodes = b.linkNodes + oldP.linkNodes
				b.frontPtr = oldP.frontPtr
				f.backPtr = oldP.backPtr
				oldP.downPtr.upPtr = nil
				if oldP.linkNodes+oldP.backPtr.linkNodes > createIndexGap {
					// 对比一下自身link与后节点link值，自身大往右倾斜，否则往左
					if oldP.linkNodes >= oldP.backPtr.linkNodes {
						if p.frontPtr.upPtr == nil {
							p.frontPtr.pushUpperLayer()
						}
					} else {
						if p.backPtr.upPtr == nil {
							p.backPtr.pushUpperLayer()
						}
					}
				}
				oldP.frontPtr = nil
				oldP.backPtr = nil
				oldP.downPtr = nil
				if sl.layersHead[oldP.nodeInLayer].frontPtr == sl.layersTail[oldP.nodeInLayer] {
					// 这一层没节点了，删掉这一层
					b.downPtr.upPtr = nil
					f.downPtr.upPtr = nil
					delete(sl.layersHead, oldP.nodeInLayer)
					delete(sl.layersTail, oldP.nodeInLayer)
				}
			} else {
				if oldP.backPtr.isHead && oldP.frontPtr.isTail {
					delete(sl.layersHead, oldP.nodeInLayer)
					delete(sl.layersTail, oldP.nodeInLayer)
				} else {
					oldP.backPtr.frontPtr = oldP.frontPtr
					oldP.frontPtr.backPtr = oldP.backPtr
					ub := oldP.findNearestUpperBack()
					for ub != nil {
						ub.linkNodes--
						ub.span--
						ub = ub.findNearestUpperBack()
					}
				}
			}
			continue
		}
		if p.compareNode(e) < 0 {
			if p.frontPtr.compareNode(e) <= 0 {
				p = p.frontPtr
			} else {
				if p.nodeInLayer == baseLayer {
					return fmt.Errorf("SkipList::Delete error: input element key not exist")
				} else {
					p = p.downPtr
				}
			}
		}
	}
	delete(sl.dict, e.v.Key())
	return nil
}

func (sl *SkipList[KeyType]) DeleteByKey(key KeyType) (ret error) {
	v, exist := sl.dict[key]
	if !exist {
		return fmt.Errorf("SkipList::DeleteByKey error: key %v not exist", key)
	}
	return sl.delete(v)
}

func (sl *SkipList[KeyType]) find(e ISkiplistElement[KeyType]) (val *SkipListNode[KeyType], err error) {
	v, exist := sl.dict[e.Key()]
	if !exist {
		return nil, fmt.Errorf("SkipList::Find error: key %v not exist", e.Key())
	}
	return v, nil
}

func (sl *SkipList[KeyType]) FindByKey(key KeyType) (val *SkipListNode[KeyType], err error) {
	v, exist := sl.dict[key]
	if !exist {
		return nil, fmt.Errorf("SkipList::FindByKey error: key %v not exist", key)
	}
	return v, nil
}

func (sl *SkipList[KeyType]) getRank(e ISkiplistElement[KeyType]) (ret int32, err error) {
	_, err = sl.find(e)
	if err != nil {
		return
	}
	head := sl.layersHead[sl.GetLayersCount()]
	for p := head; p != nil && p.compareElem(e) != 0; {
		if p.compareElem(e) < 0 {
			if p.frontPtr.compareElem(e) > 0 {
				if p.nodeInLayer == baseLayer {
					// 理论上不能走到这里
					panic("SkipList::GetRank error: input element not exist(but Find() passed)")
				}
				p = p.downPtr
			} else {
				ret += p.span
				p = p.frontPtr
			}
		}
	}
	return
}

func (sl *SkipList[KeyType]) GetRankByKey(key KeyType) (ret int32, err error) {
	v, err := sl.FindByKey(key)
	if err != nil {
		return
	}
	return sl.getRank(v.v)
}

func (sl *SkipList[KeyType]) GetReverseRank(e ISkiplistElement[KeyType]) (ret int32, err error) {
	ret, err = sl.getRank(e)
	if err != nil {
		return
	}
	ret = sl.GetElementsCount() + 1 - ret
	return
}

func (sl *SkipList[KeyType]) GetReverseRankByKey(key KeyType) (ret int32, err error) {
	v, err := sl.FindByKey(key)
	if err != nil {
		return
	}
	return sl.GetReverseRank(v.v)
}

func (sl *SkipList[KeyType]) getNodeByRank(rank int32) (ret *SkipListNode[KeyType], err error) {
	if rank <= 0 || rank > sl.GetElementsCount() {
		return nil, fmt.Errorf("SkipList::GetElementByRank error, input rank %d is out of range(total elements: %d)", rank, sl.GetElementsCount())
	}
	remainStep := rank
	p := sl.layersHead[sl.GetLayersCount()]
	for remainStep > 0 {
		if p.span > remainStep {
			if p.nodeInLayer == baseLayer {
				// 理论上不能走到这里（baseLayer的span必须都是1）
				panic(fmt.Sprintf("SkipList::GetElementByRank error, base layer node span = %d", p.span))
			}
			p = p.downPtr
		} else if p.span == remainStep {
			// found
			u := p.frontPtr
			for u.nodeInLayer != baseLayer {
				u = u.downPtr
			}
			return u, nil
		} else {
			remainStep -= p.span
			p = p.frontPtr
		}
	}
	return nil, fmt.Errorf("SkipList::GetElementByRank error, input rank %d not found", rank)
}

func (sl *SkipList[KeyType]) GetElementByRank(rank int32) (ret ISkiplistElement[KeyType], err error) {
	node, err := sl.getNodeByRank(rank)
	if err != nil {
		return
	}
	ret = node.v
	return
}

func (sl *SkipList[KeyType]) GetElementByReverseRank(rank int32) (ret ISkiplistElement[KeyType], err error) {
	return sl.GetElementByRank(sl.GetElementsCount() + 1 - rank)
}

func (sl *SkipList[KeyType]) GetRange(rankStart, rankEnd int32) (ret []ISkiplistElement[KeyType], err error) {
	if rankStart > rankEnd {
		rankStart, rankEnd = rankEnd, rankStart
	}
	if rankStart <= 0 || rankStart > sl.GetElementsCount() || rankEnd <= 0 || rankEnd > sl.GetElementsCount() {
		return nil, fmt.Errorf("SkipList::GetRange error, input rank %d or %d is out of range(total elements: %d)", rankStart, rankEnd, sl.GetElementsCount())
	}
	ret = make([]ISkiplistElement[KeyType], 0, rankEnd-rankStart+1)
	startNode, err := sl.getNodeByRank(rankStart)
	if err != nil {
		return
	}
	p := startNode
	for i := rankStart; i <= rankEnd; i++ {
		lp := p
		if lp.isTail {
			panic("SkipList::GetRange error, loop out of bounds(but rankEnd check passed)")
		}
		ret = append(ret, lp.v)
		p = p.frontPtr
	}
	return
}

func (sl *SkipList[KeyType]) GetReverseRange(rankRevStart, rankRevEnd int32) (ret []ISkiplistElement[KeyType], err error) {
	if rankRevStart > rankRevEnd {
		rankRevStart, rankRevEnd = rankRevEnd, rankRevStart
	}
	rankStart := sl.GetElementsCount() + 1 - rankRevEnd
	rankEnd := sl.GetElementsCount() + 1 - rankRevStart
	r, err := sl.GetRange(rankStart, rankEnd)
	if err != nil {
		return
	}
	ret = make([]ISkiplistElement[KeyType], 0, rankRevEnd-rankRevStart+1)
	for i := len(r) - 1; i >= 0; i-- {
		ret = append(ret, r[i])
	}
	return
}

// Print debug用 打印内部结构
func (sl *SkipList[KeyType]) Print() {
	var b bytes.Buffer
	for i := sl.GetLayersCount(); i >= baseLayer; i-- {
		for p := sl.layersHead[i]; p != nil; p = p.frontPtr {
			b.WriteString(fmt.Sprintf("[V:%v,lay:%d,span:%d,link:%d]\t", p.v, p.nodeInLayer, p.span, p.linkNodes))
			for i := int32(0); i < p.span-1; i++ {
				b.WriteString("\t\t\t\t\t\t\t\t")
			}
		}
		b.WriteString("\n")
	}
	fmt.Println(b.String())
}
