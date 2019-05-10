// +build !js !wasm

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/crhntr/httplog"
	"github.com/crhntr/ot"
)

func main() {
	errLogger := log.New(os.Stderr, "", 0)
	outLogger := log.New(os.Stdout, "", 0)
	webapp, err := NewBuildHandler(".", true)
	if err != nil {
		log.Fatal(err)
	}
	server := &ot.Server{
		RegisterListeners: make(chan chan []ot.Applier),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webapp/", webapp)
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(indexHTML))
	})
	rootMux := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/ws" {
			log.Printf("WS connect: %s", req.RemoteAddr)
			defer log.Println("WS close: %s", req.RemoteAddr)
			server.HandlerWS(res, req)
		} else {
			httplog.Wrap(mux, httplog.JSON(outLogger, errLogger)).ServeHTTP(res, req)
		}
	})
	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), rootMux))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
  <meta charset="utf-8">
  <title>Markdown</title>
  <script src="/webapp/main.js"></script>
  <style media="screen">
  </style>
</head>
<body></body>
</html>
`
