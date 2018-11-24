package main

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

var Id string
var cookieName string
var t = template.Must(template.ParseFiles(filepath.Join("./", "tpl", "head.html"),
	filepath.Join("./", "tpl", "login.html"), filepath.Join("./", "tpl", "bookmarks.html"),
	filepath.Join("./", "tpl", "end.html"), filepath.Join("./", "tpl", "registrate.html")))

type userTy struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}
type readUserTy struct {
	ID       bson.ObjectId `bson:"_id"`
	Username string        `bson:"username"`
	Password string        `bson:"password"`
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
	Icon        string       `bson:"icon" json:"icon"`
	Images      []string     `bson:"images" json:"images"`
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

var imgSrcs = make(map[int]string)

type processUrlFinished struct {
	Url []*url.URL
}

var usersCollection *mgo.Collection
var bookmarkCollection *mgo.Collection
var gridFs *mgo.GridFS
var processedFinishedChannel = make(chan processUrlFinished)

func main() {
	cookieName = "pressMe"
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{"bernd.documents.azure.com:10255"}, // Get HOST + PORT
		Timeout:  60 * time.Second,
		Database: "bernd",                                                                                    // It can be anything
		Username: "bernd",                                                                                    // Username
		Password: "N3bwjITyGviFrY8oEBJTIZ9gF7U8y5oDj4cON33a5EZ0SrjRBIhpeqYanASoduVcGjp1S09eAZuuH5sZFDoOcQ==", // PASSWORD
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		},
	}
	dbSession, _ := mgo.DialWithInfo(dialInfo)
	defer dbSession.Close()
	dbSession.SetSafe(&mgo.Safe{})
	db := dbSession.DB("drückMich")
	usersCollection = db.C("users")
	bookmarkCollection = db.C("bookmarks")
	gridFs = db.GridFS("favicons")
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/drueckMich", pressMeHandler)
	http.HandleFunc("/Url", urlAjaxHandler)
	http.HandleFunc("/registrate", registrateHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/deleteAccount", deleteAccountHandler)
	http.HandleFunc("/update", updateHandler)
	http.ListenAndServe(":4242", nil)
}

func updateHandler(writer http.ResponseWriter, request *http.Request) {

}
func getAndProcessPage(pageUrl string) {
	fmt.Println("getAndProcessPage")

	var processUrlFinished = processUrlFinished{}

	// HTTP-GET Request senden:
	res, err := http.Get(pageUrl)
	if err != nil {
		fmt.Println(err)
	}
	byteArrayPage, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Println(err)
	}

	// Empfangene Seite parsen, in doc-tree wandeln:
	docZeiger, err := html.Parse(strings.NewReader(string(byteArrayPage)))
	if err != nil {
		fmt.Println(err)
	}
	for _, node := range docZeiger.Data {
		fmt.Println("data:", node)
	}
	//ToDo refactor sucheAlleImgAttributwerte such that the function get All the desired elements
	imgSrcs := getAllAttributes(docZeiger)

	// Alle relativen SRC-URLs in absolute URLs wandeln:
	// https://golang.org/pkg/net/url/#example_URL_Parse

	// Zunächst die pageUrl (raw-url) in eine URL-structure wandeln:
	u, err := url.Parse(pageUrl)
	if err != nil {
		fmt.Println(err)
	}

	// Nun alle URLs aus der Map im Kontext der pageUrl u in
	// absolute URLS konvertieren:

	for _, wert := range imgSrcs {
		absURL, err := u.Parse(wert)
		if err != nil {
			fmt.Println(err)
		}

		//ToDo extract gps information from images

		processUrlFinished.Url = append(processUrlFinished.Url, absURL)

	}

	processedFinishedChannel <- processUrlFinished

}

func getAllAttributes(node *html.Node) map[int]string {

	if node.Type == html.ElementNode && node.Data == "img" {
		for _, img := range node.Attr {
			if img.Key == "src" {
				// mit nächstem int-Index auf Map pushen:
				if len(imgSrcs) == 0 {
					imgSrcs[0] = img.Val
				} else {
					imgSrcs[len(imgSrcs)] = img.Val
				}

			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		getAllAttributes(child)
	}

	return imgSrcs
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
	fmt.Println("logout")
	fmt.Println(oldCookie.Value)
	docSelector := bson.M{"_id": bson.ObjectIdHex(oldCookie.Value)}
	docUpdate := bson.M{"$set": bson.M{"is_logged_in": false}}
	err := usersCollection.Update(docSelector, docUpdate)
	if err != nil {
		fmt.Println(err)
	}

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

			doc1 := userTy{userName, password}
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
	var imgUrl = processUrlFinished{}
	var bookmark bookmarkTy

	Url := request.URL.Query().Get("url")

	oldCookie, _ := request.Cookie("pressMe")
	docSelector := bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value)}
	bookmark.URL = Url

	exits, err := bookmarkCollection.Find(bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value)}).Count()
	if err != nil {
		fmt.Println(err)
	}
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

	go getAndProcessPage(Url)

	//title.Title
	imgUrl = <-processedFinishedChannel
	go getAndSaveFavicon(Url)

	docSelector = bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value), "bookmarks.url": Url}
	bookmarkCollection.Find(docSelector).One(&bookmarks)
	for _, item := range bookmarks.Bookmarks {
		for _, img := range imgUrl.Url {
			if item.URL == Url {
				item.Images = append(item.Images, img.String())
			}
		}
		bookmarks.Bookmarks = append(bookmarks.Bookmarks, item)
	}

	bookmarkCollection.Update(docSelector, bookmarks)

	if err != nil {
		fmt.Println(err.Error())
	}

}

func getAndSaveFavicon(Url string) {
	fmt.Println("getAndSaveFavicon")
	faviconUrl := "https://www.google.com/s2/favicons?" + Url
	fmt.Println("faviconUrl:", faviconUrl)
	res, err := http.Get(faviconUrl)
	if err != nil {
		fmt.Println(err)
	}
	go getTitle()
	gridFile, err := gridFs.Create("example")
	if err != nil {
		fmt.Println(err)
	}
	_, err = io.Copy(gridFile, res.Body)
	if err != nil {
		fmt.Println(err)
	}
	err = gridFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		fmt.Println(err)
	}
}
func pressMeHandler(writer http.ResponseWriter, request *http.Request) {

	cookie, _ := request.Cookie("pressMe")

	if request.Method == "POST" {

		var users []readUserTy
		request.ParseForm()
		userName := request.PostFormValue("username")
		password := request.PostFormValue("password")

		exits, _ := usersCollection.Find(bson.M{"username": userName, "password": password}).Count()
		//user exists?
		if exits == 1 {
			err := usersCollection.Find(bson.M{"username": userName, "password": password}).All(&users)
			if err != nil {
				fmt.Println(err)
			}

			if err != nil {
				log.Fatal(err)
			}
			Id = users[0].ID.Hex()

			setCookie := http.Cookie{
				Name:  cookieName,
				Value: Id,
				Path:  "/",
			}

			http.SetCookie(writer, &setCookie)

			err = t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries())
			if err != nil {
				log.Fatal(err)
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
func getBookmarksEntries() bookmarksTy {
	var entries = bookmarksTy{
		Bookmarks: []bookmarkTy{},
	}
	var docs []userBookmarks
	err := bookmarkCollection.Find(bson.M{"user_id": bson.ObjectIdHex(Id)}).All(&docs)
	if err != nil {
		log.Fatal(err)
	}
	for _, doc := range docs {
		for _, doc1 := range doc.Bookmarks {

			item := bookmarkTy{
				URL:         doc1.URL,
				ShortReview: doc1.ShortReview,
				TitleText:   doc1.TitleText,
				Images:      doc1.Images,
				Categories:  doc1.Categories,
				Position:    doc1.Position,
			}
			entries.Bookmarks = append(entries.Bookmarks, item)
		}

	}
	return entries
}
