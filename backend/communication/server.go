package communication

import (
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"

	"golang.org/x/net/websocket"
	"superstellar/backend/events"
	"superstellar/backend/monitor"
	"superstellar/backend/utils"
)

// Server struct holds server variables.
type Server struct {
	pattern          string
	clients          map[uint32]*Client
	monitor          *monitor.Monitor
	clientID         uint32
	eventsDispatcher *events.EventDispatcher
	clientIdSeq      *utils.IdSequencer
}

// NewServer initializes a new server.
func NewServer(pattern string, monitor *monitor.Monitor, eventDispatcher *events.EventDispatcher, clientIdSeq *utils.IdSequencer) *Server {
	return &Server{
		pattern:      pattern,
		clients:      make(map[uint32]*Client),
		monitor:      monitor,
		clientID:     0,
		eventsDispatcher: eventDispatcher,
		clientIdSeq:  clientIdSeq,
	}
}

// Listen runs puts server into listening mode.
func (s *Server) Listen() {
	log.Println("Listening server...")

	onConnected := func(ws *websocket.Conn) {
		defer func() {
			err := ws.Close()
			if err != nil {
				log.Println("Error:", err.Error())
			}
		}()

		client := NewClient(ws, s.monitor, s.eventsDispatcher, s.clientIdSeq.NextId())
		s.clients[client.id] = client

		log.Println("Added new client. Now", len(s.clients), "clients connected.")
		client.Listen()
	}

	http.Handle(s.pattern, websocket.Handler(onConnected))
}

func (s *Server) SendToAllClients(message proto.Message) {
	bytes := marshalMessage(message)
	for _, c := range s.clients {
		c.SendMessage(bytes)
	}
}

func (s *Server) SendToClient(clientID uint32, message proto.Message) {
	bytes := marshalMessage(message)

	client, ok := s.clients[clientID]
	if ok {
		client.SendMessage(bytes)
	} else {
		log.Println("Client %i not found", clientID)
		return
	}
}

func (s *Server) HandleUserLeft(userLeftEvent *events.UserLeft) {
	delete(s.clients, userLeftEvent.ClientID)
}

func (s *Server) ClientIDs() []uint32 {
	clientIDs := make([]uint32, 0, len(s.clients))
	for k := range s.clients {
		clientIDs = append(clientIDs, k)
	}

	return clientIDs
}