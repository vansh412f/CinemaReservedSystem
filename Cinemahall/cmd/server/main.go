package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vansh412f/cinema-reserved/internal/database"
	"github.com/vansh412f/cinema-reserved/internal/handlers"
)

func main() {
	// Initialize Database
	database.InitDB("cinema.db")
	defer database.DB.Close()

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start Cleanup Routine
	// We pass the context so it can stop when main context is done
	go cleanupRoutine(ctx)

	// Serve Static Files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API Routes
	http.HandleFunc("/api/movies", handlers.GetMoviesHandler)
	http.HandleFunc("/api/my-bookings", handlers.GetMyBookingsHandler)
	http.HandleFunc("/api/seats", handlers.GetSeatsHandler)
	http.HandleFunc("/api/hold-seats", handlers.HoldSeatsHandler)
	http.HandleFunc("/api/confirm-booking", handlers.ConfirmBookingHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: nil, // Use DefaultServeMux
	}

	// Start Server in a goroutine
	go func() {
		log.Println("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Create a deadline to wait for current operations to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func cleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping cleanup routine...")
			return
		case <-ticker.C:
			log.Println("Running cleanup routine...")
			// Delete or expire held bookings older than 5 minutes
			_, err := database.DB.Exec(`
				UPDATE bookings 
				SET status = 'EXPIRED' 
				WHERE status = 'HELD' 
				AND created_at < datetime('now', '-5 minutes')
			`)
			if err != nil {
				log.Printf("Error in cleanup routine: %v", err)
			}
		}
	}
}
