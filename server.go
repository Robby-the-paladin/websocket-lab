package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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
			fmt.Println("Ping requested")
			pinger, err := ping.NewPinger("www.google.com")
			pinger.SetPrivileged(true)
			if err != nil {
				panic(err)
			}
			ping_num, err := strconv.Atoi(request[1:])
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			pinger.Count = ping_num
			pinger.OnRecv = func(pkt *ping.Packet) {

				out := fmt.Sprintf("%d bytes from %s: attempt#%d time=%v\n",

					pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
				c.WriteMessage(websocket.TextMessage, []byte(out))
				//fmt.Println(pkt)
			}
			pinger.OnFinish = func(stats *ping.Statistics) {
				var resp string
				out := fmt.Sprintf("%s\n", stats.Addr)
				resp += out + "\n"
				out = fmt.Sprintf("%v%% packet loss\n", stats.PacketLoss)
				resp += out + "\n"
				c.WriteMessage(websocket.TextMessage, []byte(resp))
			}
			err = pinger.Run() // Blocks until finished.
			if err != nil {
				log.Print("pinger:", err)
			}
		}
	}
}

func main() {
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
