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

type LoginHandler func(string)
type SubmitMessageHandler func(string)

type LoginView struct {
	tui.Box
	frame        *tui.Box
	loginHandler LoginHandler
}

type ChatView struct {
	tui.Box
	frame    *tui.Box
	history  *tui.Box
	onSubmit SubmitMessageHandler
}

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
				return fmt.Sprintf("http://%v/#%d", ipnet.IP.String(), rand.Int31())
			}
		}
	}
	return fmt.Sprintf("http://0.0.0.0/#", rand.Int31())
}

func NewLoginView() *LoginView {
	user := tui.NewEntry()
	user.SetFocused(true)
	user.SetSizePolicy(tui.Maximum, tui.Maximum)

	label := tui.NewLabel("Enter your nick: ")
	user.SetSizePolicy(tui.Expanding, tui.Maximum)

	userBox := tui.NewHBox(
		label,
		user,
	)
	userBox.SetBorder(true)
	userBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	view := &LoginView{}
	view.frame = tui.NewVBox(
		tui.NewSpacer(),
		tui.NewPadder(-4, 0, tui.NewPadder(4, 0, userBox)),
		tui.NewSpacer(),
	)
	view.Append(view.frame)

	user.OnSubmit(func(e *tui.Entry) {
		if e.Text() != "" {
			if view.loginHandler != nil {
				view.loginHandler(e.Text())
			}
			e.SetText("")
		}
	})

	return view
}

func (v *LoginView) OnLogin(handler LoginHandler) {
	v.loginHandler = handler
}

func NewChatView() *ChatView {
	view := &ChatView{}
	view.history = tui.NewVBox()
	historyScroll := tui.NewScrollArea(view.history)
	historyScroll.SetAutoscrollToBottom(true)
	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	input.OnSubmit(func(e *tui.Entry) {
		if e.Text() != "" {
			if view.onSubmit != nil {
				view.onSubmit(e.Text())
			}
			e.SetText("")
		}
	})

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)
	view.frame = tui.NewVBox(
		historyBox,
		inputBox,
	)
	view.frame.SetBorder(false)
	view.Append(view.frame)

	return view
}

func (c *ChatView) OnSubmit(handler SubmitMessageHandler) {
	c.onSubmit = handler
}

func (c *ChatView) AddMessage(time string, user string, msg string) {
	c.history.Append(
		tui.NewHBox(
			tui.NewLabel(fmt.Sprintf("%v %v: %v", time, user, msg)),
		),
	)
}
