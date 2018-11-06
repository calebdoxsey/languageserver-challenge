package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(0)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:5000"
	}
	log.Println("starting http server on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
