package main

import (
    "fmt"
    "net/http"
)

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func resetAlarmHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        fmt.Println("Allarme resettato!")
        w.Write([]byte("OK"))
    } else {
        http.Error(w, "Metodo non consentito", http.StatusMethodNotAllowed)
    }
}

func main() {
    http.Handle("/api/reset-alarm", corsMiddleware(http.HandlerFunc(resetAlarmHandler)))
    fmt.Println("API in ascolto su :8080")
    http.ListenAndServe(":8080", nil)
}