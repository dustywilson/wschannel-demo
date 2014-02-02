package main

import (
	"bitbucket.org/dustywilson/wschannel"
	"encoding/json"
	"fmt"
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

type GenericMessage struct {
	Title   string      `json:"title,omitempty"`
	Message interface{} `json:"message,omitempty"`
}

func run() {
	t := time.Tick(time.Second * 5)
	for {
		for _, sessionId := range ws.GetSessions() {
			ss := ws.GetSession(sessionId)

			// send message via the session's send method:
			ss.Send(GenericMessage{"time", time.Now()})

			// send message via the session's channel:
			ss.C <- GenericMessage{"sessionsGreetings", fmt.Sprintf("Hello session [%s].", ss.Id())}

			for _, connectionId := range ss.GetConnections() {
				cn := ss.GetConnection(connectionId)

				// send a message directly to a connection instead of a whole session:
				cn.C <- GenericMessage{"helloConnection", fmt.Sprintf("Hello connection [%s].", cn.Id())}
			}
		}
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
