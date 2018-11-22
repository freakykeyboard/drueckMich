package main

import (
	"fmt"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
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
type messageTy struct {
	Message string
}
type categoryTy struct {
	Name string `bson:"name" json:"name"`
}
type bookmarkTy struct {
	URL         string       `bson:"url" json:"url"`
	ShortReview string       `bson:"shortReview" json:"shortReview"`
	TitleText   string       `bson:"titleText" json:"title_text"`
	Categories  []categoryTy `bson:"categories" json:"categories"`
	Position    string       `bson:"position" json:"position"`
}
type userBookmarks struct {
	UserId    bson.ObjectId `json:"user_id" bson:"user_id"`
	Bookmarks []bookmarkTy  `json:"bookmarks" bson:"bookmarks"`
}
type bookmarksTy struct {
	Bookmarks []bookmarkTy
}

var usersCollection *mgo.Collection
var bookmarkCollection *mgo.Collection

func main() {
	cookieName = "pressMe"
	dbSession, _ := mgo.Dial("localhost")
	defer dbSession.Close()
	db := dbSession.DB("drückMich")
	usersCollection = db.C("users")
	bookmarkCollection = db.C("bookmarks")
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/drueckMich", pressMeHandler)
	http.HandleFunc("/Url", urlAjaxHandler)
	http.HandleFunc("/registrate", registrateHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/deleteAccount", deleteAccountHandler)
	http.ListenAndServe(":4242", nil)
}
func getAndProcessPage(pageUrl string) {
	var imgSrcs = make(map[int]string)
	// Seite anfordern:

	//	pageUrl := "http://localhost/webscraperTest.html"

	// HTTP-GET Request senden:
	res, err := http.Get(pageUrl)
	if err != nil {
		log.Fatal(err)
	}
	byteArrayPage, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Empfangene Seite parsen, in doc-tree wandeln:
	docZeiger, err := html.Parse(strings.NewReader(string(byteArrayPage)))
	if err != nil {
		log.Fatal(err)
	}
	suchAlleImgSrcAttributwerte(docZeiger)

	// Ausgabe aller, vermutlich meistens relativen, SRC-URLs:
	for _, wert := range imgSrcs {
		fmt.Println(wert)
	}

	fmt.Println("---------------------------------------------------------------")

	// Alle relativen SRC-URLs in absolute URLs wandeln:
	// https://golang.org/pkg/net/url/#example_URL_Parse

	// Zunächst die pageUrl (raw-url) in eine URL-structure wandeln:
	u, err := url.Parse(pageUrl)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u)
	fmt.Println("---------------------------------------------------------------")

	// Nun alle URLs aus der Map im Kontext der pageUrl u in
	// absolute URLS konvertieren:
	for _, wert := range imgSrcs {
		absURL, err := u.Parse(wert)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(absURL)
	}

}
func suchAlleImgSrcAttributwerte(node *html.Node) {
	var imgSrcs = make(map[int]string)
	if node.Type == html.ElementNode && node.Data == "img" {
		for _, img := range node.Attr {
			if img.Key == "src" {
				// mit nächstem int-Index auf Map pushen:
				if len(imgSrcs) == 0 {
					imgSrcs[0] = img.Val
				} else {
					imgSrcs[len(imgSrcs)] = img.Val
				}
				break
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		suchAlleImgSrcAttributwerte(child)
	}
}

func deleteAccountHandler(writer http.ResponseWriter, request *http.Request) {
	var message messageTy
	oldCookie, _ := request.Cookie("pressMe")
	docSelector := bson.M{"_id": bson.ObjectIdHex(oldCookie.Value)}
	usersCollection.Remove(docSelector)
	message.Message = "Account gelöscht"
	t.ExecuteTemplate(writer, "login", message)
}
func logoutHandler(writer http.ResponseWriter, request *http.Request) {
	oldCookie, _ := request.Cookie("pressMe")
	docSelector := bson.M{"_id": bson.ObjectIdHex(oldCookie.Value)}
	docUpdate := bson.M{"$set": bson.M{"is_logged_in": false}}
	err := usersCollection.Update(docSelector, docUpdate)
	if err != nil {
		fmt.Println(err)
	}

	/*newCookie := http.Cookie{
		Name:   cookieName,
		MaxAge: -1,
	}
	http.SetCookie(writer, &newCookie)*/
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
			var errMessage messageTy
			errMessage.Message = "Benutzer erstellt"
			usersCollection.Insert(doc1)
			t.ExecuteTemplate(writer, "registrate", errMessage)

		} else {
			var errMessage messageTy
			errMessage.Message = "Benutzer existiert schon"
			t.ExecuteTemplate(writer, "registrate", errMessage)

		}
	}

}
func urlAjaxHandler(writer http.ResponseWriter, request *http.Request) {
	var bookmarks userBookmarks

	var bookmark bookmarkTy
	url := request.URL.Query().Get("url")

	oldCookie, _ := request.Cookie("pressMe")
	docSelector := bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value)}
	bookmark.URL = url

	exits, _ := bookmarkCollection.Find(docSelector).Count()
	fmt.Println("exits: ", exits)
	if exits == 1 {
		bookmarkCollection.Find(docSelector).One(&bookmarks)
		bookmarks.Bookmarks = append(bookmarks.Bookmarks, bookmark)
		docUpdate := bson.M{"$addToSet": bson.M{"bookmarks": bookmark}}
		err := bookmarkCollection.Update(docSelector, docUpdate)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		bookmarks.UserId = bson.ObjectIdHex(oldCookie.Value)
		bookmarks.Bookmarks = append(bookmarks.Bookmarks, bookmark)
		bookmarkCollection.Insert(bookmarks)
	}

	go getAndProcessPage(url)

}
func pressMeHandler(writer http.ResponseWriter, request *http.Request) {
	//ToDo show Login only if no cookie available
	var bookmarks userBookmarks

	cookie, _ := request.Cookie("pressMe")

	if request.Method == "POST" {

		var users []readUserTy
		request.ParseForm()
		userName := request.PostFormValue("username")
		password := request.PostFormValue("password")

		exits, _ := usersCollection.Find(bson.M{"username": userName, "password": password}).Count()
		//user exists?
		if exits == 1 {
			usersCollection.Find(bson.M{"username": userName, "password": password}).All(&users)
			if users[0].IsLoggedIn == true {
				var errMessage messageTy
				errMessage.Message = "Benutzer schon angemeldet"
				t.ExecuteTemplate(writer, "login", errMessage)
			} else {
				docSelector := bson.M{"username": userName, "password": password}
				docUpdate := bson.M{"$set": bson.M{"is_logged_in": true}}
				err := usersCollection.Update(docSelector, docUpdate)
				if err != nil {
					fmt.Println(err)
				}
				Id = users[0].ID.Hex()

				setCookie := http.Cookie{
					Name:  cookieName,
					Value: Id,
					Path:  "/",
				}
				fmt.Println(&setCookie)
				http.SetCookie(writer, &setCookie)
				//ToDo get bookmarks from DB
				bookmarkCollection.Find(bson.M{"user_id": bson.ObjectIdHex(Id)}).One(&bookmarks)

				var Bookmarks = bookmarksTy{

					Bookmarks: []bookmarkTy{},
				}

				for _, doc := range bookmarks.Bookmarks {
					fmt.Println(doc)
					item := bookmarkTy{
						URL:         doc.URL,
						ShortReview: doc.ShortReview,
						TitleText:   doc.TitleText,
						Categories:  doc.Categories,
						Position:    doc.Position,
					}
					bookmarks.Bookmarks = append(Bookmarks.Bookmarks, item)
				}
				fmt.Println("bookmarks: ", Bookmarks.Bookmarks)
				t.ExecuteTemplate(writer, "bookmarks", Bookmarks.Bookmarks)
			}

		} else {
			var errMessage messageTy
			errMessage.Message = "Benutzer existiert nicht"
			t.ExecuteTemplate(writer, "login", errMessage)
		}
	} else if request.Method == "GET" {
		if cookie != nil {
			t.ExecuteTemplate(writer, "bookmarks", nil)
		} else {
			var message messageTy
			message.Message = "gesendte Url konnte nicht zugeordnent werden. bitte voher anmelden"
			t.ExecuteTemplate(writer, "login", nil)
		}

	}

}
func analyzeUrl() {

}
