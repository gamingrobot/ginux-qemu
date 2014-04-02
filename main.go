package main

import (
	"bufio"
	//"fmt"
	"github.com/codegangsta/martini"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	//"os"
	"os/exec"
)

func main() {
	m := martini.Classic()
	m.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		//spawn qemu
		cmd := exec.Command("qemu-system-arm", "-M", "versatilepb", "-m", "20M", "-nographic", "-readconfig", "qemu.conf")
		//cmd := exec.Command("bash")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println(err)
		}
		go readLoop(stdout, ws)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println(err)
		}
		go cmd.Run()
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				//kill qemu
				cmd.Process.Kill()
				log.Println(err)
				return
			} else {
				//send msg to qemu
				stdin.Write(message)
				//fmt.Println(string(message))
			}
		}
	})
	address := "localhost:3000"
	log.Println("Martini started on", address)
	log.Fatal(http.ListenAndServe(address, m))
}

func readLoop(output io.Reader, ws *websocket.Conn) {
	r := bufio.NewReader(output)
	for {
		str, err := r.ReadString('\n')

		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("Read Line Error:", err)
			continue
		}
		if len(str) == 0 {
			continue
		}
		//fmt.Print(str)
		ws.WriteMessage(websocket.TextMessage, []byte(str))
	}
}
