//
// Author: Vinhthuy Phan, 2015 - 2018
//
package main

import (
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

//-----------------------------------------------------------------
func informIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err.Error() + "\n")
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			return ipnet.IP.String() + ":" + PORT
		}
	}
	return ""
}

//-----------------------------------------------------------------
// Make sure teacher runs server on his own laptop.
//-----------------------------------------------------------------
func Authorize(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "localhost:4030" {
			fn(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized")
		}
	}
}

//-----------------------------------------------------------------
// Authorize TAs
//-----------------------------------------------------------------
func AuthorizeTA(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		passcode, ok := TA_INFO[r.FormValue("name")]
		if ok && passcode == r.FormValue("passcode") {
			fn(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized")
			fmt.Println("Unauthorized TA access:", r.FormValue("name"), r.FormValue("passcode"))
		}
	}
}

//-----------------------------------------------------------------
// Register automatically if a student is not yet registered.
//-----------------------------------------------------------------
func AutoRegister(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := Boards[r.FormValue("uid")]; !ok {
			RegisterStudent(r.FormValue("uid"))
		}
		fn(w, r)
	}
}

//-----------------------------------------------------------------
func main() {
	SERVER = informIPAddress()
	fmt.Println("*********************************************")
	fmt.Printf("*   Code4Brownies (%s)\n", VERSION)
	fmt.Printf("*   Server address: %s\n", SERVER)
	fmt.Println("*********************************************\n")
	rand.Seed(time.Now().UnixNano())
	USER_DB = filepath.Join(".", "c4b.db")
	TA_DB = filepath.Join(".", "c4b_tas.txt")
	flag.StringVar(&USER_DB, "db", USER_DB, "user database (sqlite).")
	flag.StringVar(&TA_DB, "ta", TA_DB, "TA information.")
	flag.Parse()

	// TA handlers
	http.HandleFunc("/ta_feedback", AuthorizeTA(feedbackHandler))
	http.HandleFunc("/ta_give_points", AuthorizeTA(ta_give_pointsHandler))
	http.HandleFunc("/ta_get_posts", AuthorizeTA(get_postsHandler))

	// student handlers
	http.HandleFunc("/share", AutoRegister(shareHandler))
	http.HandleFunc("/my_points", AutoRegister(my_pointsHandler))
	http.HandleFunc("/receive_broadcast", AutoRegister(receive_broadcastHandler))
	http.HandleFunc("/checkin", AutoRegister(checkinHandler))

	// teacher handlers
	http.HandleFunc("/test", Authorize(testHandler))
	http.HandleFunc("/clear_whiteboards", Authorize(clear_whiteboardsHandler))
	http.HandleFunc("/clear_questions", Authorize(clear_questionsHandler))
	http.HandleFunc("/start_poll", Authorize(start_pollHandler))
	http.HandleFunc("/query_poll", Authorize(query_pollHandler))
	http.HandleFunc("/view_poll", Authorize(view_pollHandler))
	http.HandleFunc("/answer_poll", Authorize(answer_pollHandler))
	http.HandleFunc("/give_points", Authorize(give_pointsHandler))
	http.HandleFunc("/peek", Authorize(peekHandler))
	http.HandleFunc("/broadcast", Authorize(broadcastHandler))
	http.HandleFunc("/get_post", Authorize(get_postHandler))
	http.HandleFunc("/get_posts", Authorize(get_postsHandler))
	http.HandleFunc("/send_quiz_question", Authorize(send_quiz_questionHandler))

	// public handlers
	http.HandleFunc("/track_board", track_boardHandler)
	http.HandleFunc("/track_submissions", track_submissionsHandler)
	http.HandleFunc("/view_questions", view_questionsHandler)
	http.HandleFunc("/get_questions", get_questionsHandler)

	init_db()
	init_TA()
	loadWhiteboards()

	// Start serving app
	err := http.ListenAndServe("0.0.0.0:"+PORT, nil)
	if err != nil {
		panic(err.Error() + "\n")
	}
}
