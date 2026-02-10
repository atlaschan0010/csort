package csort_test

import (
	"fmt"
	"math/big"

	"github.com/atlaschan0010/csort"
)

func ExampleCacheZSort() {
	// 创建实例
	cache := csort.New()

	// 示例 1: 添加成员（高精度小数）
	score := new(big.Rat)
	score.SetString("12345678901234567890.12345678901234567890")
	cache.ZAdd("myzset", "member1", score)

	// 示例 2: 使用字符串添加
	cache.ZAddString("myzset", "member2", "3.14159265358979323846")

	// 示例 3: 使用 float64/int64 添加
	cache.ZAddFloat64("myzset", "member3", 100.5)
	cache.ZAddInt64("myzset", "member4", 42)

	// 获取分数
	got, _ := cache.ZScoreString("myzset", "member1")
	fmt.Println("member1 score:", got[:20]+"...") // 只显示前20位

	// 获取排名（从0开始）
	rank, _ := cache.ZRank("myzset", "member1")
	fmt.Println("member1 rank:", rank)

	// 获取成员数量
	card, _ := cache.ZCard("myzset")
	fmt.Println("total members:", card)

	// Output:
	// member1 score: 12345678901234567890...
	// member1 rank: 3
	// total members: 4
}

func ExampleCacheZSort_highPrecision() {
	cache := csort.New()

	// Redis 无法精确存储的分数
	highPrecision := "0.1234567890123456789012345678901234567890"

	// CacheZSort 可以精确存储
	cache.ZAddString("test", "member", highPrecision)

	score, _ := cache.ZScore("test", "member")

	// 验证精度
	expected := new(big.Rat)
	expected.SetString(highPrecision)

	if score.Cmp(expected) == 0 {
		fmt.Println("High precision match!")
	}

	// Output:
	// High precision match!
}

func ExampleCacheZSort_rangeQuery() {
	cache := csort.New()

	// 添加一些数据
	cache.ZAddFloat64("leaderboard", "alice", 100)
	cache.ZAddFloat64("leaderboard", "bob", 200)
	cache.ZAddFloat64("leaderboard", "charlie", 150)
	cache.ZAddFloat64("leaderboard", "david", 300)
	cache.ZAddFloat64("leaderboard", "eve", 250)

	// 获取前3名（倒序，分数高的在前）
	top3 := cache.ZRevRange("leaderboard", 0, 2, true)
	fmt.Println("Top 3:")
	for i := 0; i < len(top3); i += 2 {
		member := top3[i].(string)
		score := top3[i+1].(string)
		fmt.Printf("  %d. %s: %s\n", i/2+1, member, score)
	}

	// 获取排名在 1-3 的成员（正序）
	midRange := cache.ZRange("leaderboard", 1, 3, false)
	fmt.Println("Rank 1-3:", midRange)

	// Output:
	// Top 3:
	//   1. david: 300.00000000000000000000
	//   2. eve: 250.00000000000000000000
	//   3. bob: 200.00000000000000000000
	// Rank 1-3: [charlie bob eve]
}

func ExampleCacheZSort_scoreRange() {
	cache := csort.New()

	// 添加价格数据
	cache.ZAddString("prices", "item1", "10.99")
	cache.ZAddString("prices", "item2", "25.50")
	cache.ZAddString("prices", "item3", "5.00")
	cache.ZAddString("prices", "item4", "100.00")
	cache.ZAddString("prices", "item5", "15.00")

	// 查询价格在 10-30 之间的商品
	min := new(big.Rat)
	min.SetString("10")
	max := new(big.Rat)
	max.SetString("30")

	items := cache.ZRangeByScore("prices", min, max, true, 0, -1)
	fmt.Println("Items between 10-30:")
	for i := 0; i < len(items); i += 2 {
		member := items[i].(string)
		score := items[i+1].(string)
		fmt.Printf("  %s: $%s\n", member, score)
	}

	// 统计价格在 10-30 之间的商品数量
	count := cache.ZCount("prices", min, max)
	fmt.Println("Count:", count)

	// Output:
	// Items between 10-30:
	//   item1: $10.99000000000000000000
	//   item5: $15.00000000000000000000
	//   item2: $25.50000000000000000000
	// Count: 3
}

func ExampleCacheZSort_increment() {
	cache := csort.New()

	// 初始化玩家分数
	cache.ZAddFloat64("game", "player1", 100)

	// 增加分数
	increment := new(big.Rat)
	increment.SetString("50")
	newScore, _ := cache.ZIncrBy("game", "player1", increment)
	fmt.Println("New score:", newScore)

	// 对不存在的成员增加（相当于添加）
	newScore2, _ := cache.ZIncrBy("game", "player2", big.NewRat(25, 1))
	fmt.Println("Player2 score:", newScore2)

	// Output:
	// New score: 150.00000000000000000000
	// Player2 score: 25.00000000000000000000
}

func ExampleCacheZSort_neighbor() {
	cache := csort.New()

	// 添加排行榜数据
	cache.ZAddFloat64("ranking", "alice", 100)
	cache.ZAddFloat64("ranking", "bob", 200)
	cache.ZAddFloat64("ranking", "charlie", 300)

	// 查询 bob 的前一位
	prevMember, prevScore, _ := cache.GetPrevMember("ranking", "bob")
	fmt.Printf("Before bob: %s (score: %s)\n", prevMember, prevScore.FloatString(0))

	// 查询 bob 的后一位
	nextMember, nextScore, _ := cache.GetNextMember("ranking", "bob")
	fmt.Printf("After bob: %s (score: %s)\n", nextMember, nextScore.FloatString(0))

	// 查询 alice 的前一位（不存在）
	_, _, ok := cache.GetPrevMember("ranking", "alice")
	fmt.Printf("Before alice exists: %v\n", ok)

	// 查询 charlie 的后一位（不存在）
	_, _, ok = cache.GetNextMember("ranking", "charlie")
	fmt.Printf("After charlie exists: %v\n", ok)

	// Output:
	// Before bob: alice (score: 100)
	// After bob: charlie (score: 300)
	// Before alice exists: false
	// After charlie exists: false
}
