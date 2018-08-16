package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"gopkg.in/russross/blackfriday.v2"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/content/{id}", noteHandler)
	router.HandleFunc("/delnote/{id}", delNoteHandler)
	router.HandleFunc("/addnote", addNoteHandler)
	router.HandleFunc("/", indexHandler)
	http.Handle("/", router)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

type Post struct {
	Title   string
	Listing map[string]string
}

type Note struct {
	Id    int
	Title string
	Body  template.HTML
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var noteList map[string]string
	noteList = make(map[string]string)
	var id string
	var title string
	var body string

	db, err := sql.Open("sqlite3", "./mdnotes_db.sqlite")
	checkErr(err)
	rows, err := db.Query("SELECT * From notes")
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&id, &title, &body)
		checkErr(err)
		noteList[id] = title
	}
	rows.Close()
	db.Close()

	p := Post{
		Title:   "Список заметок",
		Listing: noteList,
	}

	t, _ := template.ParseFiles("templates/base.html", "templates/home.html")
	t.Execute(w, p)
}

func noteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteid := vars["id"]
	var id int
	var title string
	var body string
	var str string

	db, err := sql.Open("sqlite3", "./mdnotes_db.sqlite")
	checkErr(err)
	defer db.Close()
	row := db.QueryRow("SELECT * From notes WHERE id= ?", noteid)

	err = row.Scan(&id, &title, &body)
	checkErr(err)

	str = string(blackfriday.Run([]byte(body)))

	// str := string(blackfriday.Run(bs))

	n := Note{
		Id:    id,
		Title: title,
		Body:  template.HTML(str),
	}
	t, _ := template.ParseFiles("templates/base.html", "templates/note.html")
	t.Execute(w, n)
}

func addNoteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		n := Note{
			Title: "Создать заметку",
		}
		t, err := template.ParseFiles("templates/base.html", "templates/addnote.html")
		checkErr(err)
		t.Execute(w, n)

	case "POST":
		err := r.ParseForm()
		checkErr(err)
		noteTitle := r.FormValue("noteTitle")
		noteBody := r.FormValue("noteBody")

		db, err := sql.Open("sqlite3", "./mdnotes_db.sqlite")
		checkErr(err)
		defer db.Close()
		result, err := db.Exec("insert into notes (title, body) values ($1, $2)", noteTitle, noteBody)
		checkErr(err)
		fmt.Println(result.LastInsertId)
		http.Redirect(w, r, "/", 301)
	}
}

func delNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteid := vars["id"]
	db, err := sql.Open("sqlite3", "./mdnotes_db.sqlite")
	checkErr(err)
	defer db.Close()
	res, err := db.Exec("delete from notes where id = ?", noteid)
	checkErr(err)
	fmt.Println(res.RowsAffected())
	http.Redirect(w, r, "/", 301)
}
