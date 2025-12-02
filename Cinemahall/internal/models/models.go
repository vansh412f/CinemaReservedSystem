package models

import "time"

// Structs
type Movie struct {
	ID        int
	Title     string
	Duration  int
	PosterURL string
}

type Hall struct {
	ID        int
	Name      string
	TotalRows int
	TotalCols int
}

type Show struct {
	ID        int
	MovieID   int
	HallID    int
	StartTime time.Time
}

type SeatCategory struct {
	ID    int
	Name  string
	Price float64
}

type Seat struct {
	ID         int
	HallID     int
	RowLabel   string
	Number     int
	CategoryID int
}

type Booking struct {
	ID          int
	ShowID      int
	UserEmail   string
	Status      string // 'HELD', 'CONFIRMED', 'EXPIRED'
	BookingCode string
	CreatedAt   time.Time
}

type BookingSeat struct {
	ID        int
	BookingID int
	SeatID    int
}
