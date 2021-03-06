//
// Author: Vinhthuy Phan, 2015 - 2018
//
// Handlers for both instructor and TAs. Although the authorization is
// done differently (main.go), the operations are identical.
//
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//-----------------------------------------------------------------
func add_public_boardHandler(w http.ResponseWriter, r *http.Request, author string) {
	PublicBoard_SEM.Lock()
	defer PublicBoard_SEM.Unlock()
	content, ext := r.FormValue("content"), r.FormValue("ext")
	PublicBoard = append(PublicBoard, &Code{Content: content, Ext: ext})
	fmt.Fprintf(w, "Content added to public board")
}

//-----------------------------------------------------------------
func remove_public_boardHandler(w http.ResponseWriter, r *http.Request, author string) {
	PublicBoard_SEM.Lock()
	defer PublicBoard_SEM.Unlock()
	i, _ := strconv.Atoi(r.FormValue("i"))
	if i >= 0 && i < len(PublicBoard) {
		PublicBoard = append(PublicBoard[:i], PublicBoard[i+1:]...)
	}
	http.Redirect(w, r, "view_public_board?i=0", http.StatusSeeOther)
}

//-----------------------------------------------------------------
// Instructor/TAs give feedback and points to a student
//-----------------------------------------------------------------
func dequeueHandler(w http.ResponseWriter, r *http.Request, author string) {
	sid := r.FormValue("sid")
	success := DequeueSubmissionBySID(sid)
	if success {
		fmt.Fprintf(w, "Ok")
	} else {
		fmt.Fprintf(w, "Failed")
	}
}

//-----------------------------------------------------------------
// Instructor/TAs give feedback and points to a student
//-----------------------------------------------------------------
func feedbackHandler(w http.ResponseWriter, r *http.Request, author string) {
	BOARDS_SEM.Lock()
	defer BOARDS_SEM.Unlock()
	content, ext, sid := r.FormValue("content"), r.FormValue("ext"), r.FormValue("sid")
	points, _ := strconv.Atoi(r.FormValue("points"))
	has_feedback := r.FormValue("has_feedback")
	if sub, ok := AllSubs[sid]; ok {
		mesg := ""
		if has_feedback == "1" {
			_, err := InsertFeedbackSQL.Exec(author, sub.Uid, content, sid, time.Now())
			if err != nil {
				fmt.Println("Error inserting feedback.", err)
				fmt.Fprintf(w, "Error inserting feedback.")
			} else {
				bid := ""
				SelectBidFromSidSQL.QueryRow(sid).Scan(&bid)
				des := strings.SplitN(content, "\n", 2)[0]
				b := &Board{
					Content:      content,
					HelpContent:  "",
					Ext:          ext,
					Bid:          bid,
					Description:  des,
					StartingTime: time.Now(),
				}
				Boards[sub.Uid] = append(Boards[sub.Uid], b)
			}
			if author == "instructor" {
				mesg += fmt.Sprintf("Feedback sent.")
			} else {
				mesg += fmt.Sprintf("Feedback sent to %s.", sub.Uid)
			}
		}

		// Give points
		if points > 0 {
			success := RemoveSubmissionBySID(sid)
			// fmt.Println(success, author)
			if author == "instructor" || success == true {
				sub.Points = points
				_, err := UpdatePointsSQL.Exec(sub.Points, sid)
				if err != nil {
					mesg += "Failed to update points."
				} else {
					if author == "instructor" {
						mesg += fmt.Sprintf("\n%d points given.\n", sub.Points)
					} else {
						mesg += fmt.Sprintf("\n%d points given to %s.\n", sub.Points, sub.Uid)
					}
				}
			} else {
				// if instructor graded this submission, ignore TA.
				mesg += fmt.Sprintf("\nSubmission is already graded.")
			}
		} else {
			// reset status from "pending" to "new" so others can look at it
			sub.Status = "new"
			mesg += "Ungraded. Submission goes back into the queue."
		}
		fmt.Fprintf(w, mesg)
	} else {
		fmt.Fprintf(w, "sid %s is not found.", sid)
	}
}

//-----------------------------------------------------------------
// Instructor/TA retrieve all new submissions
//-----------------------------------------------------------------
func get_postsHandler(w http.ResponseWriter, r *http.Request, author string) {
	SUBS_SEM.Lock()
	defer SUBS_SEM.Unlock()
	how_many, _ := strconv.Atoi(r.FormValue("how_many"))
	selected := make([]*Submission, 0)
	get_all := false
	if how_many == -1 {
		how_many = len(NewSubs)
		get_all = true

	}
	j := 0
	for i := 0; j < how_many && i < len(NewSubs); i++ {
		if NewSubs[i].Status == "new" || get_all {
			NewSubs[i].Status = "pending"
			selected = append(selected, NewSubs[i])
			j++
		}
	}
	js, err := json.Marshal(selected)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

//-----------------------------------------------------------------
// peek at all submissions in NewSubs
//-----------------------------------------------------------------
func peekHandler(w http.ResponseWriter, r *http.Request, author string) {
	SUBS_SEM.Lock()
	defer SUBS_SEM.Unlock()
	js, err := json.Marshal(NewSubs)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

//-----------------------------------------------------------------
// retrieve a submission in NewSubs by index
//-----------------------------------------------------------------
func get_post_by_indexHandler(w http.ResponseWriter, r *http.Request, author string) {
	SUBS_SEM.Lock()
	defer SUBS_SEM.Unlock()
	idx, err := strconv.Atoi(r.FormValue("post"))
	if err != nil {
		fmt.Println(err.Error)
	} else {
		var js []byte
		if idx < 0 || idx >= len(NewSubs) {
			js, err = json.Marshal(&Submission{})
		} else {
			NewSubs[idx].Status = "pending"
			js, err = json.Marshal(NewSubs[idx])
		}
		if err != nil {
			fmt.Println(err.Error())
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	}
}

//-----------------------------------------------------------------
