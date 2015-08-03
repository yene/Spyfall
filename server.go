package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)
import (
	"github.com/satori/go.uuid"
	"github.com/speps/go-hashids"
)

var rooms = make(map[int]*Room)
var counter int = 0
var cards []Card
var HashID *hashids.HashID

type Card struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type PageLobby struct {
	ID    int
	Code  string
	Admin bool
}

type PageGame struct {
	Spy      bool
	Location int
	Cards    []Card
}

type Room struct {
	ID       int
	Players  []Player `json:"players"`
	Started  bool     `json:"started"`
	Location int
}

type Player struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Admin bool   `json:"-"`
	UUID  string `json:"-"`
	Spy   bool
}

func (r Room) playerForUUID(uuid string) Player {
	for _, value := range r.Players {
		if value.UUID == uuid {
			return value
		}
	}
	return Player{"Grey", "grey", false, "", false} // TODO: return error
}

func (r *Room) setup() {
	r.Location = rand.Intn(len(cards))
	spy := rand.Intn(len(r.Players))
	r.Players[spy].Spy = true
	r.Started = true

}

func main() {
	cardsJSON, _ := ioutil.ReadFile("static/cards/cards.json")
	json.Unmarshal([]byte(cardsJSON), &cards)

	hd := hashids.NewData()
	hd.Salt = "super secret salt"
	hd.MinLength = 4
	HashID = hashids.NewWithData(hd)

	http.HandleFunc("/room/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		c, _ := HashID.Encode([]int{counter})

		// create new room
		uuidS := uuid.NewV4().String()
		admin := Player{"Red", "red", true, uuidS, false}
		players := []Player{admin}
		rooms[counter] = &Room{ID: counter, Players: players, Started: false}

		// give user a cookie with UUID
		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "spyfall", Value: uuidS, Path: "/", Expires: expiration}
		http.SetCookie(w, &cookie)

		t, _ := template.ParseFiles("static/room.html")
		p := &PageLobby{Code: c, ID: counter, Admin: true}
		t.Execute(w, p)
		fmt.Println("created new room", c, counter)

		counter++
	})

	http.HandleFunc("/room/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		r.ParseForm()
		lobbyCode := r.Form["lobbyCode"][0]
		d := HashID.Decode(lobbyCode)
		roomID := d[0]
		if room, ok := rooms[roomID]; ok {
			if room.Started {
				http.Redirect(w, r, "/", 303)
				return
			}

			t, _ := template.ParseFiles("static/room.html")
			p := &PageLobby{Code: lobbyCode, ID: d[0], Admin: false}
			t.Execute(w, p)
		} else {
			http.Redirect(w, r, "/", 303)
			return
		}
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		roomIDStr := r.URL.Query().Get("id")
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if room, ok := rooms[roomID]; ok {
			if room.Started == false {
				room.setup()
				fmt.Println("Started room", roomID)
			}
			t, _ := template.ParseFiles("static/game.html")
			cookie := r.Cookies()[0]
			uuid := cookie.Value
			player := room.playerForUUID(uuid)

			p := &PageGame{Spy: player.Spy, Location: room.Location, Cards: cards}
			t.Execute(w, p)
		} else {
			http.NotFound(w, r)
			return
		}
	})

	http.HandleFunc("/players.json", func(w http.ResponseWriter, r *http.Request) {
		roomIDStr := r.URL.Query().Get("id")
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if room, ok := rooms[roomID]; ok {
			js, err := json.Marshal(room)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		} else {
			http.NotFound(w, r)
			return
		}
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":3000", nil)
}
