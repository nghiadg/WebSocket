package main

import (
	"log"
	"net/http"
)

// Ping
func pingHanler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong"))
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := New(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	err = ws.Handshake()
	if err != nil {
		log.Println(err)
		return
	}

}

func main() {
	http.HandleFunc("/ping", pingHanler)
	http.HandleFunc("/ws", websocketHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
