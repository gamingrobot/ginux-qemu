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
	//"io/ioutil"
	"os/exec"
	"strconv"
	"time"
	"unicode/utf8"
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
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println(err)
		}
		go readLoop(stdout, ws)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println(err)
		}
		cmd.Start()
		limit := exec.Command("cpulimit", "-p", strconv.FormatInt(int64(cmd.Process.Pid), 10), "-l", "10") //10% usage
		limit.Start()
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
				//ws.WriteMessage(websocket.TextMessage, message)
				//fmt.Println(string(message))
			}
		}
	})
	address := "localhost:3000"
	log.Println("Martini started on", address)
	log.Fatal(http.ListenAndServe(address, m))
}

func readLoop(output io.Reader, ws *websocket.Conn) {
	reader := bufio.NewReader(output)
	buffer := []byte{}
	for {
		b, err := reader.ReadByte()
		if err == io.EOF {
			return
		}
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		buffer = append(buffer, b)
		valid := utf8.Valid(buffer)
		if valid {
			ws.WriteMessage(websocket.TextMessage, buffer)
			buffer = []byte{}
		}
	}
}
