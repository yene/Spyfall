package main

import (
	"math/rand"
	"net/http"
	"time"
)

// Room is the root struct of our game, it also is sent to the user as json
// Don't forget to hide unused json values
type Room struct {
	ID        int
	Players   []player `json:"players"`
	Started   bool     `json:"started"`
	Location  int      `json:"-"`
	Countdown int      `json:"-"`
}

type player struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Admin bool   `json:"-"`
	UUID  string `json:"-"`
	Spy   bool   `json:"-"`
}

func (r Room) playerForUUID(uuid string) (p player, found bool) {
	for _, value := range r.Players {
		if value.UUID == uuid {
			return value, true
		}
	}
	return player{}, false
}

func (r Room) playerForCookies(cookies []*http.Cookie) (p player, found bool) {
	if len(cookies) == 0 {
		return player{}, false
	}

	return r.playerForUUID(cookies[0].Value) // TODO: dont take first cookie, search for spyfall cookie
}

func (r *Room) setup() {
	r.Location = rand.Intn(len(cards))
	spy := rand.Intn(len(r.Players))
	r.Players[spy].Spy = true

	t := int(time.Now().Unix())
	t = t + (60 * 8)
	r.Countdown = t
	r.Started = true
}
