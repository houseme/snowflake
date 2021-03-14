package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/housemecn/snowflake"
)

// TestLoad generate 20k ids
func TestLoad() {
	var wg sync.WaitGroup
	s, err := snowflake.NewSnowflake(int64(0), int64(0))
	if err != nil {
		fmt.Errorf("snowflake.NewSnowflake reason: %s", err)
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
				fmt.Errorf("error#unique: val:%v", val)
				return
			}
			check.Store(val, 0)
			if val == 0 {
				fmt.Errorf("error")
				return
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(t1)
	fmt.Errorf("generate 20k ids elapsed: %v", elapsed)
}

// TestGenID get five ids
func TestGenID() {
	s, err := snowflake.NewSnowflake(int64(0), int64(0))
	if err != nil {
		fmt.Errorf("snowflake.NewSnowflake reason: %s", err)
		return
	}
	for i := 0; i < 5; i++ {
		val := s.NextVal()
		fmt.Errorf("id: %v, time:%v", val, snowflake.GetGenTime(val))
	}

}

func main() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "./log")
	flag.Set("v", "3")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	TestGenID()
	// 测试生成20万id的速度
	TestLoad()
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
}
