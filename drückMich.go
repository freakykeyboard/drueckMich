package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"net/http"
	"path/filepath"
)

var Id string
var cookieName string
var t = template.Must(template.ParseFiles(filepath.Join("./", "tpl", "head.html"),
	filepath.Join("./", "tpl", "login.html"), filepath.Join("./", "tpl", "bookmarks.html"),
	filepath.Join("./", "tpl", "end.html"), filepath.Join("./", "tpl", "registrate.html")))

type userTy struct {
	Username   string `bson:"username"`
	Password   string `bson:"password"`
	IsLoggedIn bool   `bson:"is_logged_in"`
}
type readUserTy struct {
	ID         bson.ObjectId `bson:"_id"`
	Username   string        `bson:"username"`
	Password   string        `bson:"password"`
	IsLoggedIn bool          `bson:"is_logged_in"`
}
type usersTY struct {
	Users []userTy
}
type regDataTy struct {
	ErrorMessage string
}

var usersCollection *mgo.Collection

func main() {
	cookieName = "pressMe"
	dbSession, _ := mgo.Dial("localhost")
	defer dbSession.Close()
	db := dbSession.DB("drückMich")
	usersCollection = db.C("users")
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/drückMich", pressMeHandler)
	http.HandleFunc("/Url", urlAjaxHandler)
	http.HandleFunc("/registrate", registrateHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/deleteAcoount", deleteAcountHandler)
	http.ListenAndServe(":4242", nil)
}
func deleteAcountHandler(writer http.ResponseWriter, request *http.Request) {

}
func logoutHandler(writer http.ResponseWriter, request *http.Request) {
	oldCookie, _ := request.Cookie("pressMe")
	fmt.Println(oldCookie.Value)
	docSelector := bson.M{"_id": oldCookie.Value}
	docUpdate := bson.M{"$set": bson.M{"is_logged_in": true}}
	err := usersCollection.Update(docSelector, docUpdate)
	if err != nil {
		fmt.Println(err)
	}
	//usersCollection.Remove(bson.M{"_id":})
	newCookie := http.Cookie{
		Name:   cookieName,
		MaxAge: -1,
	}
	http.SetCookie(writer, &newCookie)
	t.ExecuteTemplate(writer, "login", nil)
}

func registrateHandler(writer http.ResponseWriter, request *http.Request) {

	if request.Method == "GET" {

	} else {
		request.ParseForm()

		userName := request.PostFormValue("username")
		password := request.PostFormValue("password")

		var users []readUserTy
		userExists, _ := usersCollection.Find(bson.M{"username": userName}).Count()

		fmt.Println(len(users))
		if userExists == 0 {

			doc1 := userTy{userName, password, false}
			var errMessage regDataTy
			errMessage.ErrorMessage = "Benutzer erstellt"
			usersCollection.Insert(doc1)
			t.ExecuteTemplate(writer, "registrate", errMessage)

		} else {
			var errMessage regDataTy
			errMessage.ErrorMessage = "Benutzer existiert schon"
			t.ExecuteTemplate(writer, "registrate", errMessage)

		}
	}

}
func urlAjaxHandler(writer http.ResponseWriter, request *http.Request) {
	url := request.URL.Query().Get("url")
	fmt.Println(url)

}
func pressMeHandler(writer http.ResponseWriter, request *http.Request) {
	//ToDo show Login only if no cookie available
	cookie, _ := request.Cookie("pressMe")

	if request.Method == "POST" {

		var users []readUserTy
		request.ParseForm()
		userName := request.PostFormValue("username")
		password := request.PostFormValue("password")

		exits, _ := usersCollection.Find(bson.M{"username": userName, "password": password}).Count()
		fmt.Println("exitst:", exits)
		if exits == 1 {
			usersCollection.Find(bson.M{"username": userName, "password": password}).All(&users)
			docSelector := bson.M{"username": userName, "password": password}
			docUpdate := bson.M{"$set": bson.M{"is_logged_in": true}}
			err := usersCollection.Update(docSelector, docUpdate)
			if err != nil {
				fmt.Println(err)
			}
			Id = users[0].ID.Hex()
			fmt.Println("hex Id", users[0].ID.Hex())
			fmt.Println("string Id", users[0].ID.String())
			setCookie := http.Cookie{
				Name:  cookieName,
				Value: Id,
				Path:  "/",
			}
			fmt.Println(&setCookie)
			http.SetCookie(writer, &setCookie)
			//ToDo get bookmarks from DB
			t.ExecuteTemplate(writer, "bookmarks", nil)
		}
	} else if request.Method == "GET" {
		if cookie != nil {
			t.ExecuteTemplate(writer, "bookmarks", nil)
		} else {
			t.ExecuteTemplate(writer, "login", nil)
		}

	}

}
