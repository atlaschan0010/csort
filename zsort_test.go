package csort

import (
	"math/big"
	"testing"
)

// TestNew 测试创建实例
func TestNew(t *testing.T) {
	cache := New()
	if cache == nil {
		t.Fatal("New() returned nil")
	}
	if cache.sets == nil {
		t.Fatal("sets map is nil")
	}
}

// TestZAddAndZScore 测试添加和获取分数
func TestZAddAndZScore(t *testing.T) {
	cache := New()

	// 测试高精度小数
	score := new(big.Rat)
	score.SetString("12345678901234567890.12345678901234567890")

	cache.ZAdd("test", "member1", score)

	got, ok := cache.ZScore("test", "member1")
	if !ok {
		t.Fatal("ZScore failed to find member")
	}

	if got.Cmp(score) != 0 {
		t.Errorf("ZScore = %v, want %v", got.FloatString(20), score.FloatString(20))
	}
}

// TestZAddString 测试字符串分数
func TestZAddString(t *testing.T) {
	cache := New()

	ok, err := cache.ZAddString("test", "member1", "3.14159265358979323846")
	if !ok || err != nil {
		t.Fatalf("ZAddString failed: ok=%v, err=%v", ok, err)
	}

	score, ok := cache.ZScoreString("test", "member1")
	if !ok {
		t.Fatal("ZScoreString failed")
	}

	// 验证前几位
	if score[:2] != "3." {
		t.Errorf("score doesn't start with '3.': %s", score)
	}
}

// TestZRank 测试排名
func TestZRank(t *testing.T) {
	cache := New()

	// 添加多个成员
	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	rank, ok := cache.ZRank("test", "a")
	if !ok || rank != 0 {
		t.Errorf("ZRank(a) = %d, want 0", rank)
	}

	rank, ok = cache.ZRank("test", "b")
	if !ok || rank != 1 {
		t.Errorf("ZRank(b) = %d, want 1", rank)
	}

	rank, ok = cache.ZRank("test", "c")
	if !ok || rank != 2 {
		t.Errorf("ZRank(c) = %d, want 2", rank)
	}
}

// TestZRevRank 测试倒序排名
func TestZRevRank(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	rank, ok := cache.ZRevRank("test", "c")
	if !ok || rank != 0 {
		t.Errorf("ZRevRank(c) = %d, want 0", rank)
	}

	rank, ok = cache.ZRevRank("test", "b")
	if !ok || rank != 1 {
		t.Errorf("ZRevRank(b) = %d, want 1", rank)
	}
}

// TestZRange 测试范围查询
func TestZRange(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)
	cache.ZAddFloat64("test", "d", 40)

	// 测试不带分数
	result := cache.ZRange("test", 0, 2, false)
	if len(result) != 3 {
		t.Errorf("ZRange returned %d items, want 3", len(result))
	}

	// 测试带分数
	result = cache.ZRange("test", 0, 1, true)
	if len(result) != 4 {
		t.Errorf("ZRange with scores returned %d items, want 4", len(result))
	}
}

// TestZRangeNegativeIndices 测试负数索引
func TestZRangeNegativeIndices(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	result := cache.ZRange("test", -2, -1, false)
	if len(result) != 2 {
		t.Errorf("ZRange with negative indices returned %d items, want 2", len(result))
	}
}

// TestZRem 测试删除
func TestZRem(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)

	if !cache.ZRem("test", "a") {
		t.Error("ZRem failed to remove existing member")
	}

	if cache.ZRem("test", "a") {
		t.Error("ZRem should fail for non-existing member")
	}

	card, _ := cache.ZCard("test")
	if card != 1 {
		t.Errorf("ZCard = %d, want 1", card)
	}
}

// TestZIncrBy 测试分数增加
func TestZIncrBy(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)

	increment := big.NewRat(5, 1)
	newScore, ok := cache.ZIncrBy("test", "a", increment)
	if !ok {
		t.Fatal("ZIncrBy failed")
	}

	if newScore[:2] != "15" {
		t.Errorf("ZIncrBy result = %s, want starting with 15", newScore)
	}
}

// TestZCount 测试分数范围计数
func TestZCount(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)
	cache.ZAddFloat64("test", "d", 40)

	min := big.NewRat(15, 1)
	max := big.NewRat(35, 1)

	count := cache.ZCount("test", min, max)
	if count != 2 {
		t.Errorf("ZCount = %d, want 2", count)
	}
}

// TestZRemRangeByScore 测试按分数范围删除
func TestZRemRangeByScore(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	min := big.NewRat(15, 1)
	max := big.NewRat(25, 1)

	removed := cache.ZRemRangeByScore("test", min, max)
	if removed != 1 {
		t.Errorf("ZRemRangeByScore removed %d, want 1", removed)
	}

	card, _ := cache.ZCard("test")
	if card != 2 {
		t.Errorf("ZCard after remove = %d, want 2", card)
	}
}

// TestMultipleKeys 测试多 key
func TestMultipleKeys(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("key1", "a", 10)
	cache.ZAddFloat64("key2", "b", 20)

	card1, _ := cache.ZCard("key1")
	card2, _ := cache.ZCard("key2")

	if card1 != 1 || card2 != 1 {
		t.Errorf("Cards: key1=%d, key2=%d, want both 1", card1, card2)
	}

	keys := cache.Keys()
	if len(keys) != 2 {
		t.Errorf("Keys count = %d, want 2", len(keys))
	}
}

// TestZPopMin 测试弹出最小
func TestZPopMin(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	popped := cache.ZPopMin("test", 2)
	if len(popped) != 2 {
		t.Errorf("ZPopMin returned %d items, want 2", len(popped))
	}

	if popped[0].Member != "a" {
		t.Errorf("First popped = %s, want a", popped[0].Member)
	}

	card, _ := cache.ZCard("test")
	if card != 1 {
		t.Errorf("ZCard after pop = %d, want 1", card)
	}
}

// TestZPopMax 测试弹出最大
func TestZPopMax(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	popped := cache.ZPopMax("test", 2)
	if len(popped) != 2 {
		t.Errorf("ZPopMax returned %d items, want 2", len(popped))
	}

	if popped[0].Member != "c" {
		t.Errorf("First popped = %s, want c", popped[0].Member)
	}
}

// TestHighPrecision 测试高精度小数
func TestHighPrecision(t *testing.T) {
	cache := New()

	// 测试 Redis 无法精确存储的分数
	highPrecision := "0.1234567890123456789012345678901234567890"

	ok, err := cache.ZAddString("test", "member", highPrecision)
	if !ok || err != nil {
		t.Fatalf("ZAddString failed: %v", err)
	}

	score, ok := cache.ZScore("test", "member")
	if !ok {
		t.Fatal("ZScore failed")
	}

	// 解析回 rat 比较
	expected := new(big.Rat)
	expected.SetString(highPrecision)

	if score.Cmp(expected) != 0 {
		t.Errorf("High precision mismatch:\ngot:      %s\nexpected: %s", score.FloatString(40), highPrecision)
	}
}

// TestUpdateScore 测试更新分数
func TestUpdateScore(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "a", 20) // 更新分数

	score, ok := cache.ZScore("test", "a")
	if !ok {
		t.Fatal("ZScore failed")
	}

	if score.Cmp(big.NewRat(20, 1)) != 0 {
		t.Errorf("Score = %v, want 20", score)
	}

	// 排名应该更新
	rank, ok := cache.ZRank("test", "a")
	if !ok || rank != 0 {
		t.Errorf("ZRank = %d, want 0", rank)
	}
}

// TestEmptyKey 测试空 key 操作
func TestEmptyKey(t *testing.T) {
	cache := New()

	if cache.Exists("nonexistent") {
		t.Error("Exists should return false for non-existent key")
	}

	result := cache.ZRange("nonexistent", 0, 10, false)
	if result != nil {
		t.Error("ZRange on non-existent key should return nil")
	}

	card, ok := cache.ZCard("nonexistent")
	if ok {
		t.Error("ZCard on non-existent key should return ok=false")
	}
	if card != 0 {
		t.Errorf("ZCard = %d, want 0", card)
	}
}

// TestGetMemberRank 测试根据 member 查询排名
func TestGetMemberRank(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	// 测试 GetMemberRank（1-based）
	rank, ok := cache.GetMemberRank("test", "a")
	if !ok || rank != 1 {
		t.Errorf("GetMemberRank(a) = %d, want 1", rank)
	}

	rank, ok = cache.GetMemberRank("test", "b")
	if !ok || rank != 2 {
		t.Errorf("GetMemberRank(b) = %d, want 2", rank)
	}

	rank, ok = cache.GetMemberRank("test", "c")
	if !ok || rank != 3 {
		t.Errorf("GetMemberRank(c) = %d, want 3", rank)
	}

	// 测试不存在的 member
	rank, ok = cache.GetMemberRank("test", "nonexistent")
	if ok {
		t.Error("GetMemberRank should return false for non-existent member")
	}

	// 测试不存在的 key
	rank, ok = cache.GetMemberRank("nonexistent", "a")
	if ok {
		t.Error("GetMemberRank should return false for non-existent key")
	}
}

// TestGetPrevMember 测试查询前一位成员
func TestGetPrevMember(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	// 查询 b 的前一位
	prevMember, prevScore, ok := cache.GetPrevMember("test", "b")
	if !ok {
		t.Fatal("GetPrevMember(b) should return true")
	}
	if prevMember != "a" {
		t.Errorf("GetPrevMember(b) member = %s, want a", prevMember)
	}
	if prevScore.Cmp(big.NewRat(10, 1)) != 0 {
		t.Errorf("GetPrevMember(b) score = %v, want 10", prevScore)
	}

	// 查询 c 的前一位
	prevMember, _, ok = cache.GetPrevMember("test", "c")
	if !ok || prevMember != "b" {
		t.Errorf("GetPrevMember(c) = %s, want b", prevMember)
	}

	// 查询 a 的前一位（应该不存在）
	_, _, ok = cache.GetPrevMember("test", "a")
	if ok {
		t.Error("GetPrevMember(a) should return false (first member)")
	}

	// 查询不存在的 member
	_, _, ok = cache.GetPrevMember("test", "nonexistent")
	if ok {
		t.Error("GetPrevMember(nonexistent) should return false")
	}

	// 测试字符串版本
	prevMember, prevScoreStr, ok := cache.GetPrevMemberString("test", "c")
	if !ok || prevMember != "b" {
		t.Errorf("GetPrevMemberString failed")
	}
	if prevScoreStr[:2] != "20" {
		t.Errorf("GetPrevMemberString score = %s, want starting with 20", prevScoreStr)
	}
}

// TestGetNextMember 测试查询后一位成员
func TestGetNextMember(t *testing.T) {
	cache := New()

	cache.ZAddFloat64("test", "a", 10)
	cache.ZAddFloat64("test", "b", 20)
	cache.ZAddFloat64("test", "c", 30)

	// 查询 a 的后一位
	nextMember, nextScore, ok := cache.GetNextMember("test", "a")
	if !ok {
		t.Fatal("GetNextMember(a) should return true")
	}
	if nextMember != "b" {
		t.Errorf("GetNextMember(a) member = %s, want b", nextMember)
	}
	if nextScore.Cmp(big.NewRat(20, 1)) != 0 {
		t.Errorf("GetNextMember(a) score = %v, want 20", nextScore)
	}

	// 查询 b 的后一位
	nextMember, _, ok = cache.GetNextMember("test", "b")
	if !ok || nextMember != "c" {
		t.Errorf("GetNextMember(b) = %s, want c", nextMember)
	}

	// 查询 c 的后一位（应该不存在）
	_, _, ok = cache.GetNextMember("test", "c")
	if ok {
		t.Error("GetNextMember(c) should return false (last member)")
	}

	// 查询不存在的 member
	_, _, ok = cache.GetNextMember("test", "nonexistent")
	if ok {
		t.Error("GetNextMember(nonexistent) should return false")
	}

	// 测试字符串版本
	nextMember, nextScoreStr, ok := cache.GetNextMemberString("test", "a")
	if !ok || nextMember != "b" {
		t.Errorf("GetNextMemberString failed")
	}
	if nextScoreStr[:2] != "20" {
		t.Errorf("GetNextMemberString score = %s, want starting with 20", nextScoreStr)
	}
}

// BenchmarkZAdd 基准测试添加操作
func BenchmarkZAdd(b *testing.B) {
	cache := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ZAddFloat64("bench", string(rune('a'+i%26)), float64(i))
	}
}

// BenchmarkZRange 基准测试范围查询
func BenchmarkZRange(b *testing.B) {
	cache := New()

	// 填充数据
	for i := 0; i < 10000; i++ {
		cache.ZAddFloat64("bench", string(rune('a'+i%26))+string(rune('0'+i/26)), float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ZRange("bench", 0, 100, false)
	}
}

// BenchmarkZScore 基准测试获取分数
func BenchmarkZScore(b *testing.B) {
	cache := New()
	cache.ZAddFloat64("bench", "member", 123.456)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ZScore("bench", "member")
	}
}
