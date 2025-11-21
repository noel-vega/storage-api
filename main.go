package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	storageRootPath := os.Getenv("STORAGE_ROOT_PATH")
	if storageRootPath == "" {
		log.Fatal("STORAGE_ROOT_PATH not set")
	}

	info, err := os.Stat(storageRootPath)
	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal("storage root path must be a directory")
	}

	log.Println("===============================================")
	log.Println("Storage API Starting...")
	log.Println("===============================================")
	log.Printf("Configuration:")
	log.Printf("	- Port: %s", port)
	log.Printf("	- Storage Path: %s", storageRootPath)
	log.Printf("	- Go Version: %s", runtime.Version())
	log.Println("===============================================")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("storage online"))
	})

	r.Post("/contents", ListContentsHandler(storageRootPath))

	log.Printf("Server listening on http://0.0.0.0:%s", port)
	log.Println("Press Ctrl+c to stop")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func ListContentsHandler(storageRootPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Path string `json:"path"`
		}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		cleanPath := filepath.Clean(body.Path)
		cleanPath = strings.TrimPrefix(cleanPath, "/")

		if strings.Contains(cleanPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(cleanPath) == "" {
			cleanPath = "."
		}

		path := filepath.Join(storageRootPath, cleanPath)

		entries, err := os.ReadDir(path)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "Path not found", http.StatusNotFound)
			} else if os.IsPermission(err) {
				http.Error(w, "Permission denied", http.StatusForbidden)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		type Item struct {
			Name  string `json:"name"`
			IsDir bool   `json:"isDir"`
			Size  int64  `json:"size"`
		}

		items := []Item{}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, Item{Name: info.Name(), IsDir: info.IsDir(), Size: info.Size()})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}
}
