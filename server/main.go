package main

import (
	"fmt"
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

	frame := Frame{
		PayloadData: []byte("say hello to client"),
		Opcode:      0x80 | 0x1, //Set FIN bit is 1
	}

	err = ws.Send(frame)

	if err != nil {
		log.Println(err)
		return
	}

	for {
		frameRecv, err := ws.Recv()
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println("Client say: ", string(frameRecv.PayloadData))
	}

}

func main() {
	http.HandleFunc("/ping", pingHanler)
	http.HandleFunc("/ws", websocketHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
