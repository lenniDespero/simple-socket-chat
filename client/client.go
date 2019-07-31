package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
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
	host = flag.String("host", "8888", "Host to connection")
	nick = flag.String("nick", "", "Your nickname")
)

func main() {
	flag.Parse()

	// connect
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
					chatView.AddMessage(m.Time.Format("13:13:13"), "<ERROR>", err.Error())
				})
			}
		}
	})
	
	// receive
	var m Message
	go func() {
		for {
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				chatView.AddMessage(m.Time.Format("13:13:13"), "<ERROR>", err.Error())
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
	fmt.Println(nick)
	nick = &nickname
	fmt.Println(nick)
}

// connect connects to the local chat server at port <port>
func connect() (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("ws://localhost:%s", *port), "", mockedIP())
}

// mockedIP is a demo-only utility that generates a random IP address for this client
func mockedIP() string {
	var arr [4]int
	for i := 0; i < 4; i++ {
		rand.Seed(time.Now().UnixNano())
		arr[i] = rand.Intn(256)
	}
	return fmt.Sprintf("http://%d.%d.%d.%d", arr[0], arr[1], arr[2], arr[3])
}

type LoginHandler func(string)

type LoginView struct {
	tui.Box
	frame        *tui.Box
	loginHandler LoginHandler
}

func NewLoginView() *LoginView {
	// https://github.com/marcusolsson/tui-go/blob/master/example/login/main.go
	user := tui.NewEntry()
	user.SetFocused(true)
	user.SetSizePolicy(tui.Maximum, tui.Maximum)

	label := tui.NewLabel("Enter your name: ")
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

type SubmitMessageHandler func(string)

type ChatView struct {
	tui.Box
	frame    *tui.Box
	history  *tui.Box
	onSubmit SubmitMessageHandler
}

func NewChatView() *ChatView {
	view := &ChatView{}

	// ref: https://github.com/marcusolsson/tui-go/blob/master/example/chat/main.go
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