package main

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type user chan string

var (
	connecting 	= 	make(chan user)
	leaving 	= 	make(chan user)
	message 	= 	make(chan string)
)

func broadcaster(){
	users := make(map[user]bool)
	for {
		select {
		case con := <- connecting:
			users[con] = true
		case cli := <- leaving:
			delete(users, cli)
			close(cli)
		case msg := <- message:
			for cli := range users {
				cli <- msg
			}
		}
	}
}

func EchoServer(ws *websocket.Conn) {
	lenBuf := make([]byte, 5)

	log.Print("Web socket connection " + ws.RemoteAddr().String())
	defer log.Print("Web socket disconnection " + ws.RemoteAddr().String())

	go func() {
		for msg := range message {
			ws.Write([]byte(msg))
		}
	}()

	for {
		_, err := ws.Read(lenBuf)
		if err != nil {
			log.Println("Error: ", err.Error())
			return
		}

		length, _ := strconv.Atoi(strings.TrimSpace(string(lenBuf)))
		buf := make([]byte, length)
		_, e := ws.Read(buf)
		if e != nil {
			log.Println("My be your message > " + string(length))
			return
		}
		//ws.Write([]byte(buf))

		message <- string(buf)
		//fmt.Println(string(buf))
	}
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "template/index.html")
}

func main()  {
	http.Handle("/websocket", websocket.Handler(EchoServer))
	http.HandleFunc("/", MainPage)
	http.HandleFunc("/chat.js", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./static/chat.js")
	})
	err := http.ListenAndServe(":8012", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

	go broadcaster()
}