// +build !js !wasm

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/crhntr/httplog"
)

func main() {
	errLogger := log.New(os.Stderr, "", 0)
	outLogger := log.New(os.Stdout, "", 0)
	webapp, err := NewBuildHandler(".", true)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/webapp/", webapp)
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(indexHTML))
	})
	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), httplog.Wrap(mux, httplog.JSON(outLogger, errLogger))))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
  <meta charset="utf-8">
  <title>Playground</title>
  <script src="/webapp/main.js"></script>
  <style media="screen">
  </style>
</head>
<body></body>
</html>
`
