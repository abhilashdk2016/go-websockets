package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
)

var wsChan = make(chan WsPayload)
var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(jet.NewOSFileSystemLoader("./html"), jet.InDevelopmentMode())

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WebSocketConnection struct {
	*websocket.Conn
}

type WsJsonResponse struct {
	Action string `json:"action"`
	Message string `json:"message"`
	MessageType string `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action string `json:"action"`
	Message string `json:"message"`
	Conn WebSocketConnection `json:"-"`
	UserName string `json:"username"`
}

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client connected to endpoint")
	var response WsJsonResponse
	response.Message= `<em><small>Connected to Server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""
	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}
	go ListenForWs(&conn)
}

// Listen for a connection from WebSocket and send the payload to wsChan channel
func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {} else {
			payload.Conn = *conn
			wsChan <- payload
		}

	}
}

func ListenToWsChannel() {
	var response WsJsonResponse
	for {
		e := <-wsChan
		switch e.Action {
			case "username":
				clients[e.Conn] = e.UserName
				users := getUserList()
				response.Action = "list_users"
				response.ConnectedUsers = users
				broadCastToAll(response)
				break
			case "left":
				response.Action = "list_users"
				delete(clients, e.Conn)
				users := getUserList()
				response.ConnectedUsers = users
				broadCastToAll(response)
				break
			case "broadcast":
				response.Action = "broadcast"
				response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.UserName, e.Message)
				broadCastToAll(response)

		}
	}

}

func getUserList() []string {
	var userList []string
	for _, c := range clients {
		if c != "" {
			userList = append(userList, c)
		}
	}
	sort.Strings(userList)
	return userList
}

func broadCastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("Websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}