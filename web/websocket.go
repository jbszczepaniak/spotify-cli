package web

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/zmb3/spotify"
	"log"
	"net/http"
	"time"
)

type WebsocketHandler struct {
	// signal sent by the application to the websocket handler that triggers tear down
	// of websocket.
	PlayerShutdown chan bool

	// as soon as web player is ready it sends player device id as the very first
	// message on the websocket connection. Application may want to use this ID in
	// order to control on which player to play music (the one that it created, or
	// maybe on the other device that is also working)
	PlayerDeviceID chan spotify.ID

	// each time web player changes it's State it sens information about what is
	// currently played on the websocket
	PlayerStateChange chan *WebPlaybackState
}

type WebPlaybackReadyDevice struct {
	DeviceId string
}

type WebPlaybackState struct {
	CurrentTrackName  string
	CurrentAlbumName  string
	CurrentArtistName string
}

func (s *WebsocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var v WebPlaybackReadyDevice
	_, message, err := conn.ReadMessage()
	err = json.Unmarshal(message, &v)
	if err != nil {
		log.Fatalf("could not unmarshall message from web player, err: %v", err)
	}
	deviceId := spotify.ID(v.DeviceId)
	log.Printf("web player is ready, it's ID is %s", deviceId.String())
	s.PlayerDeviceID <- deviceId

	go func() {
		for range time.Tick(500 * time.Millisecond) {
			var state WebPlaybackState
			_, message, err = conn.ReadMessage()
			err = json.Unmarshal(message, &state)
			if err != nil {
				log.Printf("could not Unmarshall message: %s, err: %s", message, err)
			}
			s.PlayerStateChange <- &state
		}
	}()

	<-s.PlayerShutdown
	err = conn.WriteJSON("{\"close\": true}")
	if err != nil {
		log.Printf("could not close connection, err %v", err)
	}

}