package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fashion-look-generator/internal/api"
	"fashion-look-generator/internal/auth"
	"fashion-look-generator/internal/database"
	"fashion-look-generator/internal/fashion"
	"fashion-look-generator/internal/storage"

	"github.com/gorilla/mux"
)

func main() {
	kandinskyAPIKey := os.Getenv("KANDINSKY_API_KEY")
	kandinskySecretKey := os.Getenv("KANDINSKY_SECRET_KEY")

	if kandinskyAPIKey != "" && kandinskySecretKey != "" {
		log.Println("✅ Kandinsky API: настроен")
	} else {
		log.Println("⚠️ Kandinsky API: не настроен")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8085"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "development-secret-key-change-in-production"
		log.Println("WARNING: Using default JWT secret")
	}

	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./storage"
	}
	absPath, _ := filepath.Abs(storagePath)
	storagePath = absPath

	fileStorage := storage.NewFileStorage(storagePath)

	fashionDB, _ := fashion.LoadFashionDB("./data/fashion_db.json")

	authService := auth.NewService(jwtSecret, db)
	handlers := api.NewHandlers(db, authService, fileStorage, fashionDB)

	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/register", handlers.Register).Methods("POST")
	router.HandleFunc("/api/login", handlers.Login).Methods("POST")
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.Middleware(authService))
	apiRouter.HandleFunc("/upload-item", handlers.UploadItem).Methods("POST")
	apiRouter.HandleFunc("/add-item-link", handlers.AddItemLink).Methods("POST")
	apiRouter.HandleFunc("/user-items", handlers.GetUserItems).Methods("GET")
	apiRouter.HandleFunc("/item/{id}", handlers.DeleteItem).Methods("DELETE")
	apiRouter.HandleFunc("/generate-looks", handlers.GenerateLooks).Methods("POST")
	apiRouter.HandleFunc("/user-looks", handlers.GetUserLooks).Methods("GET")
	apiRouter.HandleFunc("/look/{id}", handlers.GetLook).Methods("GET")
	apiRouter.HandleFunc("/look/{id}", handlers.UpdateLook).Methods("PATCH")
	apiRouter.HandleFunc("/look/{id}", handlers.DeleteLook).Methods("DELETE")
	apiRouter.HandleFunc("/models", handlers.GetModels).Methods("GET")

	// Static files - order matters! More specific paths first
	router.PathPrefix("/models/").Handler(http.StripPrefix("/models/", http.FileServer(http.Dir("./storage/models"))))
	router.PathPrefix("/looks/").Handler(http.StripPrefix("/looks/", http.FileServer(http.Dir("./storage/looks"))))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./storage/uploads"))))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./dist/assets"))))

	// Serve all other static files from dist (including landing.jpg)
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./dist"))))

	// SPA fallback - serve index.html for root
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./dist/index.html")
		}
	})

	handler := enableCORS(router)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
