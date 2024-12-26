package main

import (
	"log"
	"net/http"

	"loyalty-points-system-api/config"
	"loyalty-points-system-api/internal/handlers"
	"loyalty-points-system-api/pkg/middleware"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("dev")

	// Connect to the database
	db := config.ConnectDB(cfg)
	defer db.Close()

	// Set up the cron job for points expiration
	c := cron.New()
	_, err := c.AddFunc("@every 1m", func() {
		handlers.ExpirePoints(db)
	})
	if err != nil {
		log.Fatalf("Failed to schedule expiration job: %v", err)
	}
	c.Start()
	defer c.Stop()

	// Create a new ServeMux for grouping routes
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/health", handlers.HealthCheckHandler)
	mux.HandleFunc("/create-user", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUserHandler(w, r, db)
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.LoginHandler(w, r, db, cfg)
	})
	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlers.RefreshTokenHandler(w, r, db, cfg)
	})

	// Protected routes
	protectedRoutes := http.NewServeMux()
	protectedRoutes.HandleFunc("/add-transaction", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddTransactionHandler(w, r, db)
	})
	protectedRoutes.HandleFunc("/points-balance", func(w http.ResponseWriter, r *http.Request) {
		handlers.PointsBalanceHandler(w, r, db)
	})
	protectedRoutes.HandleFunc("/redeem", func(w http.ResponseWriter, r *http.Request) {
		handlers.RedeemPointsHandler(w, r, db)
	})
	protectedRoutes.HandleFunc("/points-history", func(w http.ResponseWriter, r *http.Request) {
		handlers.PointsHistoryHandler(w, r, db)
	})
	protectedRoutes.HandleFunc("/get-all-users", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetAllUsersHandler(w, r, db)
	})

	// Wrap protected routes with authentication middleware
	mux.Handle("/protected/", middleware.AuthMiddleware(http.StripPrefix("/protected", protectedRoutes)))

	// Start the server
	log.Printf("Starting server on port %s...", cfg.AppPort)
	err = http.ListenAndServe(":"+cfg.AppPort, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
