package main

import (
	"crypto/sha256"
    "crypto/subtle"
	"net/http"
	"database/sql"
	"server/service"
	"fmt"
	// "os"
	"log"
	"time"
	"encoding/hex"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
    db *sql.DB
}


func main() {
    app := &application{}

	db, err := service.InitDB("alyona:suntrack@tcp(127.0.0.1:3306)/usersdb")
	if err != nil {
		fmt.Println(err)
	}
	app.db = db

  
	mux := http.NewServeMux()

	mux.HandleFunc("/register", service.Register(app.db))
	mux.HandleFunc("/read", app.basicAuth(service.Read(app.db)))
	

	   srv2 := &http.Server{
        Addr:         ":4000",
        Handler:      mux,
        IdleTimeout:  time.Minute,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
    }

	
	// mux.HandleFunc("/unprotected", app.unprotectedHandler)
    // mux.HandleFunc("/protected", app.basicAuth(app.protectedHandler))

	log.Printf("starting server on %s", srv2.Addr)
    err = srv2.ListenAndServeTLS("./localhost.pem", "./localhost-key.pem")
    log.Fatal(err)

	
}



func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var dbPassword string
		err := app.db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
		if err != nil || !checkPassword(password, dbPassword) {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
				next.ServeHTTP(w, r)
		
	})
} 

func checkPassword(inPassword, dbPassword string) bool {
	inHash := sha256.Sum256([]byte(inPassword))
	decodeDBHash, err := hex.DecodeString(dbPassword)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(inHash[:], decodeDBHash[:]) == 1
}