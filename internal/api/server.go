package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	router *mux.Router
	port   int
	secret string
}

// NewServer creates a new API server
func NewServer(port int, secret string) *Server {
	router := mux.NewRouter()
	
	server := &Server{
		router: router,
		port:   port,
		secret: secret,
	}
	
	server.setupRoutes()
	return server
}

// setupRoutes sets up API routes
func (s *Server) setupRoutes() {
	// CORS middleware
	s.router.Use(handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	))

	// Health check
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")
	
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	
	// Authentication
	api.HandleFunc("/auth/login", s.loginHandler).Methods("POST")
	
	// Protected routes
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(s.authMiddleware)
	
	protected.HandleFunc("/keys", s.getKeysHandler).Methods("GET")
	protected.HandleFunc("/stats", s.getStatsHandler).Methods("GET")
	protected.HandleFunc("/metrics", s.getMetricsHandler).Methods("GET")
	
	// Serve static files
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, s.router)
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// loginHandler handles login requests
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Simple authentication (in production, use proper authentication)
	if loginReq.Username == "admin" && loginReq.Password == "admin" {
		token, err := s.generateToken(loginReq.Username)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": token,
			"expires_in": 3600,
		})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// getKeysHandler handles key retrieval requests
func (s *Server) getKeysHandler(w http.ResponseWriter, r *http.Request) {
	// Mock data for now
	keys := []map[string]interface{}{
		{
			"key": "AIza***1234",
			"platform": "gemini",
			"repository": "example/repo",
			"file_path": "config.py",
			"is_valid": true,
			"discovered_at": time.Now().Format(time.RFC3339),
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"keys": keys,
		"total": len(keys),
	})
}

// getStatsHandler handles stats requests
func (s *Server) getStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Mock data for now
	stats := map[string]interface{}{
		"total_keys": 100,
		"valid_keys": 85,
		"rate_limited_keys": 15,
		"platforms": map[string]interface{}{
			"gemini": map[string]interface{}{
				"keys_found": 50,
				"valid_keys": 45,
			},
			"openrouter": map[string]interface{}{
				"keys_found": 30,
				"valid_keys": 25,
			},
			"siliconflow": map[string]interface{}{
				"keys_found": 20,
				"valid_keys": 15,
			},
		},
		"last_updated": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// getMetricsHandler handles metrics requests
func (s *Server) getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Mock data for now
	metrics := map[string]interface{}{
		"throughput_keys_per_second": 10.5,
		"cache_hit_rate": 0.85,
		"detection_rate": 0.92,
		"memory_usage_mb": 128.5,
		"uptime_seconds": 3600,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// authMiddleware handles JWT authentication
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		
		tokenString := authHeader[7:] // Remove "Bearer " prefix
		
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.secret), nil
		})
		
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// generateToken generates a JWT token
func (s *Server) generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	
	return token.SignedString([]byte(s.secret))
}