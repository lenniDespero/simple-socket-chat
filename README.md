# Simple websocket cli chat
This simple chat created with packages:
- [websocket](http://github.com/gorilla/websocket)
- [tui-go](http://github.com/marcusolsson/tui-go)

### Usage
`$ go run server/server.go [-port=8888]` - to start server

`$ go run client/client.go [-host localhost] [-port 8888] [-nick someNickName]` - запуск клиента/клиентов