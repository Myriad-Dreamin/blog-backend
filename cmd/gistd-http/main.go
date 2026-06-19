package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

const indexFile = "index.html"

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		log.Fatal("Usage: paseo-http <port> [root] (:80)")
	}

	port := os.Args[1]
	root := "."
	if len(os.Args) == 3 {
		root = os.Args[2]
	}

	if err := checkIndex(root); err != nil {
		log.Fatal(err)
	}

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

	mux := http.NewServeMux()
	mux.Handle("/", limiter.Handle(Gzip(newStaticHandler(root))))

	log.Printf("Paseo Server listening on %s, serving %s", port, root)
	log.Fatal(http.ListenAndServe(port, mux))
}

func checkIndex(root string) error {
	indexPath := filepath.Join(root, indexFile)
	info, err := os.Stat(indexPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("missing index.html in static root")
		}
		return err
	}
	if info.IsDir() {
		return errors.New("index.html is a directory")
	}
	return nil
}

func newStaticHandler(root string) http.Handler {
	fileServer := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldServeIndex(r.URL.Path) {
			http.ServeFile(w, r, filepath.Join(root, indexFile))
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}

func shouldServeIndex(path string) bool {
	return path == "/" || path == "/h" || strings.HasPrefix(path, "/h/")
}
