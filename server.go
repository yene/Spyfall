package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)
import (
	"github.com/satori/go.uuid"
	"github.com/speps/go-hashids"
)

var currentGameID = 55 //P0vQ
var rooms []Room
var counter int = 55

// fmt.Fprintf(w, "GET, %q", html.EscapeString(r.URL.Path))
// http.Error(w, "Invalid request method.", 405)

type Page struct {
	Code  string
	Admin bool
}

type Room struct {
	ID      int
	Players []Player
}

type Player struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Admin bool   `json:"-"`
	UUID  string `json:"-"`
}

func main() {
	http.HandleFunc("/room/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		hd := hashids.NewData()
		hd.Salt = "super secret salt"
		hd.MinLength = 4
		h := hashids.NewWithData(hd)
		c, _ := h.Encode([]int{counter})

		// create new room
		uuidS := uuid.NewV4().String()
		admin := Player{"Red", "red", true, uuidS}
		players := []Player{admin}
		rooms = append(rooms, Room{counter, players})

		// give user a cookie with id
		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: "spyfall", Value: uuidS, Expires: expiration}
		http.SetCookie(w, &cookie)

		//
		counter++

		t, _ := template.ParseFiles("static/room.html")
		p := &Page{Code: c, Admin: true}
		t.Execute(w, p)
		fmt.Println("created new room ", c)
	})

	http.HandleFunc("/room/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r) // w.WriteHeader(http.StatusMethodNotAllowed) return
			return
		}
		r.ParseForm()
		lobbyCode := r.Form["lobbyCode"][0]
		hd := hashids.NewData()
		hd.Salt = "super secret salt"
		h := hashids.NewWithData(hd)
		d := h.Decode(lobbyCode)
		if d[0] == currentGameID {
			fmt.Println(d)
			t, _ := template.ParseFiles("static/room.html")
			p := &Page{Code: lobbyCode, Admin: false}
			t.Execute(w, p)
		} else {
			http.NotFound(w, r)
			return
		}
	})

	type Profile struct {
		Name    string
		Hobbies []string
	}

	http.HandleFunc("/players.json", func(w http.ResponseWriter, r *http.Request) {
		lobbyCode := r.URL.Query().Get("code")
		if len(lobbyCode) != 4 { // Error: no code provided
			http.NotFound(w, r)
			return
		}
		hd := hashids.NewData()
		hd.Salt = "super secret salt"
		h := hashids.NewWithData(hd)
		d := h.Decode(lobbyCode)
		room := roomWithID(d[0]) // TODO error handling when room not found

		js, err := json.Marshal(room.Players)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":3000", nil)
}

func roomWithID(id int) Room {
	for _, r := range rooms {
		if r.ID == id {
			return r
		}
	}
	return rooms[0]
}
