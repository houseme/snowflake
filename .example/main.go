package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/houseme/snowflake"
)

// TestLoad generate 20k ids
func TestLoad() {
	var wg sync.WaitGroup
	s, err := snowflake.NewSnowflake(int64(1), int64(1))
	if err != nil {
		fmt.Printf("snowflake.NewSnowflake reason: %s \n", err)
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
				// id冲突检查
				fmt.Printf("error#unique: val:%v \n", val)
				return
			}
			check.Store(val, 0)
			if val == 0 {
				fmt.Printf("error")
				return
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(t1)
	fmt.Printf("generate 20k ids elapsed: %v \n", elapsed)
}

// TestGenID get five ids
func TestGenID() {
	s, err := snowflake.NewSnowflake(int64(0), int64(0))
	if err != nil {
		fmt.Printf("snowflake.NewSnowflake reason: %s \n", err)
		return
	}
	t1 := time.Now()
	for i := 0; i < 5; i++ {
		val := s.NextVal()
		fmt.Printf("id: %v, time:%v \n", val, snowflake.GetGenTime(val))
	}
	elapsed := time.Since(t1)
	fmt.Printf("generate 5k ids end  elapsed: %v \n", elapsed)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	TestGenID()
	// 获取时间戳字段已经使用的占比（0.0 - 1.0）
	// 默认开始时间为：2020年01月01日 00:00:00
	fmt.Printf("Timestamp status: %f\n", snowflake.GetTimestampStatus())
	s, _ := snowflake.NewSnowflake(int64(0), int64(0))
	id := s.NextVal()
	// Print out the ID in a few different ways.
	fmt.Printf("Int64  ID: %d\n", id)
	fmt.Printf("String ID: %s\n", id)
	fmt.Printf("Base2  ID: %s\n", id.Base2())
	fmt.Printf("Base64 ID: %s\n", id.Base64())

	// Print out the ID's timestamp
	fmt.Printf("ID Time  : %d\n", id.Time())

	fmt.Println("ID GetTimestamp time ", snowflake.GetTimestamp(id))
	fmt.Println("ID GetGenTimestamp time ", snowflake.GetGenTimestamp(id))
	fmt.Println("ID GetGenTime time ", snowflake.GetGenTime(id))

	// Generate and print, all in one.
	fmt.Printf("ID       : %d\n", s.NextVal().Int64())

	fmt.Println("TestLoad start time:", time.Now())
	// 测试生成20万id的速度
	TestLoad()
	fmt.Println("TestLoad end time:", time.Now())
}
