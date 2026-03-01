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
		log.Println("⚠️ Kandinsky API: не настроен (будет использован fallback - простое наложение)")
		log.Println("   Для настройки: KANDINSKY_API_KEY и KANDINSKY_SECRET_KEY")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8085"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {

		jwtSecret = "development-secret-key-change-in-production"
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET env var for production.")
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
	// Ensure absolute path for proper file operations
	absPath, err := filepath.Abs(storagePath)
	if err == nil {
		storagePath = absPath
	}
	fileStorage := storage.NewFileStorage(storagePath)

	fashionDB, err := fashion.LoadFashionDB("./data/fashion_db.json")
	if err != nil {
		log.Printf("Warning: Failed to load fashion database: %v", err)

	}

	authService := auth.NewService(jwtSecret, db)

	handlers := api.NewHandlers(db, authService, fileStorage, fashionDB)

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Fashion Look Generator is running!"))
	})

	router.PathPrefix("/models/").Handler(
		http.StripPrefix("/models/", http.FileServer(http.Dir("./storage/models"))),
	)

	router.PathPrefix("/looks/").Handler(
		http.StripPrefix("/looks/", http.FileServer(http.Dir("./storage/looks"))),
	)

	router.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./storage/uploads"))),
	)

	router.HandleFunc("/api/register", handlers.Register).Methods("POST")
	router.HandleFunc("/api/login", handlers.Login).Methods("POST")

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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

	handler := enableCORS(router)

	log.Println("Registered routes:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		if path != "" {
			log.Printf("  %s %v", path, methods)
		}
		return nil
	})

	log.Printf("Server starting on port %s", port)
	log.Printf("Storage path: %s", storagePath)
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
