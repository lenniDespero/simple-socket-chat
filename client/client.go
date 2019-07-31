package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"golang.org/x/net/websocket"
	"github.com/marcusolsson/tui-go"
)

type Message struct {
	Time	time.Time	`json:"time"`
	Nick	string		`json:"nick"`
	Text	string		`json:"text"`
}

var (
	port = flag.String("port", "8888", "Port to connection")
	host = flag.String("host", "localhost", "Host to connection")
	nick = flag.String("nick", "", "Your nickname")
	myOrigin string
)

func main() {
	flag.Parse()

	ws, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	loginView := NewLoginView()
	chatView := NewChatView()
	ui, err := tui.New(loginView)
	if err != nil {
		panic(err)
	}
	loginView.OnLogin(func(username string) {
		setNickname(username)
		ui.SetWidget(chatView)
	})

	if len(*nick) > 0 {
		ui.SetWidget(chatView)
	}

	quit := func() { ui.Quit() }
	ui.SetKeybinding("Esc", quit)
	ui.SetKeybinding("Ctrl+c", quit)

	//sending messages
	chatView.OnSubmit(func(msg string) {
		if msg != "" {
			m := Message{
				Time: time.Now(),
				Nick: *nick,
				Text: msg,
			}
			err = websocket.JSON.Send(ws, m)
			if err != nil {
				ui.Update(func() {
					chatView.AddMessage(m.Time.Format("15:04:05"), "<ERROR>", err.Error())
				})
			}
		}
	})

	//Getting messages
	var m Message
	go func() {
		for {
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				chatView.AddMessage(m.Time.Format("15:04:05"), "<ERROR>", err.Error())
				break
			}
			ui.Update(func() {
				chatView.AddMessage(m.Time.Format("15:04:05"), m.Nick, m.Text)
			})
		}
	}()

	if err := ui.Run(); err != nil {
		panic(err)
	}
}

func setNickname(nickname string ) {
	nick = &nickname
}

// connect connects to the local chat server at port <port>
func connect() (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("ws://%s:%s", *host, *port), "", genMyOrigin())
}

func genMyOrigin() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return fmt.Sprintf("http://%v#%d", ipnet.IP.String(), rand.Int31())
			}
		}
	}
	return fmt.Sprintf("http://0.0.0.0", rand.Int31())
}
