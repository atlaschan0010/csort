package csort

import (
	"math/big"
	"math/rand/v2"
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
	span     []int       // 每层的跨度（用于 O(log n) 排名计算）
	backward *skipNode   // 后向指针，用于反向遍历
	level    int
}

// SkipList 跳表实现
type SkipList struct {
	head      *skipNode
	tail      *skipNode
	length    int
	level     int
	maxLevel  int
	p         float64              // 节点晋升概率
	memberMap map[string]*skipNode // member → node 索引（O(1) 查找）
	mu        sync.RWMutex
}

// NewSkipList 创建新的跳表
func NewSkipList() *SkipList {
	maxLevel := 32
	return &SkipList{
		head:      &skipNode{forward: make([]*skipNode, maxLevel), span: make([]int, maxLevel)},
		level:     1,
		maxLevel:  maxLevel,
		p:         0.25,
		memberMap: make(map[string]*skipNode),
	}
}

// randomLevel 随机生成节点层级
func (sl *SkipList) randomLevel() int {
	level := 1
	for level < sl.maxLevel && rand.Float64() < sl.p {
		level++
	}
	return level
}

// compare 比较两个分数
// 返回值: -1 表示 a < b, 0 表示 a == b, 1 表示 a > b
func compare(a, b *big.Rat) int {
	return a.Cmp(b)
}

// Insert 插入或更新元素
func (sl *SkipList) Insert(member string, score *big.Rat) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.insertInternal(member, score)
}

// insertInternal 内部插入方法（无锁版本，调用者必须持有写锁）
func (sl *SkipList) insertInternal(member string, score *big.Rat) {
	// 检查成员是否已存在
	if existingNode, exists := sl.memberMap[member]; exists {
		// 分数相同，不需要更新
		if compare(existingNode.score, score) == 0 {
			return
		}
		// 分数不同，先删除旧节点
		sl.deleteByNode(existingNode)
	}

	// 查找插入位置并记录每层的 update 节点和 rank
	update := make([]*skipNode, sl.maxLevel)
	rank := make([]int, sl.maxLevel)
	node := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		if i < sl.level-1 {
			rank[i] = rank[i+1]
		}
		for node.forward[i] != nil {
			cmp := compare(node.forward[i].score, score)
			if cmp < 0 || (cmp == 0 && node.forward[i].member < member) {
				rank[i] += node.span[i]
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
			rank[i] = 0
			update[i] = sl.head
			update[i].span[i] = sl.length
		}
		sl.level = newLevel
	}

	// 创建新节点
	newNode := &skipNode{
		member:  member,
		score:   new(big.Rat).Set(score), // 复制分数
		forward: make([]*skipNode, newLevel),
		span:    make([]int, newLevel),
		level:   newLevel,
	}

	// 更新指针和跨度
	for i := 0; i < newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode

		// 计算新节点在第 i 层的跨度
		newNode.span[i] = update[i].span[i] - (rank[0] - rank[i])
		update[i].span[i] = rank[0] - rank[i] + 1
	}

	// 更新未涉及层的跨度（增量 +1）
	for i := newLevel; i < sl.level; i++ {
		update[i].span[i]++
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
	sl.memberMap[member] = newNode
}

// deleteByNode 通过节点指针删除（内部方法，调用者必须持有写锁）
func (sl *SkipList) deleteByNode(target *skipNode) {
	update := make([]*skipNode, sl.maxLevel)
	node := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil {
			if node.forward[i] == target {
				break
			}
			cmp := compare(node.forward[i].score, target.score)
			if cmp < 0 || (cmp == 0 && node.forward[i].member < target.member) {
				node = node.forward[i]
			} else {
				break
			}
		}
		update[i] = node
	}

	sl.deleteNode(target, update)
}

// deleteNode 删除节点并更新指针和跨度
func (sl *SkipList) deleteNode(node *skipNode, update []*skipNode) {
	for i := 0; i < node.level; i++ {
		if update[i].forward[i] == node {
			update[i].span[i] += node.span[i] - 1
			update[i].forward[i] = node.forward[i]
		} else {
			update[i].span[i]--
		}
	}

	if node.forward[0] != nil {
		node.forward[0].backward = node.backward
	} else {
		sl.tail = node.backward
	}

	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.level--
	}

	delete(sl.memberMap, node.member)
	sl.length--
}

// Delete 删除指定成员
func (sl *SkipList) Delete(member string, score *big.Rat) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	node, exists := sl.memberMap[member]
	if !exists {
		return false
	}

	// 验证 score 一致性
	if compare(node.score, score) != 0 {
		return false
	}

	sl.deleteByNode(node)
	return true
}

// DeleteByMember 仅根据 member 名称删除（不需要 score）
func (sl *SkipList) DeleteByMember(member string) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	node, exists := sl.memberMap[member]
	if !exists {
		return false
	}

	sl.deleteByNode(node)
	return true
}

// GetRank 获取成员的排名（从1开始）— O(log n) 通过 span 计算
func (sl *SkipList) GetRank(member string, score *big.Rat) int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	rank := 0
	node := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil {
			cmp := compare(node.forward[i].score, score)
			if cmp < 0 || (cmp == 0 && node.forward[i].member <= member) {
				rank += node.span[i]
				node = node.forward[i]
				if node.member == member {
					return rank
				}
			} else {
				break
			}
		}
	}

	return 0 // 未找到
}

// GetByRank 根据排名获取成员 — O(log n) 通过 span 定位
func (sl *SkipList) GetByRank(rank int) (string, *big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if rank < 1 || rank > sl.length {
		return "", nil, false
	}

	node := sl.head
	traversed := 0
	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil && traversed+node.span[i] <= rank {
			traversed += node.span[i]
			node = node.forward[i]
		}
		if traversed == rank {
			return node.member, new(big.Rat).Set(node.score), true
		}
	}

	return "", nil, false
}

// GetScore 获取成员的分数 — O(1) 通过 memberMap
func (sl *SkipList) GetScore(member string) (*big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	node, exists := sl.memberMap[member]
	if !exists {
		return nil, false
	}
	return new(big.Rat).Set(node.score), true
}

// GetPrevMember 获取前一位成员（分数更小，或分数相同但 member 字典序更小）
func (sl *SkipList) GetPrevMember(member string) (string, *big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	node, exists := sl.memberMap[member]
	if !exists {
		return "", nil, false
	}

	if node.backward != nil {
		return node.backward.member, new(big.Rat).Set(node.backward.score), true
	}
	return "", nil, false
}

// GetNextMember 获取后一位成员（分数更大，或分数相同但 member 字典序更大）
func (sl *SkipList) GetNextMember(member string) (string, *big.Rat, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	node, exists := sl.memberMap[member]
	if !exists {
		return "", nil, false
	}

	if node.forward[0] != nil {
		next := node.forward[0]
		return next.member, new(big.Rat).Set(next.score), true
	}
	return "", nil, false
}

// Range 获取排名范围内的成员 [start, stop] 闭区间（1-based）
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
		// 反向：从 stop 位置开始，向 backward 方向遍历到 start
		node := sl.getNodeByRankInternal(stop)
		count := stop - start + 1
		for node != nil && count > 0 {
			result = append(result, ScoreMember{
				Score:  new(big.Rat).Set(node.score),
				Member: node.member,
			})
			node = node.backward
			count--
		}
	} else {
		// 正向：定位到 start 位置
		node := sl.getNodeByRankInternal(start)
		for node != nil && start <= stop {
			result = append(result, ScoreMember{
				Score:  new(big.Rat).Set(node.score),
				Member: node.member,
			})
			node = node.forward[0]
			start++
		}
	}

	return result
}

// getNodeByRankInternal 根据排名获取节点（内部方法，无锁，O(log n)）
func (sl *SkipList) getNodeByRankInternal(rank int) *skipNode {
	if rank < 1 || rank > sl.length {
		return nil
	}

	node := sl.head
	traversed := 0
	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil && traversed+node.span[i] <= rank {
			traversed += node.span[i]
			node = node.forward[i]
		}
		if traversed == rank {
			return node
		}
	}
	return nil
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
		// 正向遍历：利用跳表快速定位到 >= min 的第一个节点
		node := sl.head
		for i := sl.level - 1; i >= 0; i-- {
			for node.forward[i] != nil && compare(node.forward[i].score, min) < 0 {
				node = node.forward[i]
			}
		}
		node = node.forward[0]

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
	// 利用跳表快速定位
	node := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil && compare(node.forward[i].score, min) < 0 {
			node = node.forward[i]
		}
	}
	node = node.forward[0]

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

	// 收集要删除的节点
	var toDelete []*skipNode
	node := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil && compare(node.forward[i].score, min) < 0 {
			node = node.forward[i]
		}
	}
	node = node.forward[0]

	for node != nil && compare(node.score, max) <= 0 {
		toDelete = append(toDelete, node)
		node = node.forward[0]
	}

	for _, n := range toDelete {
		sl.deleteByNode(n)
	}

	return len(toDelete)
}

// RemoveByRank 删除排名范围内的所有成员 [start, stop] 1-based
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

	// 定位到 start 位置
	node := sl.getNodeByRankInternal(start)

	count := 0
	for node != nil && start+count <= stop {
		next := node.forward[0]
		sl.deleteByNode(node)
		count++
		node = next
	}

	return count
}

// InRankRange 检查成员是否在指定排名范围内
func (sl *SkipList) InRankRange(member string, score *big.Rat, start, stop int) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	rank := 0
	node := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for node.forward[i] != nil {
			cmp := compare(node.forward[i].score, score)
			if cmp < 0 || (cmp == 0 && node.forward[i].member <= member) {
				rank += node.span[i]
				node = node.forward[i]
				if node.member == member {
					return rank >= start && rank <= stop
				}
			} else {
				break
			}
		}
	}

	return false
}

// IncrementBy 增加成员的分数
func (sl *SkipList) IncrementBy(member string, increment *big.Rat) (*big.Rat, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	existingNode, exists := sl.memberMap[member]
	var newScore *big.Rat

	if !exists {
		// 成员不存在，直接插入
		newScore = new(big.Rat).Set(increment)
	} else {
		// 计算新分数
		newScore = new(big.Rat).Add(existingNode.score, increment)
		// 删除旧节点
		sl.deleteByNode(existingNode)
	}

	// 插入新节点
	sl.insertInternal(member, newScore)
	return new(big.Rat).Set(newScore), true
}

// Len 返回元素数量
func (sl *SkipList) Len() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
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

	sl.head = &skipNode{forward: make([]*skipNode, sl.maxLevel), span: make([]int, sl.maxLevel)}
	sl.tail = nil
	sl.length = 0
	sl.level = 1
	sl.memberMap = make(map[string]*skipNode)
}
