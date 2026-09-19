package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"pollingapp/backend/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type pollHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var hubs = make(map[string]*pollHub)
var hubsMu sync.Mutex

func getHub(pollId string) *pollHub {
	hubsMu.Lock()
	defer hubsMu.Unlock()

	hub, exists := hubs[pollId]
	if !exists {
		hub = &pollHub{clients: make(map[*websocket.Conn]bool)}
		hubs[pollId] = hub
		go subscribeToPoll(pollId, hub)
	}

	return hub
}

func subscribeToPoll(pollId string, hub *pollHub) {
	sub := config.RedisClient.Subscribe(config.Ctx, "poll:"+pollId)
	ch := sub.Channel()

	for msg := range ch {
		hub.mu.Lock()
		for conn := range hub.clients {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				conn.Close()
				delete(hub.clients, conn)
			}
		}
		hub.mu.Unlock()
	}
}

func PublishUpdate(pollId string, data gin.H) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return
	}

	config.RedisClient.Publish(config.Ctx, "poll:"+pollId, string(bytes))
}

func PollSocket(c *gin.Context) {
	pollId := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade failed:", err)
		return
	}

	hub := getHub(pollId)
	hub.mu.Lock()
	hub.clients[conn] = true
	hub.mu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			hub.mu.Lock()
			delete(hub.clients, conn)
			hub.mu.Unlock()
			conn.Close()
			break
		}
	}
}
