package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Result struct {
	count int
	total int64
	max   int64
}

var rChan = make(chan Result, 10000)
var wg = new(sync.WaitGroup)

const HOST = "10.0.1.4"
const PORT = "6379"
const KEYS = 1000000

func main() {
	for i, v := range os.Args {
		fmt.Printf("args[%d] -> %s\n", i, v)
	}

	t := flag.Int("thread", 1, "thread numbers")
	p := flag.Bool("load", false, "initial load")
	flag.Parse()

	if *p {
		load()
	}

	for i := 0; i < *t; i++ {
		wg.Add(1)
		go perf()
	}

	wg.Wait()
	close(rChan)

	count := 0
	var total, max int64 = 0, 0
	for r := range rChan {
		count = count + r.count
		total = total + r.total
		if max < r.max {
			max = r.max
		}
	}

	fmt.Printf("count = %d\n", count)
	fmt.Printf("total = %d\n", total)
	fmt.Printf("avg = %f\n", float64(total)/float64(count))
	fmt.Printf("max = %d\n", max)
}

func perf() {
	redisAddr := fmt.Sprintf("%s:%s", HOST, PORT)
	fmt.Printf("redis addr: %s\n", redisAddr)

	const LOOP = 10000

	conn, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(err)
	}

	var start, end time.Time
	var total, max int64 = 0, 0
	for i := 0; i < LOOP; i++ {
		rand.Seed(time.Now().UnixNano())
		start = time.Now()
		r, err := redis.String(conn.Do("GET", string(rand.Intn(KEYS))))
		end = time.Now()
		if err != nil {
			panic(err)
		}
		if i%1000 == 0 {
			fmt.Printf("count: %d, out: %s\n", i, r)
		}

		t := end.Sub(start).Nanoseconds()
		total = total + t
		if max < t {
			max = t
		}

	}

	var result Result
	result.count = LOOP
	result.max = max
	result.total = total

	rChan <- result
	defer wg.Done()
}

func load() {
	const base = "d"
	const length = 1000
	var data string = ""
	for i := 0; i < length; i++ {
		data = data + base
	}

	redisAddr := fmt.Sprintf("%s:%s", HOST, PORT)
	fmt.Printf("redis addr: %s\n", redisAddr)
	conn, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		panic(err)
	}

	for i := 0; i < KEYS; i++ {
		_, err := redis.String(conn.Do("SET", string(i), data))
		if err != nil {
			panic(err)
		}
	}
}
