# 🎬 Cinema Reserved

> A modern cinema ticket booking system built with Go and vanilla JavaScript

Cinema Reserved is a lightweight, efficient ticket reservation platform that provides seamless movie booking experiences with real-time seat availability and secure transaction handling.

---

## ✨ Features

- 🎥 **Movie Catalog** - Browse through available movies with detailed information
- 🪑 **Interactive Seat Map** - Visual seat selection with real-time availability status
- ⏱️ **Smart Booking System** - Reserve seats with a 5-minute hold period before confirmation
- 🎫 **Booking History** - Access your ticket history and booking details
- 🔒 **Concurrency Safe** - Prevents double-booking with robust transaction handling
- 💾 **SQLite Database** - Lightweight, file-based database with automatic schema creation

---

## 🚀 Quick Start

### Prerequisites

- [Go](https://go.dev/dl/) (version 1.25 or higher)
- Make (optional, for using the Makefile)

### Installation & Setup

1. **Install Dependencies**
   ```bash
   make deps
   # Or manually:
   go mod download
   ```

2. **Run the Application**
   ```bash
   make run
   # Or manually:
   go run ./cmd/server
   ```
   
   The server will start at [http://localhost:8080](http://localhost:8080)
   
   > 💡 The database (`cinema.db`) is automatically created and seeded with sample data on first run

3. **Build the Application**
   ```bash
   make build
   # Or manually:
   go build -o bin/server ./cmd/server
   ```

---

## 🧹 Cleanup

Remove compiled binaries and database files:

```bash
make clean
```

---

## 📁 Project Structure

```
cinema-reserved/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── database/        # Database initialization and schema
│   ├── handlers/        # HTTP API handlers
│   └── models/          # Data structures
├── static/              # Frontend assets (HTML, CSS, JS)
│   ├── css/
│   ├── js/
│   └── *.html
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
└── makefile             # Build automation
```

---

## 🛠️ Technology Stack

- **Backend**: Go 1.25+
- **Database**: SQLite3
- **Frontend**: Vanilla JavaScript, HTML5, CSS3
- **Architecture**: RESTful API with static file serving

---

## 📝 License

This project is open source and available for use.

---

**Built with ❤️ for cinema enthusiasts**
