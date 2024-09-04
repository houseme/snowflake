# ❄️ Go-Snowflake


### A Snowflake Generator for Go
### A simple to use Go (golang) package to generate or parse Twitter snowflake IDs


[![Go Reference](https://pkg.go.dev/badge/github.com/houseme/snowflake.svg)](https://pkg.go.dev/github.com/houseme/snowflake)
[![GoFrame CI](https://github.com/houseme/snowflake/actions/workflows/go.yml/badge.svg)](https://github.com/houseme/snowflake/actions/workflows/gf.yml)
[![Go Report](https://goreportcard.com/badge/github.com/houseme/snowflake?v=1)](https://goreportcard.com/report/github.com/houseme/snowflake)
[![Production Ready](https://img.shields.io/badge/production-ready-blue.svg)](https://github.com/housemecn/snowflake)
[![License](https://img.shields.io/github/license/housemecn/snowflake.svg?style=flat)](https://github.com/housemecn/snowflake)

## Snowflake 简介

在单机系统中我们会使用自增 id 作为数据的唯一 id，自增 id 在数据库中有利于排序和索引，但是在分布式系统中如果还是利用数据库的自增 id 会引起冲突，自增 id 非常容易被爬虫爬取数据。在分布式系统中有使用 uuid 作为数据唯一 id 的，但是 uuid 是一串随机字符串，所以它无法被排序。

Twitter 设计了 Snowflake 算法为分布式系统生成 ID,Snowflake 的 id 是 int64 类型，它通过 datacenterId 和 workerId 来标识分布式系统，下面看下它的组成：

| 1bit      | 41bit             | 5bit    | 5bit  | 12bit |
|-----------|-------------------|---------|-------|-------|
| 符号位（保留字段） | 时间戳 (当前时间 - 纪元时间) | 数据中心 id | 机器 id | 自增序列  |

### 算法简介

在使用 Snowflake 生成 id 时，首先会计算时间戳 timestamp（当前时间 - 纪元时间），如果 timestamp 数据超过 41bit 则异常。同样需要判断 datacenterId 和 workerId 不能超过 5bit(0-31)，在处理自增序列时，如果发现自增序列超过 12bit 时需要等待，因为当前毫秒下 12bit 的自增序列被用尽，需要进入下一毫秒后自增序列继续从 0 开始递增。

---

## 🚀 快速开始

### 🕹 克隆 & 运行

```bash
git clone https://github.com/houseme/snowflake.git

go run ./.example/main.go
```

### 💾 安装 & 导入

```bash
go get github.com/houseme/snowflake
```
```go
// 在项目中导入模块
import "github.com/houseme/snowflake"
```

### ⚠️注意事项

* 在多实例（多个 snowflake 对象）的并发环境下，请确保每个实例（datacenterId，workerId）的唯一性，否则生成的 ID 可能冲突。

### 📊 测试

本机测试：

| 参数 | 配置 |
|---|---|
| OS | MacBook Pro (16-inch, 2019)|
| CPU | Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz |
| RAM | 64 GB 2667 MHz DDR4 |

> 测试代码

```go
func TestLoad() {
    var wg sync.WaitGroup
    s, err := snowflake.NewSnowflake(int64(0), int64(0))
    if err != nil {
        glog.Error(err)
        return
    }
    var check sync.Map
    t1 := time.Now()
    for i := 0; i < 200000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            val := s.NextVal()
            if _, ok := check.Load(val); ok {
                // id 冲突检查
                glog.Error(fmt.Errorf("error#unique: val:%v", val))
                return
            }
            check.Store(val, 0)
            if val == 0 {
                glog.Error(fmt.Errorf("error"))
                return
            }
        }()
    }
    wg.Wait()
    elapsed := time.Since(t1)
    glog.Infof("generate 20k ids elapsed: %v", elapsed)
}
```

> 运行结果

![load](https://github.com/houseme/snowflake/raw/main/docs/WX20210314-234124@2x.png)

## 🗂 使用说明

### 创建 Snowflake 对象

```go
// NewSnowflake(datacenterId, workerId int64) (*Snowflake, error)
// 参数 1 (int64): 数据中心 ID (可用范围:0-31)
// 参数 2 (int64): 机器 ID    (可用范围:0-31)
// 返回 1 (*Snowflake): Snowflake 对象 | nil
// 返回 2 (error): 错误码
s, err := snowflake.NewSnowflake(int64(0), int64(0))
if err != nil {
    glog.Error(err)
    return
}
```

### 生成唯一 ID

```go
s, err := snowflake.NewSnowflake(int64(0), int64(0))
// ......
// (s *Snowflake) NextVal() int64
// 返回 1 (int64): 唯一 ID
id := s.NextVal()
// ......
```

### 通过 ID 获取数据中心 ID 与机器 ID

```go
// ......
// GetDeviceID(sid int64) (datacenterId, workerId int64)
// 参数 1 (int64): 唯一 ID
// 返回 1 (int64): 数据中心 ID
// 返回 2 (int64): 机器 ID
datacenterid, workerid := snowflake.GetDeviceID(id))
```

### 通过 ID 获取时间戳（创建 ID 时的时间戳 - epoch）

```go
// ......
// GetTimestamp(sid int64) (timestamp int64)
// 参数 1 (int64): 唯一 ID
// 返回 1 (int64): 从 epoch 开始计算的时间戳
t := snowflake.GetTimestamp(id)
```

### 通过 ID 获取生成 ID 时的时间戳

```go
// ......
// GetGenTimestamp(sid int64) (timestamp int64)
// 参数 1 (int64): 唯一 ID
// 返回 1 (int64): 唯一 ID 生成时的时间戳
t := snowflake.GetGenTimestamp(id)
```

### 通过 ID 获取生成 ID 时的时间（精确到：秒）

```go
// ......
// GetGenTime(sid int64)
// 参数 1 (int64): 唯一 ID
// 返回 1 (string): 唯一 ID 生成时的时间
tStr := snowflake.GetGenTime(id)
```

### 查看时间戳字段使用占比（41bit 能存储的范围：从 epoch 开始往后 69 年）

```go
// ......
// GetTimestampStatus() (state float64)
// 返回 1 (float64): 时间戳字段使用占比（范围 0.0 - 1.0）
status := snowflake.GetTimestampStatus()
```
### Performance

With default settings, this snowflake generator should be sufficiently fast
enough on most systems to generate 4096 unique ID's per millisecond. This is
the maximum that the snowflake ID format supports. That is, around 243-244
nanoseconds per operation.

Since the snowflake generator is single threaded the primary limitation will be
the maximum speed of a single processor on your system.

To benchmark the generator on your system run the following command inside the
snowflake package directory.

```sh
go test -run=^$ -bench=.
```

## License

Go-snowflake is primarily distributed under the terms of both the Apache License (Version 2.0), thanks for [GUAIK-ORG](https://github.com/GUAIK-ORG/go-snowflake) and [Bwmarrin](https://github.com/bwmarrin/snowflake).
