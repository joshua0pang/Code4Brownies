//
// Author: Vinhthuy Phan, 2015 - 2018
//
package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
)

//-----------------------------------------------------------------
var database *sql.DB
var InsertBroadcastSQL *sql.Stmt
var InsertUserSQL *sql.Stmt
var InsertSubmissionSQL *sql.Stmt
var InsertPollSQL *sql.Stmt
var UpdatePointsSQL *sql.Stmt
var UpdateCompletionTimeSQL *sql.Stmt
var InsertAttendanceSQL *sql.Stmt
var InsertQuizSQL *sql.Stmt
var InsertQuizAnswerSQL *sql.Stmt
var SelectBidFromSidSQL *sql.Stmt
var InsertFeedbackSQL *sql.Stmt

//-----------------------------------------------------------------
func init_db() {
	var err error
	prepare := func(s string) *sql.Stmt {
		stmt, err := database.Prepare(s)
		if err != nil {
			panic(err)
		}
		return stmt
	}

	database, err = sql.Open("sqlite3", USER_DB)
	if err != nil {
		panic(err)
	}

	create_tables()

	InsertBroadcastSQL = prepare("insert into broadcast (bid, content, language, date, hints, author) values (?, ?, ?, ?, ?, ?)")
	InsertUserSQL = prepare("insert into user (uid) values (?)")
	InsertSubmissionSQL = prepare("insert into submission (sid, uid, bid, points, description, language, date, content, hints_used, completion_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	InsertPollSQL = prepare("insert into poll (uid, is_correct, points, date) values (?, ?, ?, ?)")
	UpdatePointsSQL = prepare("update submission set points=? where sid=?")
	UpdateCompletionTimeSQL = prepare("update submission set completion_time=? where sid=?")
	InsertAttendanceSQL = prepare("insert into attendance (uid, date) values (?, ?)")
	InsertQuizSQL = prepare("insert into quiz (qid, question, answer, date) values (?, ?, ?, ?)")
	InsertQuizAnswerSQL = prepare("insert into quiz_answer (uid, qid, answer, point, date) values (?, ?, ?, ?, ?)")
	SelectBidFromSidSQL = prepare("select bid from submission where sid = ?")
	InsertFeedbackSQL = prepare("insert into feedback (author, uid, content, sid, date) values (?, ?, ?, ?, ?)")
}

//-----------------------------------------------------------------

func create_tables() {
	execSQL := func(s string) {
		sql_stmt, err := database.Prepare(s)
		if err != nil {
			panic(err)
		}
		sql_stmt.Exec()
	}
	execSQL("create table if not exists user (id integer primary key, uid text unique)")
	execSQL("create table if not exists broadcast (id integer primary key, bid text unique, content blob, language text, date timestamp, hints integer, author text)")
	// To change: bid should be integer (broadcast.id) instead of bid.
	execSQL("create table if not exists submission (id integer primary key, sid text unique, uid text, bid text, points integer, description text, language text, date timestamp, content blob, hints_used integer, completion_time timestamp)")
	execSQL("create table if not exists poll (id integer primary key, uid text, is_correct integer, points integer, date timestamp)")
	execSQL("create table if not exists attendance (id integer primary key, uid text, date timestamp)")
	execSQL("create table if not exists quiz (id integer primary key, qid text unique, question blob, answer text, date timestamp)")
	execSQL("create table if not exists quiz_answer (id integer primary key, uid text, qid text, answer text, point integer, date timestamp)")
	execSQL("create table if not exists feedback (id integer primary key, author text, uid text, content blob, sid text, date timestamp)")
}

//-----------------------------------------------------------------
func RegisterStudent(uid string) {
	BOARDS_SEM.Lock()
	defer BOARDS_SEM.Unlock()

	if _, ok := Boards[uid]; ok {
		fmt.Println(uid + " is already registered.")
		return
	}
	Boards[uid] = make([]*Board, 0)
	for i := 0; i < len(Boards["__default__"]); i++ {
		b := &Board{
			Content:      Boards["__default__"][i].Content,
			HelpContent:  Boards["__default__"][i].HelpContent,
			Description:  Boards["__default__"][i].Description,
			StartingTime: Boards["__default__"][i].StartingTime,
			Ext:          Boards["__default__"][i].Ext,
			Bid:          Boards["__default__"][i].Bid,
		}
		Boards[uid] = append(Boards[uid], b)
	}

	_, err := InsertUserSQL.Exec(uid)
	if err != nil {
		fmt.Println("Error inserting into user table.", err)
	} else {
		fmt.Println("New user", uid)
	}
}

//-----------------------------------------------------------------
func loadWhiteboards() {
	rows, _ := database.Query("select uid from user")
	defer rows.Close()
	var uid string

	BOARDS_SEM.Lock()
	defer BOARDS_SEM.Unlock()

	for rows.Next() {
		rows.Scan(&uid)
		Boards[uid] = make([]*Board, 0)
	}
	Boards["__default__"] = make([]*Board, 0)
}

//-----------------------------------------------------------------
func init_TA() {
	f, err := os.Open(TA_DB)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(bufio.NewReader(f))
	for {
		record, err2 := r.Read()
		if err2 == io.EOF {
			break
		}
		if err2 != nil {
			panic(err2)
		}
		TA_INFO[record[0]] = record[1]
	}
}
