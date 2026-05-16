package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"

	"fake-review-ai2/config"
	"fake-review-ai2/database"
	"fake-review-ai2/handlers"
	"fake-review-ai2/middleware"
	"fake-review-ai2/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if err := database.Init(cfg); err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	if err := database.DeleteUnverifiedUsers(); err != nil {
		log.Printf("Initial cleanup error: %v", err)
	} else {
		log.Println("Initial cleanup of unverified users completed")
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			if err := database.DeleteUnverifiedUsers(); err != nil {
				log.Printf("Scheduled cleanup error: %v", err)
			} else {
				log.Println("Scheduled cleanup of unverified users completed")
			}
		}
	}()

	frontendPath := findFrontendPath()
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")

	emailService := services.NewEmailService()
	jwtService := services.NewJWTService(cfg.JWT.Secret)
	authService := services.NewAuthService(emailService, jwtService, cfg.EmailVerificationEnabled)
	analysisService := services.NewAnalysisService(cfg.PythonServer.Host, cfg.PythonServer.Port)
	wbParserService := services.NewWBParserService()

	authHandler := handlers.NewAuthHandler(authService)
	analysisHandler := handlers.NewAnalysisHandler(analysisService, wbParserService)
	adminHandler := handlers.NewAdminHandler()

	router := mux.NewRouter()

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(frontendPath))))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(frontendPath, "index.html"))
	})

	router.HandleFunc("/api/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/verify", authHandler.Verify).Methods("POST")
	router.HandleFunc("/api/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/analyze", analysisHandler.AnalyzeReview).Methods("POST")
	router.HandleFunc("/admin.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(frontendPath, "admin.html"))
	})

	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.HandleFunc("/profile", authHandler.Profile).Methods("GET")
	protected.HandleFunc("/analyze_product", analysisHandler.AnalyzeProduct).Methods("POST")
	protected.HandleFunc("/account", authHandler.DeleteAccount).Methods("DELETE")
	protected.HandleFunc("/history", analysisHandler.GetHistory).Methods("GET")

	admin := router.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AuthMiddleware(jwtService))
	admin.Use(middleware.RequireRole("admin", "super_admin"))
	admin.HandleFunc("/users", adminHandler.ListUsers).Methods("GET")
	admin.HandleFunc("/users/{id:[0-9]+}", adminHandler.DeleteUser).Methods("DELETE")
	admin.HandleFunc("/products", adminHandler.ListProductAnalysis).Methods("GET")
	admin.HandleFunc("/products/{id:[0-9]+}", adminHandler.DeleteProductAnalysis).Methods("DELETE")
	admin.HandleFunc("/stats", adminHandler.GetStats).Methods("GET")

	superAdmin := router.PathPrefix("/admin").Subrouter()
	superAdmin.Use(middleware.AuthMiddleware(jwtService))
	superAdmin.Use(middleware.RequireRole("super_admin"))
	superAdmin.HandleFunc("/users/role", adminHandler.UpdateUserRole).Methods("PUT")

	addr := fmt.Sprintf(":%d", cfg.GoServer.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Go server started on http://localhost%s", addr)
	log.Fatal(server.ListenAndServe())
}

func findFrontendPath() string {
	for _, cand := range []string{"../static", "./static", "static"} {
		if info, err := os.Stat(cand); err == nil && info.IsDir() {
			return cand
		}
	}
	return "../static"
}
