package streaming

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NickBlow/gqlssehandlers/internal/orchestration"
	gonanoid "github.com/matoous/go-nanoid"
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
	clientID, err := gonanoid.Nanoid()
	if err != nil {
		fmt.Println("Couldn't generate client ID")
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	messageChan := make(chan string)

	s.Broker.NewClients <- orchestration.ClientInfo{
		ClientID:             clientID,
		CommunicationChannel: messageChan,
	}

Loop:
	for {
		select {
		case <-closed:
			s.Broker.ClosedClients <- clientID
			break Loop
		case <-time.After(time.Second * 3):
			fmt.Fprint(w, ":KEEPALIVE \n")
			flusher.Flush()
		case val := <-messageChan:
			fmt.Fprintf(w, "data:%v \n", val)
			flusher.Flush()
		}
	}
	fmt.Println("stopped main thread")
}