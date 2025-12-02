package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/vansh412f/cinema-reserved/internal/database"
)

type MovieResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Duration  int    `json:"duration"`
	PosterURL string `json:"poster_url"`
	ShowID    int    `json:"show_id"` // Assuming 1 show per movie for simplicity
}

type SeatResponse struct {
	ID       int     `json:"id"`
	Row      string  `json:"row"`
	Number   int     `json:"number"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"` // 'AVAILABLE', 'HELD', 'SOLD'
}

type HoldRequest struct {
	ShowID    int    `json:"show_id"`
	SeatIDs   []int  `json:"seat_ids"`
	UserEmail string `json:"user_email"`
}

type HoldResponse struct {
	BookingID int       `json:"booking_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ConfirmRequest struct {
	BookingID int `json:"booking_id"`
}

type ConfirmResponse struct {
	Status      string   `json:"status"`
	BookingCode string   `json:"booking_code"`
	MovieTitle  string   `json:"movie_title"`
	Seats       []string `json:"seats"`
}

func GetMoviesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT m.id, m.title, m.duration, m.poster_url, s.id 
		FROM movies m 
		JOIN shows s ON m.id = s.movie_id
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []MovieResponse
	for rows.Next() {
		var m MovieResponse
		if err := rows.Scan(&m.ID, &m.Title, &m.Duration, &m.PosterURL, &m.ShowID); err != nil {
			log.Println(err)
			continue
		}
		movies = append(movies, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func GetSeatsHandler(w http.ResponseWriter, r *http.Request) {
	showIDStr := r.URL.Query().Get("show_id")
	showID, err := strconv.Atoi(showIDStr)
	if err != nil {
		http.Error(w, "Invalid show_id", http.StatusBadRequest)
		return
	}

	rows, err := database.DB.Query(`
		SELECT s.id, s.row_label, s.number, c.name, c.price,
		       CASE 
		           WHEN b.status = 'CONFIRMED' THEN 'SOLD'
		           WHEN b.status = 'HELD' AND b.created_at > datetime('now', '-5 minutes') THEN 'HELD'
		           ELSE 'AVAILABLE'
		       END as status
		FROM seats s
		JOIN seat_categories c ON s.category_id = c.id
		LEFT JOIN booking_seats bs ON s.id = bs.seat_id
		LEFT JOIN bookings b ON bs.booking_id = b.id AND b.show_id = ? AND (b.status = 'CONFIRMED' OR (b.status = 'HELD' AND b.created_at > datetime('now', '-5 minutes')))
		WHERE s.hall_id = (SELECT hall_id FROM shows WHERE id = ?)
		ORDER BY s.row_label, s.number
	`, showID, showID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var seats []SeatResponse
	for rows.Next() {
		var s SeatResponse
		var status sql.NullString
		if err := rows.Scan(&s.ID, &s.Row, &s.Number, &s.Category, &s.Price, &status); err != nil {
			log.Println(err)
			continue
		}
		if status.Valid {
			s.Status = status.String
		} else {
			s.Status = "AVAILABLE"
		}
		seats = append(seats, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seats)
}

func HoldSeatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HoldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, seatID := range req.SeatIDs {
		var count int
		if err = tx.QueryRow(`
			SELECT COUNT(*) 
			FROM booking_seats bs 
			JOIN bookings b ON bs.booking_id = b.id 
			WHERE bs.seat_id = ? 
			AND (b.status = 'CONFIRMED' OR (b.status = 'HELD' AND b.created_at > datetime('now', '-5 minutes')))
		`, seatID).Scan(&count); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if count > 0 {
			http.Error(w, "One or more seats are no longer available", http.StatusConflict)
			return
		}
	}

	bookingCode := generateBookingCode()

	res, err := tx.Exec(`INSERT INTO bookings (show_id, user_email, status, booking_code, created_at) VALUES (?, ?, 'HELD', ?, datetime('now'))`, req.ShowID, req.UserEmail, bookingCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bookingID, _ := res.LastInsertId()

	for _, seatID := range req.SeatIDs {
		_, err := tx.Exec(`INSERT INTO booking_seats (booking_id, seat_id) VALUES (?, ?)`, bookingID, seatID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HoldResponse{
		BookingID: int(bookingID),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})
}

func ConfirmBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec(`
		UPDATE bookings 
		SET status = 'CONFIRMED' 
		WHERE id = ? 
		AND status = 'HELD' 
		AND created_at > datetime('now', '-5 minutes')
	`, req.BookingID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Booking expired or invalid", http.StatusConflict)
		return
	}

	var bookingCode, movieTitle string
	err = database.DB.QueryRow(`
		SELECT b.booking_code, m.title 
		FROM bookings b
		JOIN shows s ON b.show_id = s.id
		JOIN movies m ON s.movie_id = m.id
		WHERE b.id = ?
	`, req.BookingID).Scan(&bookingCode, &movieTitle)
	if err != nil {
		log.Println("Error fetching booking details:", err)
	}

	rows, err := database.DB.Query(`
		SELECT s.row_label || s.number 
		FROM booking_seats bs
		JOIN seats s ON bs.seat_id = s.id
		WHERE bs.booking_id = ?
	`, req.BookingID)
	if err != nil {
		log.Println("Error fetching seats:", err)
	}
	defer rows.Close()

	var seats []string
	for rows.Next() {
		var seat string
		rows.Scan(&seat)
		seats = append(seats, seat)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ConfirmResponse{
		Status:      "confirmed",
		BookingCode: bookingCode,
		MovieTitle:  movieTitle,
		Seats:       seats,
	})
}

func generateBookingCode() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "BOOKING"
	}
	return hex.EncodeToString(bytes)
}

func GetMyBookingsHandler(w http.ResponseWriter, r *http.Request) {
	userEmail := "test@example.com"

	rows, err := database.DB.Query(`
		SELECT b.id, b.booking_code, m.title, b.created_at
		FROM bookings b
		JOIN shows s ON b.show_id = s.id
		JOIN movies m ON s.movie_id = m.id
		WHERE b.user_email = ? AND b.status = 'CONFIRMED'
		ORDER BY b.created_at DESC
	`, userEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type BookingDTO struct {
		BookingCode string   `json:"booking_code"`
		MovieTitle  string   `json:"movie_title"`
		Seats       []string `json:"seats"`
		Date        string   `json:"date"`
	}

	var bookings []BookingDTO

	for rows.Next() {
		var bID int
		var code, title string
		var createdAt time.Time
		if err := rows.Scan(&bID, &code, &title, &createdAt); err != nil {
			log.Println(err)
			continue
		}

		seatRows, err := database.DB.Query(`
			SELECT s.row_label || s.number 
			FROM booking_seats bs
			JOIN seats s ON bs.seat_id = s.id
			WHERE bs.booking_id = ?
		`, bID)
		if err != nil {
			log.Println(err)
			continue
		}

		var seats []string
		for seatRows.Next() {
			var seat string
			seatRows.Scan(&seat)
			seats = append(seats, seat)
		}
		seatRows.Close()

		bookings = append(bookings, BookingDTO{
			BookingCode: code,
			MovieTitle:  title,
			Seats:       seats,
			Date:        createdAt.Format("2006-01-02 15:04"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}
