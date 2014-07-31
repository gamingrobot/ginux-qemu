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
	"sync/atomic"
	"time"
	"unicode/utf8"
)

const MAX_VMS = 20

var currentVms int64 = 0

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
		if currentVms >= MAX_VMS {
			http.Error(w, "Over Capacity", 400)
			return
		}
		//spawn qemu
		cmd := exec.Command("qemu-system-arm", "-M", "versatilepb", "-m", "20M", "-nographic", "-readconfig", "qemu.conf")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println(err)
			return
		}
		go readLoop(stdout, ws)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println(err)
			return
		}
		cmd.Start()
		atomic.AddInt64(&currentVms, 1)
		limit := exec.Command("cpulimit", "-p", strconv.FormatInt(int64(cmd.Process.Pid), 10), "-l", "10") //10% usage
		limit.Start()
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				//kill qemu
				cmd.Process.Kill()
				cmd.Process.Wait()
				limit.Process.Wait() //dont zombie process
				atomic.AddInt64(&currentVms, -1)
				//log.Println(err)
				return
			} else {
				//send msg to qemu
				stdin.Write(message)
			}
		}
	})
	log.Println("Martini started")
	log.Fatal(http.ListenAndServe(":3000", m))
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
