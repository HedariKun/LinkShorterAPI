package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type newThing struct {
	string
	int
}

type user struct {
	Username, Password, Email string
	Token                     string
}

type tokenData struct {
	Statu bool
	Token string
}

type shortURLData struct {
	OriginalLink string
	URLHash      string
	DeleteHash   string
	CreateDate   int64
	Statu        bool
}

type message struct {
	Statu   bool
	Message string
}

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		db, err := sql.Open("mysql", Config.DBUser+":"+Config.DBPass+"@/"+Config.DBName)
		checkErr(err)

		Username := r.PostFormValue("Username")
		Password := r.PostFormValue("Password")

		if Username == "" || Password == "" {
			msg := message{false, "You need to Provide a Username and Password"}
			j, _ := json.Marshal(msg)
			fmt.Fprintf(w, string(j))
		} else {
			quer, _ := db.Query("SELECT `ID` FROM `user` WHERE `Username`=\"" + Username + "\"")
			var rows int

			for quer.Next() {
				var ID int
				quer.Scan(&ID)
				if ID > 0 {
					rows++
				}
			}
			quer.Close()

			if rows > 0 {
				msg := message{false, "Username Already Used"}
				j, _ := json.Marshal(msg)
				fmt.Fprintf(w, string(j))
			} else {
				Token := String(25)
				ins, _ := db.Prepare("INSERT INTO `user`(`ID`, `Username`, `Password`, `Token`) VALUES (null, ?, ?, ?)")
				ins.Exec(Username, Password, Token)
				msg := message{true, "User was Craeted Successfully"}
				j, _ := json.Marshal(msg)
				fmt.Fprintf(w, string(j))
			}
		}
		db.Close()
	}
}

func getToken(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Println(err)
			return
		}
		db, err := sql.Open("mysql", "root:@/linkshorter")
		checkErr(err)

		Username := r.PostFormValue("Username")
		Password := r.PostFormValue("Password")
		rows, _ := db.Query("SELECT * FROM user")
		defer db.Close()
		for rows.Next() {
			var id int
			var User string
			var Pass string
			var Token string
			rows.Scan(&id, &User, &Pass, &Token)
			if Username == User && Password == Pass {
				TokenD := tokenData{true, Token}
				j, _ := json.Marshal(TokenD)

				fmt.Fprintf(w, string(j))
				return
			}
		}
		msg := message{false, "Incorrect Username or Password"}
		j, _ := json.Marshal(msg)
		fmt.Fprintf(w, string(j))
	} else {
		msg := message{false, "You need to call it from a Post Request"}
		j, _ := json.Marshal(msg)
		fmt.Fprintf(w, string(j))
	}
}

func shortURL(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	url := r.FormValue("url")
	Token := r.FormValue("token")
	db, _ := sql.Open("mysql", "root:@/linkshorter")

	rows, _ := db.Query("SELECT Token FROM user")

	defer db.Close()
	for rows.Next() {
		var DataToken string
		rows.Scan(&DataToken)
		if DataToken == Token {
			if isValidURL(url) {
				DeleteHash := String(10)
				URLHash := String(5)
				shortData := shortURLData{url, URLHash, DeleteHash, time.Now().Unix(), true}

				insertData, err := db.Prepare("INSERT INTO `links`(`ID`, `OriginalLink`, `UsedToken`, `UrlHash`, `DeleteHash`, `CreateDate`) VALUES (null, ?, ?, ?, ?, ?)")
				checkErr(err)

				insertData.Exec(shortData.OriginalLink, Token, shortData.URLHash, shortData.DeleteHash, shortData.CreateDate)

				j, _ := json.Marshal(shortData)
				fmt.Fprintf(w, string(j))

				return
			}
			msg := message{false, "You need to provide a vaild URL"}
			j, _ := json.Marshal(msg)
			fmt.Fprintf(w, string(j))
			return
		}
	}

	msg := message{false, "Invaild Token Provided"}
	j, _ := json.Marshal(msg)
	fmt.Fprintf(w, string(j))

}

func redirectURL(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	Parameter := r.URL.String()[1:]
	if len(Parameter) > 0 {
		db, _ := sql.Open("mysql", "root:@/linkshorter")

		rows, _ := db.Query("SELECT `ID`, `UrlHash`, `DeleteHash`, `OriginalLink` FROM `links`")

		for rows.Next() {
			var ID int
			var Hash string
			var OriginalLink string
			var DeleteHash string
			rows.Scan(&ID, &Hash, &DeleteHash, &OriginalLink)
			if Hash == Parameter {
				http.Redirect(w, r, OriginalLink, 301)
				db.Close()
				return
			}

			if DeleteHash == Parameter {
				req, _ := db.Prepare("DELETE FROM `links` WHERE `ID`=?")
				req.Exec(ID)
				http.Redirect(w, r, "/", 301)
				db.Close()
				return
			}

		}
		db.Close()
	}
}
