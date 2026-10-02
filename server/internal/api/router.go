package api

import (
	"net/http"
	"os"
	"path/filepath"
)

// NewRouter constructs the HTTP handler with CORS middleware.
func NewRouter(server *Server) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Rooms endpoints (supporting both /rooms and /api/rooms prefixes)
	handleRooms := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rooms" || r.URL.Path == "/rooms/" ||
			r.URL.Path == "/api/rooms" || r.URL.Path == "/api/rooms/" {
			server.HandleCreateRoom(w, r)
			return
		}
		server.HandleRoomDispatch(w, r)
	}

	mux.HandleFunc("/rooms", handleRooms)
	mux.HandleFunc("/rooms/", handleRooms)
	mux.HandleFunc("/api/rooms", handleRooms)
	mux.HandleFunc("/api/rooms/", handleRooms)

	// Static web client serving
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			for _, p := range []string{"index.html", "../index.html", "../../index.html", "../../../index.html"} {
				if _, err := os.Stat(p); err == nil {
					http.ServeFile(w, r, p)
					return
				}
			}
		}
		for _, dir := range []string{".", "..", "../..", "../../.."} {
			filePath := filepath.Join(dir, filepath.Clean(r.URL.Path))
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				http.ServeFile(w, r, filePath)
				return
			}
		}
		http.NotFound(w, r)
	})

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
