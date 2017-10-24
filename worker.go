package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/rpcc"
	"github.com/minio/minio-go"
	"github.com/pkg/errors"

	"github.com/reinho/cdp-screenshots/screenshot"
)

func screenshotWorker(queue string, args ...interface{}) error {
	msgID := uniuri.New()

	msg := &Message{
		HTML:         args[0].(string),
		URL:          args[1].(string),
		Width:        mustInt64(args[2].(json.Number).Int64()),
		Height:       mustInt64(args[3].(json.Number).Int64()),
		Scaling:      mustFloat64(args[4].(json.Number).Float64()),
		Delay:        mustInt64(args[5].(json.Number).Int64()),
		FullPage:     args[6].(bool),
		Format:       args[7].(string),
		Quality:      mustInt64(args[8].(json.Number).Int64()),
		Callback:     args[9].(string),
		CallbackType: args[10].(string),
	}

	mainCtx, cancel := context.WithTimeout(context.Background(), time.Duration(msg.Delay)*time.Millisecond+*screenshotTimeout+*callbackTimeout)
	defer cancel()

	// First we need to prepare a URL to open
	var targetURL string
	if msg.HTML == "" {
		targetURL = msg.URL
	} else {
		// Load it up into our HTTP server
		id := uniuri.New()
		httpServer.Set(id, msg.HTML)
		defer httpServer.Delete(id)
		targetURL = "http://127.0.0.1" + *httpBind + "/" + id
	}

	log.Printf("[%s] Started processing %s", msgID, targetURL)

	var (
		err    error
		result []byte
	)
	chromeProcess.Execute(func() {
		log.Printf("[%s] Acquired a Chrome process", msgID)

		// remember to not shadow err! we are 100% sure we have chrome on :9222
		ctx, screenshotCancel := context.WithTimeout(mainCtx, time.Duration(msg.Delay)*time.Millisecond+*screenshotTimeout)
		defer screenshotCancel()

		devt := devtool.New("http://127.0.0.1:9222")

		start := time.Now()

		var target *devtool.Target
		target, err = devt.Create(ctx)
		if err != nil {
			return
		}
		defer devt.Close(ctx, target)

		log.Printf("[%s] Acquired a target %s", msgID, target.ID)

		var conn *rpcc.Conn
		conn, err = rpcc.DialContext(ctx, target.WebSocketDebuggerURL)
		if err != nil {
			return
		}
		defer conn.Close()

		// Take a screenshot using the library
		client := cdp.NewClient(conn)

		log.Printf("[%s] Entered the devtools of %s", msgID, target.ID)

		result, err = screenshot.TakeScreenshot(
			ctx, client, targetURL,
			int(msg.Width), int(msg.Height), msg.Scaling,
			time.Duration(msg.Delay)*time.Millisecond,
			msg.FullPage, msg.Format, int(msg.Quality),
		)
		if err != nil {
			return
		}

		log.Printf("[%s] Screenshot of %s taken - elapsed %s", msgID, msg.URL, time.Now().Sub(start).String())
	})
	if err != nil {
		return errors.Wrap(err, "failed to take a screenshot")
	}

	if msg.CallbackType == "" {
		msg.CallbackType = "s3"
	}

	var contentType string
	if msg.Format == "png" {
		contentType = "image/png"
	} else {
		contentType = "image/jpeg"
	}

	var code int
	if msg.CallbackType == "s3" {
		key := msgID + "." + msg.Format

		log.Printf("[%s] Starting upload to S3 at %s", msgID, key)

		if _, err := s3Service.PutObject(
			*s3Bucket,
			key,
			bytes.NewReader(result),
			int64(len(result)),
			minio.PutObjectOptions{
				ContentType: contentType,
			},
		); err != nil {
			return errors.Wrap(err, "unable to upload to s3")
		}

		callbackContext, callbackCancel := context.WithTimeout(mainCtx, *callbackTimeout)
		defer callbackCancel()

		req, err := http.NewRequest("POST", msg.Callback, strings.NewReader(*s3BasePath+"/"+key))
		req.Header.Set("Content-Type", "text/plain")
		resp, err := http.DefaultClient.Do(req.WithContext(callbackContext))
		if err != nil {
			return errors.Wrap(err, "unable to post the callback")
		}
		resp.Body.Close()
		code = resp.StatusCode
	} else if msg.CallbackType == "blob" {
		callbackContext, callbackCancel := context.WithTimeout(mainCtx, *callbackTimeout)
		defer callbackCancel()

		req, err := http.NewRequest("POST", msg.Callback, bytes.NewReader(result))
		req.Header.Set("Content-Type", contentType)
		resp, err := http.DefaultClient.Do(req.WithContext(callbackContext))
		if err != nil {
			return errors.Wrap(err, "unable to post the callback")
		}
		resp.Body.Close()
		code = resp.StatusCode
	} else {
		return errors.Wrap(err, "invalid callback type")
	}

	log.Printf("[%s] Callback %s - %s done, code was %d.", msgID, msg.CallbackType, msg.Callback, code)

	return nil
}
