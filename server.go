package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/go-ping/ping"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8000", "http service address")
var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		request := string(message)
		if request[0] == 'P' {
			var num = 10
			var url = "www.google.com"
			for i := 0; i < len(request); i++ {
				if request[i] == '#' {
					anum, err := strconv.Atoi(request[1:i])
					if err != nil {
						log.Print("atoi:", err)
						os.Exit(2)
					}
					num = anum
					url = request[i+1:]
				}
			}
			go func(conn *websocket.Conn, num int, url string) {
				fmt.Println("Ping requested")
				pinger, err := ping.NewPinger(url)
				if err != nil {
					log.Print("pinger:", err)
				}
				pinger.SetPrivileged(true)
				ping_num := num
				pinger.Count = ping_num
				pinger.OnRecv = func(pkt *ping.Packet) {

					out := fmt.Sprintf("%d bytes from %s: attempt#%d time=%v\n",

						pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
					conn.WriteMessage(websocket.TextMessage, []byte(out))
				}
				pinger.OnFinish = func(stats *ping.Statistics) {
					var resp string
					out := fmt.Sprintf("%s\n", stats.Addr)
					resp += out + "\n"
					out = fmt.Sprintf("%v%% packet loss\n", stats.PacketLoss)
					resp += out + "\n"
					conn.WriteMessage(websocket.TextMessage, []byte(resp))
				}
				err = pinger.Run() // Blocks until finished.
				if err != nil {
					log.Print("pinger:", err)
				}
			}(c, num, url)
		}
		if request[0] == 'T' {
			url := request[2:]
			go func(conn *websocket.Conn, url string) {
				fmt.Println("Trace requested")
				command := exec.Command("cmd", "/C", "tracert", url)
				data, err := command.Output()
				if err != nil {
					fmt.Println("Error: ", err)
				}
				fmt.Println("Trace completed")
				start := 0
				for i := 0; i < len(data); i++ {
					if data[i] == byte(10) {
						conn.WriteMessage(websocket.TextMessage, []byte(string(data[start:i])+"\n"))
						start = i + 1
						time.Sleep(100 * time.Millisecond)
					}
				}
			}(c, url)
		}
	}
}

func main() {
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
