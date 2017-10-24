package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/dchest/uniuri"
	"github.com/go-redis/redis"
)

var (
	bind         = flag.String("bind", ":7000", "http server binding")
	callbackBase = flag.String("callback_base", "http://192.168.1.25:7000/callback", "url of the endpoint callback")
	redisAddr    = flag.String("redis_addr", "redis:6379", "redis address")
	redisKey     = flag.String("redis_key", "resque:queue:screenshots", "key for the queue")
	indexFile    = flag.String("index_file", "./index.html", "path of the index file")
)

type result struct {
	ContentType string
	Response    []byte
	Error       error
}

var (
	callbacks   = map[string]chan result{}
	callbacksMu sync.RWMutex
)

type message struct {
	Class string        `json:"class"`
	Args  []interface{} `json:"args"`
}

type input struct {
	HTML         string  `json:"html"` // either HTML or URL, preferred HTML
	URL          string  `json:"url"`
	Width        int64   `json:"width"`
	Height       int64   `json:"height"`
	Scaling      float64 `json:"scaling"`   // 1.00 by default
	Delay        int64   `json:"delay"`     // in ms
	FullPage     bool    `json:"full_page"` // take a screenshot of the full page
	Format       string  `json:"format"`    // jpeg or png
	Quality      int64   `json:"quality"`
	CallbackType string  `json:"callback_type"` // "blob" or "s3", "blob" by default
}

func main() {
	flag.Parse()

	/*
		index, err := ioutil.ReadFile(*indexFile)
		if err != nil {
			panic(err)
		}
	*/

	client := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

		callbacksMu.RLock()
		resultChan, ok := callbacks[key]
		callbacksMu.RUnlock()

		if !ok {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			resultChan <- result{
				ContentType: r.Header.Get("Content-Type"),
				Response:    nil,
				Error:       err,
			}
			return
		}

		resultChan <- result{
			ContentType: r.Header.Get("Content-Type"),
			Response:    body,
			Error:       nil,
		}

		w.Write([]byte("OK"))

		callbacksMu.Lock()
		delete(callbacks, key)
		callbacksMu.Unlock()
	})

	http.HandleFunc("/screenshot", func(w http.ResponseWriter, r *http.Request) {
		inputMsg := &input{}
		if err := json.NewDecoder(r.Body).Decode(inputMsg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		key := uniuri.New()
		callback := *callbackBase + "?key=" + key
		resultChan := make(chan result)

		callbacksMu.Lock()
		callbacks[key] = resultChan
		callbacksMu.Unlock()

		queueMsg, err := json.Marshal(&message{
			Class: "Screenshot",
			Args: []interface{}{
				inputMsg.HTML,
				inputMsg.URL,
				inputMsg.Width,
				inputMsg.Height,
				inputMsg.Scaling,
				inputMsg.Delay,
				inputMsg.FullPage,
				inputMsg.Format,
				inputMsg.Quality,
				callback,
				inputMsg.CallbackType,
			},
		})
		if err != nil {
			panic(err)
		}

		if err := client.RPush(*redisKey, string(queueMsg)).Err(); err != nil {
			panic(err)
		}

		// wait for the callback here
		result := <-resultChan

		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", result.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(result.Response)))
		w.Write(result.Response)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index, err := ioutil.ReadFile(*indexFile)
		if err != nil {
			panic(err)
		}
		w.Write(index)
	})

	http.ListenAndServe(*bind, nil)
}
