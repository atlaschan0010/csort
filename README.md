# CacheZSort

一个基于 Go 实现的高精度内存排序组件，兼容 Redis ZSet API，但使用 `math/big.Rat` 支持任意精度的小数分数。

## 特性

- **任意精度小数**: 使用 `big.Rat` 存储分数，解决 Redis ZSet 对超长小数精度丢失的问题
- **Redis API 兼容**: 提供与 Redis ZSet 一致的 API 接口
- **高性能**: 基于跳表（SkipList）实现，O(log N) 的插入、删除、查询性能
- **多 Key 支持**: 支持多个独立的有序集合
- **无第三方依赖**: 仅使用 Go 标准库
- **线程安全**: 所有操作都是并发安全的

## 安装

```bash
go get github.com/atlas/cache-zsort
```

## 快速开始

```go
package main

import (
	"fmt"
	"math/big"
	zsort "github.com/atlas/cache-zsort"
)

func main() {
	// 创建实例
	cache := zsort.New()
	
	// 添加成员（支持高精度小数）
	score := new(big.Rat)
	score.SetString("12345678901234567890.12345678901234567890")
	cache.ZAdd("myzset", "member1", score)
	
	// 使用字符串添加
	cache.ZAddString("myzset", "member2", "3.14159265358979323846")
	
	// 使用 float64/int64 添加
	cache.ZAddFloat64("myzset", "member3", 100.5)
	cache.ZAddInt64("myzset", "member4", 42)
	
	// 获取分数
	got, _ := cache.ZScoreString("myzset", "member1")
	fmt.Println("Score:", got)
	
	// 获取排名（从0开始）
	rank, _ := cache.ZRank("myzset", "member1")
	fmt.Println("Rank:", rank)
	
	// 获取范围
	members := cache.ZRange("myzset", 0, 10, true)
	fmt.Println("Members:", members)
}
```

## API 文档

### 添加操作

| 方法 | 说明 |
|------|------|
| `ZAdd(key, member string, score *big.Rat)` | 添加成员（big.Rat 分数）|
| `ZAddString(key, member, score string)` | 添加成员（字符串分数）|
| `ZAddFloat64(key, member string, score float64)` | 添加成员（float64 分数）|
| `ZAddInt64(key, member string, score int64)` | 添加成员（int64 分数）|
| `ZAddMultiple(key string, members map[string]*big.Rat)` | 批量添加成员 |
| `ZIncrBy(key, member string, increment *big.Rat)` | 增加成员分数 |

### 删除操作

| 方法 | 说明 |
|------|------|
| `ZRem(key, member string)` | 删除单个成员 |
| `ZRemMultiple(key string, members []string)` | 删除多个成员 |
| `ZRemRangeByRank(key string, start, stop int)` | 按排名范围删除 |
| `ZRemRangeByScore(key string, min, max *big.Rat)` | 按分数范围删除 |
| `Del(keys ...string)` | 删除整个有序集合 |
| `ZPopMin(key string, count int)` | 弹出分数最低的成员 |
| `ZPopMax(key string, count int)` | 弹出分数最高的成员 |

### 查询操作

| 方法 | 说明 |
|------|------|
| `ZScore(key, member string)` | 获取成员分数（*big.Rat）|
| `ZScoreString(key, member string)` | 获取成员分数（字符串）|
| `ZRank(key, member string)` | 获取正序排名（从0开始）|
| `ZRevRank(key, member string)` | 获取倒序排名（从0开始）|
| `GetMemberRank(key, member string)` | 获取正序排名（从1开始）|
| `GetPrevMember(key, member string)` | 获取前一位成员（分数更小）|
| `GetNextMember(key, member string)` | 获取后一位成员（分数更大）|
| `GetPrevMemberString(key, member string)` | 获取前一位成员（分数为字符串）|
| `GetNextMemberString(key, member string)` | 获取后一位成员（分数为字符串）|
| `ZCard(key string)` | 获取成员数量 |
| `ZCount(key string, min, max *big.Rat)` | 统计分数范围成员数 |
| `ZScore(key, member string)` | 获取成员分数（*big.Rat）|
| `ZScoreString(key, member string)` | 获取成员分数（字符串）|
| `ZRank(key, member string)` | 获取正序排名（从0开始）|
| `ZRevRank(key, member string)` | 获取倒序排名（从0开始）|
| `ZCard(key string)` | 获取成员数量 |
| `ZCount(key string, min, max *big.Rat)` | 统计分数范围成员数 |

### 范围查询

| 方法 | 说明 |
|------|------|
| `ZRange(key string, start, stop int, withScores bool)` | 按排名范围查询 |
| `ZRevRange(key string, start, stop int, withScores bool)` | 按排名范围倒序查询 |
| `ZRangeByScore(key string, min, max *big.Rat, withScores bool, offset, count int)` | 按分数范围查询 |
| `ZRevRangeByScore(key string, max, min *big.Rat, withScores bool, offset, count int)` | 按分数范围倒序查询 |

### 管理操作

| 方法 | 说明 |
|------|------|
| `Exists(key string)` | 检查 Key 是否存在 |。
| `Keys()` | 获取所有 Key |
| `Flush()` | 清空所有数据 |

## 高精度示例

Redis 使用 double（64位浮点数）存储分数，对于超长小数会丢失精度：

```go
// Redis 无法精确存储的分数
highPrecision := "0.1234567890123456789012345678901234567890"

// CacheZSort 可以精确存储
cache.ZAddString("test", "member", highPrecision)
score, _ := cache.ZScore("test", "member")

// 分数完全匹配，无精度丢失
```

## 性能

基于 Apple M3 Max 的基准测试结果：

```
BenchmarkZAdd-16      776056    1565 ns/op    3210 B/op    65 allocs/op
```

- **ZAdd**: ~640,000 操作/秒
- **ZRange**: 快速的范围查询（取决于范围大小）
- **ZScore**: O(1) 的直接查找

## 实现细节

### 数据结构

- **跳表（Skip List）**: 核心排序结构，提供 O(log N) 的操作复杂度
- **big.Rat**: Go 标准库的任意精度有理数类型，精确表示分数
- **两级锁**: 全局锁管理 keys，每个 ZSet 有自己的锁，减少锁竞争

### 分数比较

使用 `big.Rat.Cmp()` 进行精确比较，避免浮点数精度问题。

## 注意事项

1. **内存使用**: 由于是内存存储，数据量受限于可用内存
2. **持久化**: 当前版本不支持持久化，重启后数据会丢失
3. **分数精度**: 虽然支持任意精度，但 `FloatString()` 输出时会指定小数位数

## 许可证

MIT
