package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/koketama/auth"
	"go.uber.org/zap"
)

func newRest(host string) {
	restNormal(host)
	restError(host)
	restPanic(host)
}

func restNormal(host string) {
	fmt.Println("---------------------------------------------------------")

	payload := []byte(`{"serial_key":"00000111","message":"normal"}`)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/signup/0987654321", host), bytes.NewReader(payload))
	if err != nil {
		logger.Fatal("rest normal new request err", zap.Error(err))
	}

	proxyAuthorization, date, err := signature.Generate("webapi", auth.MethodPost, "/v1/signup/0987654321", payload)
	if err != nil {
		logger.Fatal("rest normal do signature err", zap.Error(err))
	}

	req.Header.Set("Authorization", "dummy token")
	req.Header.Set("Proxy-Authorization", proxyAuthorization)
	req.Header.Set("Date", date)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatal("rest normal do request err", zap.Error(err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal("rest normal read body err", zap.Error(err))
	}

	logger.Info("rest normal", zap.String("resp", string(body)))
}

func restError(host string) {
	fmt.Println("---------------------------------------------------------")

	payload := []byte(`{"serial_key":"00000111","message":"error"}`)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/signup/0987654321", host), bytes.NewReader(payload))
	if err != nil {
		logger.Fatal("rest error new request err", zap.Error(err))
	}

	authorization, date, err := signature.Generate("webapi", auth.MethodPost, "/v1/signup/0987654321", payload)
	if err != nil {
		logger.Fatal("rest error do signature err", zap.Error(err))
	}

	req.Header.Set("Authorization", "dummy token")
	req.Header.Set("Proxy-Authorization", authorization)
	req.Header.Set("Date", date)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatal("rest error do request err", zap.Error(err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal("rest error read body err", zap.Error(err))
	}

	logger.Info("rest error", zap.String("resp", string(body)))
}

func restPanic(host string) {
	fmt.Println("---------------------------------------------------------")

	payload := []byte(`{"serial_key":"00000111","message":"panic"}`)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/signup/0987654321", host), bytes.NewReader(payload))
	if err != nil {
		logger.Fatal("rest panic new request err", zap.Error(err))
	}

	authorization, date, err := signature.Generate("webapi", auth.MethodPost, "/v1/signup/0987654321", payload)
	if err != nil {
		logger.Fatal("rest panic do signature err", zap.Error(err))
	}

	req.Header.Set("Authorization", "dummy token")
	req.Header.Set("Proxy-Authorization", authorization)
	req.Header.Set("Date", date)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatal("rest panic do request err", zap.Error(err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal("rest panic read body err", zap.Error(err))
	}

	logger.Info("rest panic", zap.String("resp", string(body)))
}

func restDummy(ctx context.Context, host string, goroutines, payloadSize int) {
	wg := new(sync.WaitGroup)
	wg.Add(goroutines)

	const template = `{"track_id":"%s","message":"%s", "ts":"%s"}`
	do := func() {
		buf := make([]byte, 10)
		io.ReadFull(rand.Reader, buf)
		trackID := hex.EncodeToString(buf)

		buf = make([]byte, payloadSize)
		io.ReadFull(rand.Reader, buf)
		message := hex.EncodeToString(buf)

		payload := []byte(fmt.Sprintf(template, trackID, message, time.Now().Format(time.RFC3339)))
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/dummy", host), bytes.NewReader(payload))
		if err != nil {
			logger.Error("rest dummy new request err", zap.Error(err))
			return
		}

		proxyAuthorization, date, err := signature.Generate("webapi", auth.MethodPost, "/v1/dummy", payload)
		if err != nil {
			logger.Error("rest dummy do signature err", zap.Error(err))
			return
		}

		req.Header.Set("Authorization", "dummy token")
		req.Header.Set("Proxy-Authorization", proxyAuthorization)
		req.Header.Set("Date", date)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error("rest dummy do request err", zap.Error(err))
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error("rest dummy read body err", zap.Error(err))
			return
		}

		if resp.StatusCode != http.StatusOK {
			logger.Error("rest dummy", zap.String("resp", string(body)))
		}
	}

	for k := 0; k < goroutines; k++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					do()
				}
			}
		}()
	}

	wg.Wait()
}
