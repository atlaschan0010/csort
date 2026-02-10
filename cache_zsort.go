package csort

import (
	"math/big"
	"sync"
)

// ZSet 表示一个有序集合
type ZSet struct {
	sl *SkipList
	mu sync.RWMutex
}

// newZSet 创建新的有序集合
func newZSet() *ZSet {
	return &ZSet{
		sl: NewSkipList(),
	}
}

// CacheZSort 内存排序组件主结构
type CacheZSort struct {
	sets map[string]*ZSet
	mu   sync.RWMutex
}

// New 创建新的 CacheZSort 实例
func New() *CacheZSort {
	return &CacheZSort{
		sets: make(map[string]*ZSet),
	}
}

// getOrCreateZSet 获取或创建指定的 ZSet
func (c *CacheZSort) getOrCreateZSet(key string) *ZSet {
	c.mu.RLock()
	if set, ok := c.sets[key]; ok {
		c.mu.RUnlock()
		return set
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if set, ok := c.sets[key]; ok {
		return set
	}

	set := newZSet()
	c.sets[key] = set
	return set
}

// getZSet 获取指定的 ZSet，如果不存在返回 nil
func (c *CacheZSort) getZSet(key string) *ZSet {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sets[key]
}

// delZSet 删除指定的 ZSet
func (c *CacheZSort) delZSet(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sets, key)
}

// ==================== ZAdd ====================

// ZAdd 添加成员到有序集合
func (c *CacheZSort) ZAdd(key, member string, score *big.Rat) bool {
	set := c.getOrCreateZSet(key)
	set.mu.Lock()
	defer set.mu.Unlock()
	set.sl.Insert(member, score)
	return true
}

// ZAddString 添加成员（分数为字符串格式）
func (c *CacheZSort) ZAddString(key, member, scoreStr string) (bool, error) {
	score := new(big.Rat)
	if _, ok := score.SetString(scoreStr); !ok {
		return false, ErrInvalidScore
	}
	return c.ZAdd(key, member, score), nil
}

// ZAddFloat64 添加成员（分数为 float64）
func (c *CacheZSort) ZAddFloat64(key, member string, score float64) bool {
	return c.ZAdd(key, member, new(big.Rat).SetFloat64(score))
}

// ZAddInt64 添加成员（分数为 int64）
func (c *CacheZSort) ZAddInt64(key, member string, score int64) bool {
	return c.ZAdd(key, member, new(big.Rat).SetInt64(score))
}

// ZAddMultiple 添加多个成员
func (c *CacheZSort) ZAddMultiple(key string, members map[string]*big.Rat) int {
	set := c.getOrCreateZSet(key)
	set.mu.Lock()
	defer set.mu.Unlock()

	count := 0
	for member, score := range members {
		set.sl.Insert(member, score)
		count++
	}
	return count
}

// ==================== ZRem ====================

// ZRem 删除成员
func (c *CacheZSort) ZRem(key, member string) bool {
	set := c.getZSet(key)
	if set == nil {
		return false
	}

	score, ok := set.sl.GetScore(member)
	if !ok {
		return false
	}

	set.mu.Lock()
	defer set.mu.Unlock()
	return set.sl.Delete(member, score)
}

// ZRemMultiple 删除多个成员
func (c *CacheZSort) ZRemMultiple(key string, members []string) int {
	set := c.getZSet(key)
	if set == nil {
		return 0
	}

	count := 0
	for _, member := range members {
		if score, ok := set.sl.GetScore(member); ok {
			if set.sl.Delete(member, score) {
				count++
			}
		}
	}
	return count
}

// ==================== ZScore ====================

// ZScore 获取成员的分数
func (c *CacheZSort) ZScore(key, member string) (*big.Rat, bool) {
	set := c.getZSet(key)
	if set == nil {
		return nil, false
	}
	return set.sl.GetScore(member)
}

// ZScoreString 获取成员的分数（字符串格式）
func (c *CacheZSort) ZScoreString(key, member string) (string, bool) {
	score, ok := c.ZScore(key, member)
	if !ok {
		return "", false
	}
	return score.FloatString(20), true // 默认返回20位小数
}

// ==================== ZRank ====================

// ZRank 获取成员的正序排名（从0开始）
func (c *CacheZSort) ZRank(key, member string) (int, bool) {
	set := c.getZSet(key)
	if set == nil {
		return -1, false
	}

	score, ok := set.sl.GetScore(member)
	if !ok {
		return -1, false
	}

	rank := set.sl.GetRank(member, score)
	if rank == 0 {
		return -1, false
	}
	return rank - 1, true // 转换为从0开始
}

// ZRevRank 获取成员的倒序排名（从0开始）
func (c *CacheZSort) ZRevRank(key, member string) (int, bool) {
	rank, ok := c.ZRank(key, member)
	if !ok {
		return -1, false
	}

	card, _ := c.ZCard(key)
	return card - 1 - rank, true
}

// GetMemberRank 根据 member 查询排名（从1开始）
// 这是 ZRank 的别名，返回 1-based 排名
func (c *CacheZSort) GetMemberRank(key, member string) (int, bool) {
	set := c.getZSet(key)
	if set == nil {
		return 0, false
	}

	score, ok := set.sl.GetScore(member)
	if !ok {
		return 0, false
	}

	rank := set.sl.GetRank(member, score)
	return rank, rank > 0
}

// GetPrevMember 根据 member 查询前一位成员
// 返回: prevMember, prevScore, exists
func (c *CacheZSort) GetPrevMember(key, member string) (string, *big.Rat, bool) {
	set := c.getZSet(key)
	if set == nil {
		return "", nil, false
	}
	return set.sl.GetPrevMember(member)
}

// GetNextMember 根据 member 查询后一位成员
// 返回: nextMember, nextScore, exists
func (c *CacheZSort) GetNextMember(key, member string) (string, *big.Rat, bool) {
	set := c.getZSet(key)
	if set == nil {
		return "", nil, false
	}
	return set.sl.GetNextMember(member)
}

// GetPrevMemberString 根据 member 查询前一位成员（分数为字符串格式）
// 返回: prevMember, prevScoreStr, exists
func (c *CacheZSort) GetPrevMemberString(key, member string) (string, string, bool) {
	prevMember, prevScore, ok := c.GetPrevMember(key, member)
	if !ok {
		return "", "", false
	}
	return prevMember, prevScore.FloatString(20), true
}

// GetNextMemberString 根据 member 查询后一位成员（分数为字符串格式）
// 返回: nextMember, nextScoreStr, exists
func (c *CacheZSort) GetNextMemberString(key, member string) (string, string, bool) {
	nextMember, nextScore, ok := c.GetNextMember(key, member)
	if !ok {
		return "", "", false
	}
	return nextMember, nextScore.FloatString(20), true
}

// ==================== ZRange ====================

// ZRange 获取指定排名范围的成员（正序，从0开始，闭区间）
func (c *CacheZSort) ZRange(key string, start, stop int, withScores bool) []interface{} {
	set := c.getZSet(key)
	if set == nil {
		return nil
	}

	card := set.sl.Len()
	if card == 0 {
		return nil
	}

	// 处理负数索引
	if start < 0 {
		start = card + start
	}
	if stop < 0 {
		stop = card + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= card {
		stop = card - 1
	}
	if start > stop {
		return nil
	}

	// 转换为1-based索引
	result := set.sl.Range(start+1, stop+1, false)

	if withScores {
		output := make([]interface{}, 0, len(result)*2)
		for _, sm := range result {
			output = append(output, sm.Member, sm.Score.FloatString(20))
		}
		return output
	}

	output := make([]interface{}, 0, len(result))
	for _, sm := range result {
		output = append(output, sm.Member)
	}
	return output
}

// ZRevRange 获取指定排名范围的成员（倒序，从0开始，闭区间）
func (c *CacheZSort) ZRevRange(key string, start, stop int, withScores bool) []interface{} {
	set := c.getZSet(key)
	if set == nil {
		return nil
	}

	card := set.sl.Len()
	if card == 0 {
		return nil
	}

	// 处理负数索引
	if start < 0 {
		start = card + start
	}
	if stop < 0 {
		stop = card + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= card {
		stop = card - 1
	}
	if start > stop {
		return nil
	}

	// 转换为1-based索引并反转
	result := set.sl.Range(start+1, stop+1, true)

	if withScores {
		output := make([]interface{}, 0, len(result)*2)
		for _, sm := range result {
			output = append(output, sm.Member, sm.Score.FloatString(20))
		}
		return output
	}

	output := make([]interface{}, 0, len(result))
	for _, sm := range result {
		output = append(output, sm.Member)
	}
	return output
}

// ==================== ZRangeByScore ====================

// ZRangeByScore 根据分数范围获取成员（正序，闭区间）
func (c *CacheZSort) ZRangeByScore(key string, min, max *big.Rat, withScores bool, offset, count int) []interface{} {
	set := c.getZSet(key)
	if set == nil {
		return nil
	}

	result := set.sl.RangeByScore(min, max, false)

	// 应用 offset 和 count
	if offset >= len(result) {
		return nil
	}
	end := offset + count
	if count <= 0 || end > len(result) {
		end = len(result)
	}
	result = result[offset:end]

	if withScores {
		output := make([]interface{}, 0, len(result)*2)
		for _, sm := range result {
			output = append(output, sm.Member, sm.Score.FloatString(20))
		}
		return output
	}

	output := make([]interface{}, 0, len(result))
	for _, sm := range result {
		output = append(output, sm.Member)
	}
	return output
}

// ZRevRangeByScore 根据分数范围获取成员（倒序，闭区间）
func (c *CacheZSort) ZRevRangeByScore(key string, max, min *big.Rat, withScores bool, offset, count int) []interface{} {
	set := c.getZSet(key)
	if set == nil {
		return nil
	}

	result := set.sl.RangeByScore(min, max, true)

	// 应用 offset 和 count
	if offset >= len(result) {
		return nil
	}
	end := offset + count
	if count <= 0 || end > len(result) {
		end = len(result)
	}
	result = result[offset:end]

	if withScores {
		output := make([]interface{}, 0, len(result)*2)
		for _, sm := range result {
			output = append(output, sm.Member, sm.Score.FloatString(20))
		}
		return output
	}

	output := make([]interface{}, 0, len(result))
	for _, sm := range result {
		output = append(output, sm.Member)
	}
	return output
}

// ==================== ZCard ====================

// ZCard 获取有序集合的成员数量
func (c *CacheZSort) ZCard(key string) (int, bool) {
	set := c.getZSet(key)
	if set == nil {
		return 0, false
	}
	return set.sl.Len(), true
}

// ==================== ZCount ====================

// ZCount 统计分数范围内的成员数量
func (c *CacheZSort) ZCount(key string, min, max *big.Rat) int {
	set := c.getZSet(key)
	if set == nil {
		return 0
	}
	return set.sl.CountByScore(min, max)
}

// ==================== ZRemRangeByRank ====================

// ZRemRangeByRank 删除指定排名范围的成员
func (c *CacheZSort) ZRemRangeByRank(key string, start, stop int) int {
	set := c.getZSet(key)
	if set == nil {
		return 0
	}

	card := set.sl.Len()
	if card == 0 {
		return 0
	}

	// 处理负数索引
	if start < 0 {
		start = card + start
	}
	if stop < 0 {
		stop = card + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= card {
		stop = card - 1
	}
	if start > stop {
		return 0
	}

	return set.sl.RemoveByRank(start+1, stop+1)
}

// ==================== ZRemRangeByScore ====================

// ZRemRangeByScore 删除指定分数范围的成员
func (c *CacheZSort) ZRemRangeByScore(key string, min, max *big.Rat) int {
	set := c.getZSet(key)
	if set == nil {
		return 0
	}
	return set.sl.RemoveByScore(min, max)
}

// ==================== ZIncrBy ====================

// ZIncrBy 增加成员的分数
func (c *CacheZSort) ZIncrBy(key, member string, increment *big.Rat) (string, bool) {
	set := c.getOrCreateZSet(key)
	newScore, ok := set.sl.IncrementBy(member, increment)
	if !ok {
		return "", false
	}
	return newScore.FloatString(20), true
}

// ==================== Del ====================

// Del 删除整个有序集合
func (c *CacheZSort) Del(keys ...string) int {
	count := 0
	for _, key := range keys {
		if _, ok := c.sets[key]; ok {
			c.delZSet(key)
			count++
		}
	}
	return count
}

// ==================== Exists ====================

// Exists 检查有序集合是否存在
func (c *CacheZSort) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.sets[key]
	return ok
}

// ==================== Keys ====================

// Keys 获取所有有序集合的 key
func (c *CacheZSort) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.sets))
	for key := range c.sets {
		keys = append(keys, key)
	}
	return keys
}

// ==================== Flush ====================

// Flush 清空所有有序集合
func (c *CacheZSort) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sets = make(map[string]*ZSet)
}

// ==================== ZPopMin ====================

// ZPopMin 弹出分数最低的成员
func (c *CacheZSort) ZPopMin(key string, count int) []ScoreMember {
	set := c.getZSet(key)
	if set == nil {
		return nil
	}

	if count <= 0 {
		return nil
	}

	card := set.sl.Len()
	if count > card {
		count = card
	}

	result := set.sl.Range(1, count, false)
	set.sl.RemoveByRank(1, count)

	return result
}

// ==================== ZPopMax ====================

// ZPopMax 弹出分数最高的成员
func (c *CacheZSort) ZPopMax(key string, count int) []ScoreMember {
	set := c.getZSet(key)
	if set == nil {
		return nil
	}

	if count <= 0 {
		return nil
	}

	card := set.sl.Len()
	if count > card {
		count = card
	}

	start := card - count + 1
	result := set.sl.Range(start, card, true)
	set.sl.RemoveByRank(start, card)

	return result
}
