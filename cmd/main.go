package main

import (
	"log"
	"net/http"

	"loyalty-points-system-api/config"
	"loyalty-points-system-api/internal/handlers"
	"loyalty-points-system-api/pkg/middleware"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron/v3" // For scheduling the points expiration service
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("dev")

	// Connect to the database
	db := config.ConnectDB(cfg)
	defer db.Close()

	// Set up the cron job for points expiration
	c := cron.New()
	_, err := c.AddFunc("@daily", func() {
		handlers.ExpirePoints(db)
	})
	if err != nil {
		log.Fatalf("Failed to schedule expiration job: %v", err)
	}
	c.Start()
	defer c.Stop()

	// Set up routes
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.LoginHandler(w, r, db, cfg)
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlers.RefreshTokenHandler(w, r, db, cfg)
	})

	http.HandleFunc("/health", handlers.HealthCheckHandler)

	http.HandleFunc("/create-user", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUserHandler(w, r, db)
	})

	// Add Transaction API route with middleware
	http.Handle("/add-transaction", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.AddTransactionHandler(w, r, db)
	})))

	// Points Balance API route with middleware
	http.Handle("/points-balance", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.PointsBalanceHandler(w, r, db)
	})))

	// Redeem Points API route with middleware
	http.Handle("/redeem", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.RedeemPointsHandler(w, r, db)
	})))

	// Points History API route with middleware
	http.Handle("/points-history", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.PointsHistoryHandler(w, r, db)
	})))

	http.HandleFunc("/get-all-users", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetAllUsersHandler(w, r, db)
	})
	// Start the server
	log.Printf("Starting server on port %s...", cfg.AppPort)
	err = http.ListenAndServe(":"+cfg.AppPort, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
