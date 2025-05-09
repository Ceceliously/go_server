package service

import (
	"database/sql"
	"fmt"
	// "io"
	"net/http"
	"sync"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

type srv struct {
	mu *sync.RWMutex
	db *sql.DB
}



func (s *srv) Register(w http.ResponseWriter,r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	firstName := r.URL.Query().Get("first_name")
	lastName := r.URL.Query().Get("last_name")
	age := r.URL.Query().Get("age")
	gender := r.URL.Query().Get("gender")

	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Name and password are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if count > 0 {
		w.WriteHeader(http.StatusConflict)
		log.Printf("This user already exists")
		return
	}

	_, err = s.db.Exec("INSERT INTO users (username, password, first_name, last_name, age, gender) VALUES (?, ?, ?, ?, ?, ?)", username, password, firstName, lastName, age, gender)
	if err != nil {
		log.Printf("Database indert query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("User %v successfully registrated", username)
}

func (s *srv) Read(w http.ResponseWriter,r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Username and password are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var firstName, lastName, gender string
	var age int
	err := s.db.QueryRow("SELECT first_name, last_name, age, gender FROM users WHERE username = ? AND password = ?", username, password).Scan(&firstName, &lastName, &age, &gender)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			log.Printf("User %v not found or password %v is invalid", username, password)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Database query error: %v", err)
		return
	}

	res := fmt.Sprintf("%s: %s %s, %d, %s. Pass: %s", username, firstName, lastName, age, gender, password)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(res))
	if err != nil {
		log.Printf("Failed to return response: %v", err)
	}
}

func NewDB() (*srv, error) {
	db, err := sql.Open("mysql", "alyona:suntrack@tcp(127.0.0.1:3306)/usersdb")
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, username VARCHAR(50), password VARCHAR(50) NOT NULL, first_name CHAR(30), last_name CHAR(30), age INTEGER, gender CHAR(1))`); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &srv{
		mu: &sync.RWMutex{},
		db: db,
	}, nil
}