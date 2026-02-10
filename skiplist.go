package csort

import (
	"math/big"
	"sync"
)

// ScoreMember 表示一个分数-成员对
type ScoreMember struct {
	Score  *big.Rat // 使用 big.Rat 支持任意精度小数
	Member string
}

// skipNode 跳表节点
type skipNode struct {
	member   string
	score    *big.Rat
	forward  []*skipNode // 前向指针数组
	backward *skipNode   // 后向指针，用于反向遍历
	level    int
}

// SkipList 跳表实现
type SkipList struct {
	head     *skipNode
	tail     *skipNode
	length   int
	level    int
	maxLevel int
	p        float64 // 节点晋升概率
	mu       sync.RWMutex
}

// NewSkipList 创建新的跳表
func NewSkipList() *SkipList {
	maxLevel := 32
	return &SkipList{
		head:     &skipNode{forward: make([]*skipNode, maxLevel)},
		level:    1,
		maxLevel: maxLevel,
		p:        0.25,
	}
}

// randomLevel 随机生成节点层级
func (sl *SkipList) randomLevel() int {
	level := 1
	for level < sl.maxLevel && randFloat() < sl.p {
		level++
	}
	return level
}

// randFloat 简单的随机数生成
func randFloat() float64 {
	// 使用简单的伪随机数，避免导入 math/rand
	return float64(fastRand()%1000) / 1000.0
}

// fastRand xorshift 快速随机数生成
func fastRand() uint32 {
	staticSeed := uint32(1)
	staticSeed ^= staticSeed << 13
	staticSeed ^= staticSeed >> 17
	staticSeed ^= staticSeed << 5
	return staticSeed
}

// compare 比较两个分数
// 返回值: -1 表示 a < b, 0 表示 a == b, 1 表示 a > b
func compare(a, b *big.Rat) int {
	return a.Cmp(b)
}

// findNodeByMember 根据成员名查找节点（内部方法，无锁）
func (sl *SkipList) findNodeByMember(member string) *skipNode {
	node := sl.head.forward[0]
	for node != nil {
		if node.member == member {
			return node
		}
		node = node.forward[0]
	}
	return nil
}

// Insert 插入或更新元素
func (sl *SkipList) Insert(member string, score *big.Rat) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// 先检查成员是否已存在
	existingNode := sl.findNodeByMember(member)
	if existingNode != nil {
		// 分数相同，不需要更新
		if compare(existingNode.score, score) == 0 {
			return
		}
		// 分数不同，先删除旧节点
		update := make([]*skipNode, sl.maxLevel)
		node := sl.head
		for i := sl.level - 1; i >= 0; i-- {
			for node.forward[i] != nil && node.forward[i] != existingNode {
				node = node.forward[i]
			}
			update[i] = node
		}
		sl.deleteNode(existingNode, update)
	}

	// 查找新位置的插入点
	update := make([]*skipNode, sl.maxLevel)
	node := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil {
			cmp := compare(node.forward[i].score, score)
			if cmp < 0 || (cmp == 0 && node.forward[i].member < member) {
				node = node.forward[i]
			} else {
				break
			}
		}
		update[i] = node
	}

	// 生成新节点的层级
	newLevel := sl.randomLevel()
	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			update[i] = sl.head
		}
		sl.level = newLevel
	}

	// 创建新节点
	newNode := &skipNode{
		member:  member,
		score:   new(big.Rat).Set(score), // 复制分数
		forward: make([]*skipNode, newLevel),
		level:   newLevel,
	}

	// 更新指针
	for i := 0; i < newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	// 更新后向指针
	if update[0] != sl.head {
		newNode.backward = update[0]
	}
	if newNode.forward[0] != nil {
		newNode.forward[0].backward = newNode
	} else {
		sl.tail = newNode
	}

	sl.length++
}

// deleteNode 删除节点
func (sl *SkipList) deleteNode(node *skipNode, update []*skipNode) {
	for i := 0; i < node.level; i++ {
		update[i].forward[i] = node.forward[i]
	}

	if node.forward[0] != nil {
		node.forward[0].backward = node.backward
	} else {
		sl.tail = node.backward
	}

	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.level--
	}

	sl.length--
}

// Delete 删除指定成员
func (sl *SkipList) Delete(member string, score *big.Rat) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	update := make([]*skipNode, sl.maxLevel)
	node := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil {
			cmp := compare(node.forward[i].score, score)
			if cmp < 0 || (cmp == 0 && node.forward[i].member < member) {
				node = node.forward[i]
			} else {
				break
			}
		}
		update[i] = node
	}

	node = node.forward[0]
	if node != nil && node.member == member && compare(node.score, score) == 0 {
		sl.deleteNode(node, update)
		return true
	}
	return false
}

// GetRank 获取成员的排名（从1开始）
func (sl *SkipList) GetRank(member string, score *big.Rat) int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// 线性遍历查找排名
	rank := 1
	node := sl.head.forward[0]
	for node != nil {
		if node.member == member && compare(node.score, score) == 0 {
			return rank
		}
		node = node.forward[0]
		rank++
	}
	return 0
}

// GetByRank 根据排名获取成员
func (sl *SkipList) GetByRank(rank int) (string, *big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if rank < 1 || rank > sl.length {
		return "", nil, false
	}

	node := sl.head.forward[0]
	currentRank := 1

	for node != nil && currentRank < rank {
		node = node.forward[0]
		currentRank++
	}

	if node != nil {
		return node.member, new(big.Rat).Set(node.score), true
	}
	return "", nil, false
}

// Range 获取排名范围内的成员 [start, stop] 闭区间
func (sl *SkipList) Range(start, stop int, reverse bool) []ScoreMember {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if start < 1 {
		start = 1
	}
	if stop > sl.length {
		stop = sl.length
	}
	if start > stop {
		return nil
	}

	result := make([]ScoreMember, 0, stop-start+1)

	if reverse {
		// 反向遍历
		node := sl.tail
		currentRank := sl.length
		for node != nil && currentRank > stop {
			node = node.backward
			currentRank--
		}
		for node != nil && currentRank >= start {
			result = append(result, ScoreMember{
				Score:  new(big.Rat).Set(node.score),
				Member: node.member,
			})
			node = node.backward
			currentRank--
		}
	} else {
		// 正向遍历
		node := sl.head.forward[0]
		currentRank := 1
		for node != nil && currentRank < start {
			node = node.forward[0]
			currentRank++
		}
		for node != nil && currentRank <= stop {
			result = append(result, ScoreMember{
				Score:  new(big.Rat).Set(node.score),
				Member: node.member,
			})
			node = node.forward[0]
			currentRank++
		}
	}

	return result
}

// RangeByScore 根据分数范围获取成员
func (sl *SkipList) RangeByScore(min, max *big.Rat, reverse bool) []ScoreMember {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	result := make([]ScoreMember, 0)

	if reverse {
		// 反向遍历
		node := sl.tail
		for node != nil && compare(node.score, max) > 0 {
			node = node.backward
		}
		for node != nil && compare(node.score, min) >= 0 {
			result = append(result, ScoreMember{
				Score:  new(big.Rat).Set(node.score),
				Member: node.member,
			})
			node = node.backward
		}
	} else {
		// 正向遍历
		node := sl.head.forward[0]
		for node != nil && compare(node.score, min) < 0 {
			node = node.forward[0]
		}
		for node != nil && compare(node.score, max) <= 0 {
			result = append(result, ScoreMember{
				Score:  new(big.Rat).Set(node.score),
				Member: node.member,
			})
			node = node.forward[0]
		}
	}

	return result
}

// CountByScore 统计分数范围内的成员数量
func (sl *SkipList) CountByScore(min, max *big.Rat) int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	count := 0
	node := sl.head.forward[0]
	for node != nil && compare(node.score, min) < 0 {
		node = node.forward[0]
	}
	for node != nil && compare(node.score, max) <= 0 {
		count++
		node = node.forward[0]
	}
	return count
}

// RemoveByScore 删除分数范围内的所有成员
func (sl *SkipList) RemoveByScore(min, max *big.Rat) int {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	count := 0
	node := sl.head.forward[0]
	for node != nil && compare(node.score, min) < 0 {
		node = node.forward[0]
	}

	var toDelete []*skipNode
	for node != nil && compare(node.score, max) <= 0 {
		toDelete = append(toDelete, node)
		node = node.forward[0]
	}

	for _, n := range toDelete {
		update := make([]*skipNode, sl.maxLevel)
		current := sl.head
		for i := sl.level - 1; i >= 0; i-- {
			for current.forward[i] != nil && current.forward[i] != n {
				current = current.forward[i]
			}
			update[i] = current
		}
		sl.deleteNode(n, update)
		count++
	}

	return count
}

// RemoveByRank 删除排名范围内的所有成员
func (sl *SkipList) RemoveByRank(start, stop int) int {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if start < 1 {
		start = 1
	}
	if stop > sl.length {
		stop = sl.length
	}
	if start > stop {
		return 0
	}

	count := 0
	node := sl.head.forward[0]
	currentRank := 1

	for node != nil && currentRank < start {
		node = node.forward[0]
		currentRank++
	}

	for node != nil && currentRank <= stop {
		next := node.forward[0]
		update := make([]*skipNode, sl.maxLevel)
		current := sl.head
		for i := sl.level - 1; i >= 0; i-- {
			for current.forward[i] != nil && current.forward[i] != node {
				current = current.forward[i]
			}
			update[i] = current
		}
		sl.deleteNode(node, update)
		count++
		node = next
		currentRank++
	}

	return count
}

// Len 返回元素数量
func (sl *SkipList) Len() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
}

// GetScore 获取成员的分数
func (sl *SkipList) GetScore(member string) (*big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	node := sl.head.forward[0]
	for node != nil {
		if node.member == member {
			return new(big.Rat).Set(node.score), true
		}
		node = node.forward[0]
	}
	return nil, false
}

// GetPrevMember 获取前一位成员（分数更小，或分数相同但 member 字典序更小）
// 返回: prevMember, prevScore, exists
func (sl *SkipList) GetPrevMember(member string) (string, *big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	node := sl.head.forward[0]
	var prevNode *skipNode

	for node != nil {
		if node.member == member {
			// 找到了目标成员
			if prevNode != nil {
				return prevNode.member, new(big.Rat).Set(prevNode.score), true
			}
			return "", nil, false // 这是第一个成员，没有前一位
		}
		prevNode = node
		node = node.forward[0]
	}
	return "", nil, false // 成员不存在
}

// GetNextMember 获取后一位成员（分数更大，或分数相同但 member 字典序更大）
// 返回: nextMember, nextScore, exists
func (sl *SkipList) GetNextMember(member string) (string, *big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	node := sl.head.forward[0]

	for node != nil {
		if node.member == member {
			// 找到了目标成员
			if node.forward[0] != nil {
				next := node.forward[0]
				return next.member, new(big.Rat).Set(next.score), true
			}
			return "", nil, false // 这是最后一个成员，没有后一位
		}
		node = node.forward[0]
	}
	return "", nil, false // 成员不存在
}
func (sl *SkipList) InRankRange(member string, score *big.Rat, start, stop int) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	rank := 0
	node := sl.head
	found := false

	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil {
			if node.forward[i].member == member && compare(node.forward[i].score, score) == 0 {
				found = true
				break
			}
			node = node.forward[i]
			rank++
		}
		if found {
			break
		}
	}

	if !found {
		return false
	}
	return rank >= start && rank <= stop
}

// IncrementBy 增加成员的分数
func (sl *SkipList) IncrementBy(member string, increment *big.Rat) (*big.Rat, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// 查找成员
	node := sl.head.forward[0]
	for node != nil {
		if node.member == member {
			break
		}
		node = node.forward[0]
	}

	if node == nil {
		// 成员不存在，直接插入新成员
		newScore := new(big.Rat).Set(increment)
		sl.insertInternal(member, newScore)
		return newScore, true
	}

	// 保存旧分数并计算新分数
	newScore := new(big.Rat).Add(node.score, increment)

	// 删除旧节点
	update := make([]*skipNode, sl.maxLevel)
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i] != node {
			current = current.forward[i]
		}
		update[i] = current
	}
	sl.deleteNode(node, update)

	// 插入新节点
	sl.insertInternal(member, newScore)

	return newScore, true
}

// insertInternal 内部插入方法（无锁版本，调用者必须持有写锁）
func (sl *SkipList) insertInternal(member string, score *big.Rat) {
	// 先检查成员是否已存在
	existingNode := sl.findNodeByMember(member)
	if existingNode != nil {
		// 分数相同，不需要更新
		if compare(existingNode.score, score) == 0 {
			return
		}
		// 分数不同，先删除旧节点
		update := make([]*skipNode, sl.maxLevel)
		node := sl.head
		for i := sl.level - 1; i >= 0; i-- {
			for node.forward[i] != nil && node.forward[i] != existingNode {
				node = node.forward[i]
			}
			update[i] = node
		}
		sl.deleteNode(existingNode, update)
	}

	// 查找新位置的插入点
	update := make([]*skipNode, sl.maxLevel)
	node := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil {
			cmp := compare(node.forward[i].score, score)
			if cmp < 0 || (cmp == 0 && node.forward[i].member < member) {
				node = node.forward[i]
			} else {
				break
			}
		}
		update[i] = node
	}

	// 生成新节点的层级
	newLevel := sl.randomLevel()
	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			update[i] = sl.head
		}
		sl.level = newLevel
	}

	// 创建新节点
	newNode := &skipNode{
		member:  member,
		score:   new(big.Rat).Set(score),
		forward: make([]*skipNode, newLevel),
		level:   newLevel,
	}

	// 更新指针
	for i := 0; i < newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	// 更新后向指针
	if update[0] != sl.head {
		newNode.backward = update[0]
	}
	if newNode.forward[0] != nil {
		newNode.forward[0].backward = newNode
	} else {
		sl.tail = newNode
	}

	sl.length++
}

// All 获取所有成员（按分数排序）
func (sl *SkipList) All() []ScoreMember {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	result := make([]ScoreMember, 0, sl.length)
	node := sl.head.forward[0]
	for node != nil {
		result = append(result, ScoreMember{
			Score:  new(big.Rat).Set(node.score),
			Member: node.member,
		})
		node = node.forward[0]
	}
	return result
}

// Clear 清空跳表
func (sl *SkipList) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.head = &skipNode{forward: make([]*skipNode, sl.maxLevel)}
	sl.tail = nil
	sl.length = 0
	sl.level = 1
}
