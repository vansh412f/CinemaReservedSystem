package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB(filepath string) {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables()
	seedDB()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS movies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			duration INTEGER,
			poster_url TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS halls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			total_rows INTEGER,
			total_cols INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS shows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			movie_id INTEGER,
			hall_id INTEGER,
			start_time DATETIME,
			FOREIGN KEY(movie_id) REFERENCES movies(id),
			FOREIGN KEY(hall_id) REFERENCES halls(id)
		);`,
		`CREATE TABLE IF NOT EXISTS seat_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			price REAL
		);`,
		`CREATE TABLE IF NOT EXISTS seats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hall_id INTEGER,
			row_label TEXT,
			number INTEGER,
			category_id INTEGER,
			FOREIGN KEY(hall_id) REFERENCES halls(id),
			FOREIGN KEY(category_id) REFERENCES seat_categories(id)
		);`,
		`CREATE TABLE IF NOT EXISTS bookings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			show_id INTEGER,
			user_email TEXT,
			status TEXT,
			booking_code TEXT,
			created_at DATETIME,
			FOREIGN KEY(show_id) REFERENCES shows(id)
		);`,
		`CREATE TABLE IF NOT EXISTS booking_seats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			booking_id INTEGER,
			seat_id INTEGER,
			FOREIGN KEY(booking_id) REFERENCES bookings(id),
			FOREIGN KEY(seat_id) REFERENCES seats(id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_bookings_show_status ON bookings(show_id, status);`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Fatalf("Error creating table: %v", err)
		}
	}
}

func seedDB() {
	// Check if data exists
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM movies").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count > 0 {
		return // Already seeded
	}

	log.Println("Seeding database...")

	// 1. Movies
	movies := []struct {
		Title     string
		Duration  int
		PosterURL string
	}{
		{"Inception", 148, "https://m.media-amazon.com/images/M/MV5BMjExMjkwNTQ0Nl5BMl5BanBnXkFtZTcwNTY0OTk1Mw@@._V1_.jpg"},
		{"The Dark Knight", 152, "https://image.tmdb.org/t/p/w500/qJ2tW6WMUDux911r6m7haRef0WH.jpg"},
		{"Interstellar", 169, "https://image.tmdb.org/t/p/w500/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg"},
	}

	var movieIDs []int64
	var res sql.Result
	for _, m := range movies {
		res, err = DB.Exec("INSERT INTO movies (title, duration, poster_url) VALUES (?, ?, ?)", m.Title, m.Duration, m.PosterURL)
		if err != nil {
			log.Fatal(err)
		}
		id, _ := res.LastInsertId()
		movieIDs = append(movieIDs, id)
	}

	// 2. Hall
	res, err = DB.Exec("INSERT INTO halls (name, total_rows, total_cols) VALUES (?, ?, ?)", "IMAX Hall", 9, 10)
	if err != nil {
		log.Fatal(err)
	}
	hallID, _ := res.LastInsertId()

	// 3. Shows (One for each movie)
	for _, mid := range movieIDs {
		_, err = DB.Exec("INSERT INTO shows (movie_id, hall_id, start_time) VALUES (?, ?, ?)", mid, hallID, time.Now().Add(24*time.Hour))
		if err != nil {
			log.Fatal(err)
		}
	}

	// 4. Seat Categories
	cats := []struct {
		Name  string
		Price float64
	}{
		{"Silver", 10.0},
		{"Gold", 15.0},
		{"Recliner", 25.0},
	}

	catIDs := make(map[string]int64)
	for _, c := range cats {
		res, err = DB.Exec("INSERT INTO seat_categories (name, price) VALUES (?, ?)", c.Name, c.Price)
		if err != nil {
			log.Fatal(err)
		}
		id, _ := res.LastInsertId()
		catIDs[c.Name] = id
	}

	// 5. Seats
	// Rows A-E (Silver), F-H (Gold), I (Recliner)
	// Row A: 10 seats
	// Row I: 6 seats
	// Others: 8 seats (example variation)

	rows := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
	for _, r := range rows {
		catID := catIDs["Silver"]
		if r == "F" || r == "G" || r == "H" {
			catID = catIDs["Gold"]
		} else if r == "I" {
			catID = catIDs["Recliner"]
		}

		numSeats := 8
		if r == "A" {
			numSeats = 10
		} else if r == "I" {
			numSeats = 6
		}

		for n := 1; n <= numSeats; n++ {
			_, err = DB.Exec("INSERT INTO seats (hall_id, row_label, number, category_id) VALUES (?, ?, ?, ?)", hallID, r, n, catID)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Println("Database seeded successfully.")
}
