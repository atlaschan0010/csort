# csort

[![Go Reference](https://pkg.go.dev/badge/github.com/atlaschan0010/csort.svg)](https://pkg.go.dev/github.com/atlaschan0010/csort)
[![Go Report Card](https://goreportcard.com/badge/github.com/atlaschan0010/csort)](https://goreportcard.com/report/github.com/atlaschan0010/csort)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

[English](#english) | [中文](#中文)

---

## English

A high-precision, in-memory sorted set component for Go — **Redis ZSet API compatible**, powered by `math/big.Rat` for arbitrary-precision rational number scores.

### ✨ Features

- **Arbitrary Precision** — Uses `big.Rat` to store scores, eliminating the floating-point precision loss inherent in Redis ZSet's `double` (64-bit IEEE 754)
- **Redis API Compatible** — Drop-in replacement API mirroring Redis ZSet commands (`ZADD`, `ZRANGE`, `ZRANK`, `ZSCORE`, etc.)
- **High Performance** — Built on a Skip List with O(log N) insert, delete, and rank queries
- **Multi-Key Support** — Manages multiple independent sorted sets within a single instance
- **Zero Dependencies** — Only uses the Go standard library
- **Thread Safe** — All operations are fully concurrent-safe with fine-grained locking

### 📦 Installation

```bash
go get github.com/atlaschan0010/csort
```

> **Requirements:** Go 1.22+

### 🚀 Quick Start

```go
package main

import (
    "fmt"
    "math/big"

    csort "github.com/atlaschan0010/csort"
)

func main() {
    cache := csort.New()

    // Add members with arbitrary-precision scores
    score := new(big.Rat)
    score.SetString("12345678901234567890.12345678901234567890")
    cache.ZAdd("leaderboard", "player1", score)

    // Convenience methods for common types
    cache.ZAddString("leaderboard", "player2", "3.14159265358979323846")
    cache.ZAddFloat64("leaderboard", "player3", 100.5)
    cache.ZAddInt64("leaderboard", "player4", 42)

    // Query score (exact precision preserved)
    got, _ := cache.ZScoreString("leaderboard", "player1")
    fmt.Println("Score:", got)

    // Query rank (0-based)
    rank, _ := cache.ZRank("leaderboard", "player1")
    fmt.Println("Rank:", rank)

    // Range query with scores
    members := cache.ZRange("leaderboard", 0, -1, true)
    fmt.Println("Members:", members)
}
```

### 📖 API Reference

#### Add Operations

| Method | Description |
|--------|-------------|
| `ZAdd(key, member string, score *big.Rat) bool` | Add a member with a `*big.Rat` score |
| `ZAddString(key, member, score string) (bool, error)` | Add a member with a string-format score |
| `ZAddFloat64(key, member string, score float64) bool` | Add a member with a `float64` score |
| `ZAddInt64(key, member string, score int64) bool` | Add a member with an `int64` score |
| `ZAddMultiple(key string, members map[string]*big.Rat) int` | Batch add multiple members |
| `ZIncrBy(key, member string, increment *big.Rat) (string, bool)` | Increment a member's score |

#### Remove Operations

| Method | Description |
|--------|-------------|
| `ZRem(key, member string) bool` | Remove a single member |
| `ZRemMultiple(key string, members []string) int` | Remove multiple members |
| `ZRemRangeByRank(key string, start, stop int) int` | Remove members by rank range |
| `ZRemRangeByScore(key string, min, max *big.Rat) int` | Remove members by score range |
| `Del(keys ...string) int` | Delete entire sorted set(s) |
| `ZPopMin(key string, count int) []ScoreMember` | Pop members with the lowest scores |
| `ZPopMax(key string, count int) []ScoreMember` | Pop members with the highest scores |

#### Query Operations

| Method | Description |
|--------|-------------|
| `ZScore(key, member string) (*big.Rat, bool)` | Get member score as `*big.Rat` |
| `ZScoreString(key, member string) (string, bool)` | Get member score as string |
| `ZRank(key, member string) (int, bool)` | Get forward rank (0-based) |
| `ZRevRank(key, member string) (int, bool)` | Get reverse rank (0-based) |
| `GetMemberRank(key, member string) (int, bool)` | Get forward rank (1-based) |
| `ZCard(key string) (int, bool)` | Get number of members |
| `ZCount(key string, min, max *big.Rat) int` | Count members within score range |

#### Neighbor Queries

| Method | Description |
|--------|-------------|
| `GetPrevMember(key, member string) (string, *big.Rat, bool)` | Get the previous member (lower score) |
| `GetNextMember(key, member string) (string, *big.Rat, bool)` | Get the next member (higher score) |
| `GetPrevMemberString(key, member string) (string, string, bool)` | Get previous member (score as string) |
| `GetNextMemberString(key, member string) (string, string, bool)` | Get next member (score as string) |

#### Range Queries

| Method | Description |
|--------|-------------|
| `ZRange(key string, start, stop int, withScores bool) []interface{}` | Query by rank range (ascending) |
| `ZRevRange(key string, start, stop int, withScores bool) []interface{}` | Query by rank range (descending) |
| `ZRangeByScore(key string, min, max *big.Rat, withScores bool, offset, count int) []interface{}` | Query by score range (ascending) |
| `ZRevRangeByScore(key string, max, min *big.Rat, withScores bool, offset, count int) []interface{}` | Query by score range (descending) |

#### Management Operations

| Method | Description |
|--------|-------------|
| `Exists(key string) bool` | Check if a key exists |
| `Keys() []string` | Get all keys |
| `Flush()` | Clear all data |

### 📊 Use Cases

#### Leaderboard

```go
cache := csort.New()

cache.ZAddFloat64("leaderboard", "alice", 100)
cache.ZAddFloat64("leaderboard", "bob", 200)
cache.ZAddFloat64("leaderboard", "charlie", 150)
cache.ZAddFloat64("leaderboard", "david", 300)
cache.ZAddFloat64("leaderboard", "eve", 250)

// Top 3 players (descending)
top3 := cache.ZRevRange("leaderboard", 0, 2, true)
for i := 0; i < len(top3); i += 2 {
    fmt.Printf("%d. %s — %s\n", i/2+1, top3[i], top3[i+1])
}
```

#### High-Precision Financial Data

```go
cache := csort.New()

// Store prices with full decimal precision
cache.ZAddString("prices", "BTC", "67432.12345678901234567890")
cache.ZAddString("prices", "ETH", "3521.98765432109876543210")

score, _ := cache.ZScore("prices", "BTC")
// score retains all 20+ decimal places — no precision loss!
```

#### Neighbor Lookup

```go
cache := csort.New()

cache.ZAddFloat64("ranking", "alice", 100)
cache.ZAddFloat64("ranking", "bob", 200)
cache.ZAddFloat64("ranking", "charlie", 300)

prev, prevScore, _ := cache.GetPrevMember("ranking", "bob")
next, nextScore, _ := cache.GetNextMember("ranking", "bob")
fmt.Printf("Before bob: %s (%s)\n", prev, prevScore.FloatString(0))
fmt.Printf("After bob: %s (%s)\n", next, nextScore.FloatString(0))
```

### ⚡ Benchmarks

Benchmarked on **Apple M3 Max** (Go 1.25, `arm64`):

```
goos: darwin
goarch: arm64
cpu: Apple M3 Max

BenchmarkZAdd-16       1,247,178       951.3 ns/op     1,945 B/op     34 allocs/op
BenchmarkZRange-16       160,768     7,450   ns/op    14,168 B/op    405 allocs/op
BenchmarkZScore-16    22,037,368        53.66 ns/op       80 B/op      3 allocs/op
```

| Operation | Throughput | Time Complexity |
|-----------|-----------|-----------------|
| **ZAdd** | ~1,050,000 ops/sec | O(log N) |
| **ZRange** | ~134,000 ops/sec | O(log N + M) |
| **ZScore** | ~18,600,000 ops/sec | O(1) |

### 🏗️ Architecture

#### Data Structures

- **Skip List** — Core sorted structure providing O(log N) insert, delete, and rank operations with span-based rank calculation
- **`big.Rat`** — Go's standard library arbitrary-precision rational number type for exact score representation
- **`memberMap`** — Hash map for O(1) member-to-node lookups (`ZScore`, `ZRem`)

#### Concurrency Model

```
CacheZSort (global RWMutex)
├── sets map[string]*ZSet
│   ├── "key1" → ZSet (per-key RWMutex)
│   │             └── SkipList (internal RWMutex)
│   ├── "key2" → ZSet (per-key RWMutex)
│   │             └── SkipList (internal RWMutex)
│   └── ...
```

- **Two-tier locking**: A global `RWMutex` guards the key map; each `ZSet` has its own `RWMutex` to minimize contention across keys
- **Read-heavy optimization**: Read operations acquire read locks, allowing concurrent reads on the same key

### ⚠️ Notes

1. **Memory** — Data is stored entirely in memory; capacity is bounded by available RAM
2. **Persistence** — No built-in persistence; data is lost on process restart
3. **Score Output** — `ZScoreString` / `FloatString()` output is formatted with a fixed number of decimal places (20 by default)

### 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## 中文

一个基于 Go 实现的高精度内存排序组件 —— **兼容 Redis ZSet API**，使用 `math/big.Rat` 支持任意精度有理数分数。

### ✨ 特性

- **任意精度** — 使用 `big.Rat` 存储分数，解决 Redis ZSet 中 `double`（64 位 IEEE 754 浮点数）固有的精度丢失问题
- **Redis API 兼容** — 提供与 Redis ZSet 命令一致的 API 接口（`ZADD`、`ZRANGE`、`ZRANK`、`ZSCORE` 等）
- **高性能** — 基于跳表（Skip List）实现，插入、删除、排名查询均为 O(log N)
- **多 Key 支持** — 单实例内管理多个独立的有序集合
- **零依赖** — 仅使用 Go 标准库
- **线程安全** — 所有操作均为并发安全，采用细粒度锁策略

### 📦 安装

```bash
go get github.com/atlaschan0010/csort
```

> **要求：** Go 1.22+

### 🚀 快速开始

```go
package main

import (
    "fmt"
    "math/big"

    csort "github.com/atlaschan0010/csort"
)

func main() {
    cache := csort.New()

    // 添加成员（支持任意精度小数）
    score := new(big.Rat)
    score.SetString("12345678901234567890.12345678901234567890")
    cache.ZAdd("leaderboard", "player1", score)

    // 便捷方法：字符串、float64、int64
    cache.ZAddString("leaderboard", "player2", "3.14159265358979323846")
    cache.ZAddFloat64("leaderboard", "player3", 100.5)
    cache.ZAddInt64("leaderboard", "player4", 42)

    // 查询分数（精度完全保留）
    got, _ := cache.ZScoreString("leaderboard", "player1")
    fmt.Println("分数:", got)

    // 查询排名（从 0 开始）
    rank, _ := cache.ZRank("leaderboard", "player1")
    fmt.Println("排名:", rank)

    // 范围查询（带分数）
    members := cache.ZRange("leaderboard", 0, -1, true)
    fmt.Println("成员:", members)
}
```

### 📖 API 参考

#### 添加操作

| 方法 | 说明 |
|------|------|
| `ZAdd(key, member string, score *big.Rat) bool` | 添加成员（`*big.Rat` 分数）|
| `ZAddString(key, member, score string) (bool, error)` | 添加成员（字符串格式分数）|
| `ZAddFloat64(key, member string, score float64) bool` | 添加成员（`float64` 分数）|
| `ZAddInt64(key, member string, score int64) bool` | 添加成员（`int64` 分数）|
| `ZAddMultiple(key string, members map[string]*big.Rat) int` | 批量添加成员 |
| `ZIncrBy(key, member string, increment *big.Rat) (string, bool)` | 增加成员分数 |

#### 删除操作

| 方法 | 说明 |
|------|------|
| `ZRem(key, member string) bool` | 删除单个成员 |
| `ZRemMultiple(key string, members []string) int` | 删除多个成员 |
| `ZRemRangeByRank(key string, start, stop int) int` | 按排名范围删除 |
| `ZRemRangeByScore(key string, min, max *big.Rat) int` | 按分数范围删除 |
| `Del(keys ...string) int` | 删除整个有序集合 |
| `ZPopMin(key string, count int) []ScoreMember` | 弹出分数最低的成员 |
| `ZPopMax(key string, count int) []ScoreMember` | 弹出分数最高的成员 |

#### 查询操作

| 方法 | 说明 |
|------|------|
| `ZScore(key, member string) (*big.Rat, bool)` | 获取成员分数（`*big.Rat`）|
| `ZScoreString(key, member string) (string, bool)` | 获取成员分数（字符串）|
| `ZRank(key, member string) (int, bool)` | 获取正序排名（从 0 开始）|
| `ZRevRank(key, member string) (int, bool)` | 获取倒序排名（从 0 开始）|
| `GetMemberRank(key, member string) (int, bool)` | 获取正序排名（从 1 开始）|
| `ZCard(key string) (int, bool)` | 获取成员数量 |
| `ZCount(key string, min, max *big.Rat) int` | 统计分数范围内成员数量 |

#### 邻居查询

| 方法 | 说明 |
|------|------|
| `GetPrevMember(key, member string) (string, *big.Rat, bool)` | 获取前一位成员（分数更小）|
| `GetNextMember(key, member string) (string, *big.Rat, bool)` | 获取后一位成员（分数更大）|
| `GetPrevMemberString(key, member string) (string, string, bool)` | 获取前一位成员（分数为字符串）|
| `GetNextMemberString(key, member string) (string, string, bool)` | 获取后一位成员（分数为字符串）|

#### 范围查询

| 方法 | 说明 |
|------|------|
| `ZRange(key string, start, stop int, withScores bool) []interface{}` | 按排名范围查询（正序）|
| `ZRevRange(key string, start, stop int, withScores bool) []interface{}` | 按排名范围查询（倒序）|
| `ZRangeByScore(key string, min, max *big.Rat, withScores bool, offset, count int) []interface{}` | 按分数范围查询（正序）|
| `ZRevRangeByScore(key string, max, min *big.Rat, withScores bool, offset, count int) []interface{}` | 按分数范围查询（倒序）|

#### 管理操作

| 方法 | 说明 |
|------|------|
| `Exists(key string) bool` | 检查 Key 是否存在 |
| `Keys() []string` | 获取所有 Key |
| `Flush()` | 清空所有数据 |

### 📊 使用场景

#### 排行榜

```go
cache := csort.New()

cache.ZAddFloat64("leaderboard", "alice", 100)
cache.ZAddFloat64("leaderboard", "bob", 200)
cache.ZAddFloat64("leaderboard", "charlie", 150)
cache.ZAddFloat64("leaderboard", "david", 300)
cache.ZAddFloat64("leaderboard", "eve", 250)

// 获取前 3 名（倒序，分数高的在前）
top3 := cache.ZRevRange("leaderboard", 0, 2, true)
for i := 0; i < len(top3); i += 2 {
    fmt.Printf("%d. %s — %s\n", i/2+1, top3[i], top3[i+1])
}
```

#### 高精度金融数据

```go
cache := csort.New()

// 存储完整小数精度的价格
cache.ZAddString("prices", "BTC", "67432.12345678901234567890")
cache.ZAddString("prices", "ETH", "3521.98765432109876543210")

score, _ := cache.ZScore("prices", "BTC")
// score 保留所有 20+ 位小数 —— 无精度丢失！
```

#### 邻居查询

```go
cache := csort.New()

cache.ZAddFloat64("ranking", "alice", 100)
cache.ZAddFloat64("ranking", "bob", 200)
cache.ZAddFloat64("ranking", "charlie", 300)

prev, prevScore, _ := cache.GetPrevMember("ranking", "bob")
next, nextScore, _ := cache.GetNextMember("ranking", "bob")
fmt.Printf("bob 前一位: %s (%s)\n", prev, prevScore.FloatString(0))
fmt.Printf("bob 后一位: %s (%s)\n", next, nextScore.FloatString(0))
```

### ⚡ 性能基准

在 **Apple M3 Max** 上的基准测试（Go 1.25，`arm64`）：

```
goos: darwin
goarch: arm64
cpu: Apple M3 Max

BenchmarkZAdd-16       1,247,178       951.3 ns/op     1,945 B/op     34 allocs/op
BenchmarkZRange-16       160,768     7,450   ns/op    14,168 B/op    405 allocs/op
BenchmarkZScore-16    22,037,368        53.66 ns/op       80 B/op      3 allocs/op
```

| 操作 | 吞吐量 | 时间复杂度 |
|------|--------|-----------|
| **ZAdd** | ~1,050,000 次/秒 | O(log N) |
| **ZRange** | ~134,000 次/秒 | O(log N + M) |
| **ZScore** | ~18,600,000 次/秒 | O(1) |

### 🏗️ 架构设计

#### 数据结构

- **跳表（Skip List）** — 核心排序结构，提供 O(log N) 的插入、删除、排名操作，基于 span 实现排名计算
- **`big.Rat`** — Go 标准库的任意精度有理数类型，精确表示分数
- **`memberMap`** — 哈希表，O(1) 的成员到节点查找（`ZScore`、`ZRem`）

#### 并发模型

```
CacheZSort（全局 RWMutex）
├── sets map[string]*ZSet
│   ├── "key1" → ZSet（独立 RWMutex）
│   │             └── SkipList（内部 RWMutex）
│   ├── "key2" → ZSet（独立 RWMutex）
│   │             └── SkipList（内部 RWMutex）
│   └── ...
```

- **两级锁机制**：全局 `RWMutex` 守护 key 映射表；每个 `ZSet` 拥有独立的 `RWMutex`，最大程度减少跨 key 的锁竞争
- **读优化**：读操作获取读锁，允许同一 key 上的并发读取

### ⚠️ 注意事项

1. **内存使用** — 数据完全存储在内存中，容量受限于可用内存
2. **持久化** — 当前版本不支持持久化，进程重启后数据丢失
3. **分数输出** — `ZScoreString` / `FloatString()` 输出时默认保留 20 位小数

### 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

---

## License

[Apache License 2.0](LICENSE)
