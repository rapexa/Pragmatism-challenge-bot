package server

import (
	"log"
	"net/http"
	"path/filepath"
)

type FileServer struct {
	uploadPath string
	port       string
}

func NewFileServer(uploadPath, port string) *FileServer {
	return &FileServer{
		uploadPath: uploadPath,
		port:       port,
	}
}

// Start starts the HTTP file server in background
func (s *FileServer) Start() {
	// Serve uploaded files
	fs := http.FileServer(http.Dir(s.uploadPath))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("File server starting on port %s", s.port)
	absPath, _ := filepath.Abs(s.uploadPath)
	log.Printf("Files will be served from: %s", absPath)

	go func() {
		if err := http.ListenAndServe(":"+s.port, nil); err != nil {
			log.Printf("Error starting file server: %v", err)
		}
	}()
}
