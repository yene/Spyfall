package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)
import (
	"github.com/satori/go.uuid"
	"github.com/speps/go-hashids"
)

var rooms = make(map[int]*Room)
var roomsMU sync.Mutex
var counter int = 0

var cards []Card
var HashID *hashids.HashID
var colors = []string{"Magenta", "LightSlateGray", "PaleVioletRed", "Peru", "RebeccaPurple", "LightSeaGreen", "Tomato", "SeaGreen", "Maroon", "GoldenRod", "DarkSlateBlue"}

type Card struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Room struct {
	ID        int
	Players   []Player `json:"players"`
	Started   bool     `json:"started"`
	Location  int      `json:"-"`
	Countdown int      `json:"-"`
}

type Player struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Admin bool   `json:"-"`
	UUID  string `json:"-"`
	Spy   bool   `json:"-"`
}

func (r Room) playerForUUID(uuid string) (p Player, found bool) {
	for _, value := range r.Players {
		if value.UUID == uuid {
			return value, true
		}
	}
	return Player{}, false
}

func (r Room) playerForCookies(cookies []*http.Cookie) (p Player, found bool) {
	if len(cookies) == 0 {
		return Player{}, false
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

func main() {
	var (
		httpAddr = flag.String("http", ":3000", "HTTP service address")
		salt     = flag.String("salt", "super secret salt", "Salt is the secret used to make the generated id harder to guess")
	)
	flag.Parse()

	cardsJSON, _ := ioutil.ReadFile("static/cards/cards.json")
	json.Unmarshal([]byte(cardsJSON), &cards)

	hd := hashids.NewData()
	hd.Salt = *salt
	hd.MinLength = 5
	hd.Alphabet = "abcdefghijklmnopqrstuvwxyz"
	HashID = hashids.NewWithData(hd)

	http.HandleFunc("/room/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		roomsMU.Lock()
		defer roomsMU.Unlock()

		c, _ := HashID.Encode([]int{counter})

		uuidS := uuid.NewV4().String()
		admin := Player{Name: colors[0], Color: colors[0], Admin: true, UUID: uuidS, Spy: false}
		players := []Player{admin}
		rooms[counter] = &Room{ID: counter, Players: players, Started: false}

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "spyfall", Value: uuidS, Path: "/", Expires: expiration}
		http.SetCookie(w, &cookie)
		fmt.Println("created new room", c, counter)

		counter++
		http.Redirect(w, r, "/room?code="+c, 303)
	})

	http.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		roomsMU.Lock()
		defer roomsMU.Unlock()

		c := r.URL.Query().Get("code")

		d := HashID.Decode(c)
		roomID := d[0]
		if room, ok := rooms[roomID]; ok {
			if room.Started {
				http.Redirect(w, r, "/", 303) // TODO: show error message that room already started
				return
			}
			var player Player

			if p, ok := room.playerForCookies(r.Cookies()); ok {
				player = p
			} else {
				uuidS := uuid.NewV4().String()
				color := colors[len(room.Players)]
				player = Player{Name: color, Color: color, Admin: false, UUID: uuidS, Spy: false}

				expiration := time.Now().Add(365 * 24 * time.Hour)
				cookie := http.Cookie{Name: "spyfall", Value: uuidS, Path: "/", Expires: expiration}
				http.SetCookie(w, &cookie)

				room.Players = append(room.Players, player)
			}

			t, _ := template.ParseFiles("static/room.html")
			data := struct {
				ID     int
				Code   string
				Player Player
			}{
				ID: roomID, Code: c, Player: player,
			}
			t.Execute(w, data)
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
		roomsMU.Lock()
		defer roomsMU.Unlock()
		if room, ok := rooms[roomID]; ok {
			if room.Started == false {
				room.setup()
				fmt.Println("Started room", roomID)
			}
			t, _ := template.ParseFiles("static/game.html")
			cookie := r.Cookies()[0]
			uuid := cookie.Value
			player, ok := room.playerForUUID(uuid)
			if !ok {
				http.NotFound(w, r)
				return
			}

			data := struct {
				Spy       bool
				Location  int
				Cards     []Card
				Countdown int
			}{
				Spy: player.Spy, Location: room.Location, Cards: cards, Countdown: room.Countdown,
			}

			t.Execute(w, data)
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
		roomsMU.Lock()
		defer roomsMU.Unlock()
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
	fmt.Println("listen on", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
