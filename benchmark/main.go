package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dchest/uniuri"
	"github.com/go-redis/redis"
)

var (
	bind         = flag.String("bind", ":7001", "http server binding")
	redisAddr    = flag.String("redis_addr", "localhost:6379", "redis address")
	redisKey     = flag.String("redis_key", "resque:queue:screenshots", "key for the queue")
	callbackBase = flag.String("callback_base", "http://192.168.1.25:7001/callback", "url of the endpoint callback")

	senders  = flag.Int("senders", 32, "how many senders should there be")
	duration = flag.Duration("duration", time.Second*30, "how long should the test run")
)

type result struct {
	Latency time.Duration
	Error   error
}

type entry struct {
	Start time.Time
	Chan  chan result
}

var (
	callbacks   = map[string]entry{}
	callbacksMu sync.RWMutex
	finished    bool
)

type message struct {
	Class string        `json:"class"`
	Args  []interface{} `json:"args"`
}

func main() {
	flag.Parse()

	client := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		if finished {
			return
		}

		key := r.URL.Query().Get("key")

		callbacksMu.RLock()
		entry, ok := callbacks[key]
		callbacksMu.RUnlock()

		if !ok {
			return
		}

		entry.Chan <- result{
			Latency: now.Sub(entry.Start),
			Error:   nil,
		}

		w.Write([]byte("OK"))

		callbacksMu.Lock()
		delete(callbacks, key)
		callbacksMu.Unlock()
	})

	go func() {
		if err := http.ListenAndServe(*bind, nil); err != nil {
			panic(err)
		}
	}()

	accumulatedLatencies := []time.Duration{}
	accumulatedLatenciesMu := sync.Mutex{}
	accumulatedLatenciesGroup := sync.WaitGroup{}

	for i := 0; i < *senders; i++ {
		accumulatedLatenciesGroup.Add(1)

		// Spawn the senders
		go func() {
			latencies := []time.Duration{}

			for {
				if finished {
					break
				}

				key := uniuri.New()
				callback := *callbackBase + "?key=" + key

				callbacksMu.Lock()
				callbackChan := make(chan result)
				callbacks[key] = entry{
					Start: time.Now(),
					Chan:  callbackChan,
				}
				callbacksMu.Unlock()

				queueMsg, err := json.Marshal(&message{
					Class: "Screenshot",
					Args: []interface{}{
						"hello world",
						"", // URL
						1920,
						1080,
						float64(0.5),
						0,
						false,
						"png",
						0,
						callback,
						"blob",
					},
				})
				if err != nil {
					panic(err)
				}

				if err := client.RPush(*redisKey, string(queueMsg)).Err(); err != nil {
					panic(err)
				}

				select {
				case res := <-callbackChan:
					latencies = append(latencies, res.Latency)
				case <-time.After(time.Second * 3):
					continue
				}
			}

			accumulatedLatenciesMu.Lock()
			accumulatedLatencies = append(accumulatedLatencies, latencies...)
			accumulatedLatenciesMu.Unlock()

			accumulatedLatenciesGroup.Done()
		}()
	}

	time.Sleep(*duration)
	finished = true
	accumulatedLatenciesGroup.Wait()

	file, err := os.OpenFile("./results.txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, dur := range accumulatedLatencies {
		fmt.Fprintln(file, int64(dur))
	}
}
