package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: blog-http <port> (:80)")
	}
	var port = os.Args[1]

	var tokens uint64 = 120

	store, err := memorystore.New(&memorystore.Config{
		Tokens:   tokens,
		Interval: time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using memorystore with %d tokens per minute", tokens)

	limiter, err := httplimit.NewMiddleware(store, httplimit.IPKeyFunc())
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir(".")) // Serve files from the current directory
	http.Handle("/", limiter.Handle(Gzip(fs)))

	log.Println("Server listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
