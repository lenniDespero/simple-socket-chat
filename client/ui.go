package main

import (
	"fmt"
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
