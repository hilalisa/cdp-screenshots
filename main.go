package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/benmanns/goworker"
	"github.com/minio/minio-go"

	https "github.com/reinho/cdp-screenshots/http"
	"github.com/reinho/cdp-screenshots/process"
)

var (
	redisURI               = flag.String("redis_uri", "redis://redis:6379/", "redis uri")
	chromePath             = flag.String("chrome_path", "google-chrome", "google chrome path")
	chromeFlags            = flag.String("chrome_flags", "--headless,--disable-gpu,--remote-debugging-port=9222,--no-sandbox,--hide-scrollbars", "google chrome flags")
	screenshotsPerInstance = flag.Int("screenshots_per_instance", 1000, "screenshots per a chrome restart")
	chromeStartDelay       = flag.Duration("chrome_start_delay", 3*time.Second, "how much time to wait after chrome starts")
	httpBind               = flag.String("http_bind", ":8001", "port of the html server")
	screenshotTimeout      = flag.Duration("screenshot_timeout", 5*time.Second, "how long should it take to take the screenshot")
	callbackTimeout        = flag.Duration("callback_timeout", 5*time.Second, "length of the callback timeout")

	s3Endpoint = flag.String("s3_endpoint", "localstack:8000", "s3 endpoint")
	s3UseSSL   = flag.Bool("s3_use_ssl", false, "use ssl for s3?")
	s3KeyID    = flag.String("s3_key_id", "asd", "s3 key id")
	s3Secret   = flag.String("s3_secret", "123", "s3 secret key")
	s3Bucket   = flag.String("s3_bucket", "screenshots-demo", "s3 target bucket")
	s3Region   = flag.String("s3_region", "eu-west-1", "s3 region")
	s3BasePath = flag.String("s3_base_path", "https://s3-eu-west-1.amazonaws.com/screenshots-demo", "base path of the bucket")
)

var (
	chromeProcess *process.Process
	httpServer    *https.HTTP
	s3Service     *minio.Client
)

func main() {
	flag.Parse()

	var err error
	s3Service, err = minio.New(*s3Endpoint, *s3KeyID, *s3Secret, *s3UseSSL)
	if err != nil {
		panic(err)
	}

	err = s3Service.MakeBucket(*s3Bucket, *s3Region)
	if err != nil {
		exists, err := s3Service.BucketExists(*s3Bucket)
		if err == nil && exists {
			log.Printf("We already own %s\n", *s3Bucket)
		} else {
			log.Fatalln(err)
		}
	}

	chromeProcess, err = process.New(
		*screenshotsPerInstance,
		*chromeStartDelay,
		*chromePath,
		strings.Split(*chromeFlags, ",")...,
	)
	if err != nil {
		log.Fatalf("Unable to start Chrome: %+v", err)
	}

	goworker.SetSettings(goworker.WorkerSettings{
		URI:            *redisURI,
		Connections:    32,
		Queues:         []string{"screenshots"},
		UseNumber:      true,
		ExitOnComplete: false,
		Concurrency:    25,
		Namespace:      "resque:",
		Interval:       1.0,
	})

	goworker.Register("Screenshot", screenshotWorker)

	httpServer = &https.HTTP{
		Data: map[string]string{},
	}
	go func() {
		if err := http.ListenAndServe(*httpBind, httpServer); err != nil {
			log.Fatal(err)
		}
	}()

	if err := goworker.Work(); err != nil {
		log.Fatal(err)
	}
}
