package main

import (
	"context"
	"fmt"
	"log"

	// "math"
	"net/http"
	"personal-project/connection"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type Project struct {
	Title           string
	StartDate       time.Time
	EndDate         time.Time
	Description     string
	Duration        string
	Golang          string
	JavaScript      string
	React           string
	Node            string
	Id              int
	Formatstartdate string
	Formatenddate   string
	IsLogin         bool
}

type SessionData struct {
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = SessionData{}

type user struct {
	ID       int
	Name     string
	Email    string
	Password string
}

var DetailProject = Project{}

var projectData = []Project{}

//===============================================

func main() {

	route := mux.NewRouter()

	connection.DatabaseConnect()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/hi", helloworld).Methods("GET")

	route.HandleFunc("/home", home).Methods("GET")

	route.HandleFunc("/project", project).Methods("GET")

	route.HandleFunc("/submit", submit).Methods("POST")

	route.HandleFunc("/contact", contact).Methods("GET")

	route.HandleFunc("/delete/{index}", delete).Methods("GET")

	route.HandleFunc("/edit/{index}", edit).Methods("GET")

	route.HandleFunc("/editButton/{id}", editButton).Methods("POST")

	route.HandleFunc("/detail/{id}", detail).Methods("GET")

	route.HandleFunc("/form-register", formRegister).Methods("GET")

	route.HandleFunc("/register", register).Methods("POST")

	route.HandleFunc("/form-login", formLogin).Methods("GET")

	route.HandleFunc("/login", login).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("server is Running")
	http.ListenAndServe("localhost:8000", route)
}

//===============================================

func helloworld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

//===============================================

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf8")

	var tmplt, err = template.ParseFiles("pages/home.html")

	if err != nil {
		w.Write([]byte("file doesn't exist: " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)

	}

	data, _ := connection.Conn.Query(context.Background(), "SELECT id, title, description, duration FROM db_project")

	var result []Project
	for data.Next() {
		var each = Project{}

		var err = data.Scan(&each.Id, &each.Title, &each.Description, &each.Duration)
		if err != nil {
			fmt.Println(err.Error)
			return
		}

		result = append(result, each)
	}

	resData := map[string]interface{}{
		"DataSession": Data,
		"Project":     result,
	}

	tmplt.Execute(w, resData)
}

//===============================================

func project(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf8")

	var tmplt, err = template.ParseFiles("pages/project.html")

	if err != nil {
		w.Write([]byte("file doesn't exist: " + err.Error()))
		return
	}

	tmplt.Execute(w, "")
}

//===============================================

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf8")

	var tmplt, err = template.ParseFiles("pages/contact.html")

	if err != nil {
		w.Write([]byte("file doesn't exist: " + err.Error()))
		return
	}

	tmplt.Execute(w, "")
}

//===============================================

func detail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf8")

	var tmplt, err = template.ParseFiles("pages/detail.html")

	if err != nil {
		w.Write([]byte("file doesn't exist: " + err.Error()))
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, title, start_date, end_date, description, duration FROM db_project WHERE id=$1", id).Scan(
		&DetailProject.Id, &DetailProject.Title, &DetailProject.StartDate, &DetailProject.EndDate, &DetailProject.Description, &DetailProject.Duration)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	DetailProject.Formatstartdate = DetailProject.StartDate.Format("2 january 2006")
	DetailProject.Formatenddate = DetailProject.EndDate.Format("2 january 2006")

	data := map[string]interface{}{
		"DetailProject": DetailProject,
	}

	tmplt.Execute(w, data)
}

//===============================================

func submit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	title := r.PostForm.Get("addTitle")
	startDate := r.PostForm.Get("addStartDate")
	endDate := r.PostForm.Get("addEndDate")
	description := r.PostForm.Get("addDescription")

	layout := "2006-01-02"
	parsingstartdate, _ := time.Parse(layout, startDate)
	parsingenddate, _ := time.Parse(layout, endDate)

	hours := parsingenddate.Sub(parsingstartdate).Hours()
	days := hours / 24

	var duration string

	if days > 0 {
		duration = strconv.FormatFloat(days, 'f', 0, 64) + " days"
	}

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO db_project(title, start_date, end_date, description, duration) VALUES($1, $2, $3, $4, $5)", title, parsingstartdate, parsingenddate, description, duration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/home", http.StatusMovedPermanently)

}

//===============================================

func delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["index"])
	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM db_project WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}
	http.Redirect(w, r, "/home", http.StatusFound)
}

//===============================================

func edit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf8")

	var tmplt, err = template.ParseFiles("pages/project-edit.html")

	if err != nil {
		w.Write([]byte("file doesn't exist: " + err.Error()))
		return
	}

	var DetailProject = Project{}

	index, _ := strconv.Atoi(mux.Vars(r)["index"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, title, start_date, end_date, description, duration FROM db_project WHERE id=$1", index).Scan(
		&DetailProject.Id, &DetailProject.Title, &DetailProject.StartDate, &DetailProject.EndDate, &DetailProject.Description, &DetailProject.Duration)

	data := map[string]interface{}{
		"ProjectEdit": DetailProject,
	}

	tmplt.Execute(w, data)
}

//===============================================

func editButton(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	http.Redirect(w, r, "/home", http.StatusMovedPermanently)

	title := r.PostForm.Get("addTitle")
	startDate := r.PostForm.Get("addStartDate")
	endDate := r.PostForm.Get("addEndDate")
	description := r.PostForm.Get("addDescription")

	golang := r.PostForm.Get("addGolang")
	javaScript := r.PostForm.Get("addJavaScript")
	react := r.PostForm.Get("addReact")
	nodejs := r.PostForm.Get("addNode")

	layout := "2006-01-02"
	parsingstartdate, _ := time.Parse(layout, startDate)
	parsingenddate, _ := time.Parse(layout, endDate)

	hours := parsingenddate.Sub(parsingstartdate).Hours()
	days := hours / 24

	var duration string

	if days > 0 {
		duration = strconv.FormatFloat(days, 'f', 0, 64) + " days"
	}

	newProject := Project{
		Title:       title,
		Duration:    duration,
		Description: description,
		Golang:      golang,
		JavaScript:  javaScript,
		React:       react,
		Node:        nodejs,
		Id:          id,
	}
	projectData[id] = newProject

}

//===============================================

func formRegister(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("pages/form-register.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)

}

//===============================================

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("inputName")
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	// fmt.Println(passwordHash)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO db_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}

//===============================================

func formLogin(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("pages/form-login.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)

}

//===============================================

func login(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	user := user{}

	// mengambil data email, dan melakukan pengecekan email
	err = connection.Conn.QueryRow(context.Background(),
		"SELECT * FROM db_user WHERE email=$1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	// mengambil data email, dan melakukan pengecekan password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	//================SESSION=====================
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["IsLogin"] = true
	session.Options.MaxAge = 1800
	session.AddFlash("succesfull login", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/form-login", http.StatusSeeOther)
}
