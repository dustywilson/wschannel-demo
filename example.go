package main

import (
	"bitbucket.org/dustywilson/wschannel"
	"encoding/json"
	"github.com/dchest/uniuri"
	"net/http"
	"time"
)

var ws = wschannel.NewService()

func main() {
	http.Handle("/ws/", ws.Handler("/ws"))
	http.HandleFunc("/api/session", getSession)
	http.HandleFunc("/api/ping", ping)
	http.Handle("/", http.FileServer(http.Dir("www")))
	go run()
	panic(http.ListenAndServe(":5555", nil))
}

func run() {
	fred, _ := ws.NewSession("fredflintstone")
	t := time.Tick(time.Millisecond * 500)
	i := 0
	for {
		i++
		fred.C <- time.Now()
		<-t
	}
}

type SessionMessage struct {
	SessionId string `json:"sessionId,omitempty"`
}

func getSession(w http.ResponseWriter, r *http.Request) {
	sessionCookie, _ := r.Cookie("session")
	if sessionCookie == nil {
		sessionCookie = new(http.Cookie)
		sessionCookie.Name = "session"
		sessionCookie.Value = uniuri.NewLen(32)
	}
	sessionCookie.HttpOnly = true
	sessionCookie.Expires = time.Now().AddDate(1, 0, 0)
	sessionId := sessionCookie.Value // FIXME: shouldn't trust the user's cookie so blindly...  consider encrypted and/or signed cookie.
	ws.NewSession(sessionId)
	http.SetCookie(w, sessionCookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SessionMessage{sessionId})
}

type PingMessage struct {
	Message string `json:"message,omitempty"`
	Random  string `json:"random,omitempty"`
}

func ping(w http.ResponseWriter, r *http.Request) {
	ss := ws.GetSession(r.FormValue("sessionId"))
	if ss == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	ss.Send(PingMessage{
		Message: r.FormValue("message"),
		Random:  r.FormValue("random"),
	})
}
