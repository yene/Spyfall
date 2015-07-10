package main

import (
	"fmt"
	"net/http"
)
import "github.com/speps/go-hashids"

var currentGameID = 55

// fmt.Fprintf(w, "GET, %q", html.EscapeString(r.URL.Path))
// http.Error(w, "Invalid request method.", 405)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		hd := hashids.NewData()
		hd.Salt = "super secret salt"
		hd.MinLength = 4
		h := hashids.NewWithData(hd)
		e, _ := h.Encode([]int{currentGameID})
		fmt.Println(e)

	})

	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
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
		} else {
			http.NotFound(w, r)
			return
		}

	})

	http.ListenAndServe(":3000", nil)
}
