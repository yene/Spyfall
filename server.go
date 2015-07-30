package main

import (
	"fmt"
	"html/template"
	"net/http"
)
import "github.com/speps/go-hashids"

var currentGameID = 55 //P0vQ

// fmt.Fprintf(w, "GET, %q", html.EscapeString(r.URL.Path))
// http.Error(w, "Invalid request method.", 405)

type Page struct {
	Code  string
	Admin bool
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/room/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		hd := hashids.NewData()
		hd.Salt = "super secret salt"
		hd.MinLength = 4
		h := hashids.NewWithData(hd)
		c, _ := h.Encode([]int{currentGameID})
		t, _ := template.ParseFiles("static/room.html")
		p := &Page{Code: c, Admin: true}
		t.Execute(w, p)
	})

	http.HandleFunc("/room/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r) // w.WriteHeader(http.StatusMethodNotAllowed) return
			return
		}
		r.ParseForm()
		lobbyID := r.Form["lobbyID"][0]
		hd := hashids.NewData()
		hd.Salt = "super secret salt"
		h := hashids.NewWithData(hd)
		d := h.Decode(lobbyID)
		if d[0] == currentGameID {
			fmt.Println(d)
			t, _ := template.ParseFiles("static/room.html")
			p := &Page{Code: lobbyID, Admin: false}
			t.Execute(w, p)
		} else {
			http.NotFound(w, r)
			return
		}
	})

	http.ListenAndServe(":3000", nil)
}
