package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var (
	kamiyaUrl            = "https://p0.kamiya.dev/api/openai/chat/completions"
	httpAuthorizationKey = "Authorization"
	port                 = 12375
)

func getAuthorizationHeader(h http.Header) string {
	return h.Get(httpAuthorizationKey)
}

func setAuthorizationHeader(h http.Header, key string) {
	h.Set(httpAuthorizationKey, key)
}

func setupHttpRequest(key string, body io.ReadCloser) (req *http.Request, err error) {
	req, err = http.NewRequest(http.MethodPost, kamiyaUrl, body)
	setAuthorizationHeader(req.Header, key)
	req.Header.Add("Content-Type", "application/json")
	return
}

func extracBytes(reader io.Reader) (buf []byte, err error) {
	buf, err = io.ReadAll(reader)
	if err != nil {
		return
	}
	return
}

func main() {
	if len(os.Args) > 1 && os.Args[1] != "" {
		tmp, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Printf("%v", err)
			return
		}
		port = tmp
	}
	clientPool := &sync.Pool{
		New: func() interface{} {
			return &http.Client{}
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		clientWarp := clientPool.Get()
		client := clientWarp.(*http.Client)
		defer clientPool.Put(client)
		key := getAuthorizationHeader(r.Header)
		req, err := setupHttpRequest(key, r.Body)
		defer func() {
			if err != nil {
				log.Printf("Err %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}()
		log.Printf("Using key %v", getAuthorizationHeader(req.Header))
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		buf, err := extracBytes(resp.Body)
		if err != nil {
			return
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(buf)
		log.Printf("%v", string(buf))

	})
	listen := fmt.Sprintf("localhost:%v", port)
	log.Printf("Listen on %v", listen)
	err := http.ListenAndServe(listen, mux)
	if err != nil {
		log.Panicf("%v", err)
		return
	}
}
