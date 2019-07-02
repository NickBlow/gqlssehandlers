package streaming

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NickBlow/gqlssehandlers/clientid"
	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	"github.com/NickBlow/gqlssehandlers/protocol"
)

// Handler handles the endpoint for streaming and contains a reference to the SubscriptionBroker
type Handler struct {
	Broker *orchestration.Broker
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	closed := w.(http.CloseNotifier).CloseNotify()
	terminate := make(chan bool)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	messageChan := make(chan []byte)
	clientID := clientid.GetClientIDFromRequest(r)
	s.Broker.NewClients <- orchestration.ClientInfo{
		ClientID:             clientID,
		CommunicationChannel: messageChan,
		CloseChannel:         terminate,
	}

Loop:
	for {
		select {
		case <-closed:
			s.Broker.ClosedClients <- clientID
			break Loop
		case <-terminate:
			s.Broker.ClosedClients <- clientID
			break Loop
		case <-time.After(time.Second * 15):
			fmt.Fprintf(w, "data:%v \n\n", protocol.KeepAlivePayload)
			flusher.Flush()
		case data := <-messageChan:
			fmt.Fprintf(w, "data:%v \n\n", string(data))
			flusher.Flush()
		}
	}
	fmt.Println("stopped main thread")
}
