package main

import (
	"crypto/tls"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
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

type messageTy struct {
	Message string
}

type bookmarkTy struct {
	URL         string   `bson:"url" json:"url"`
	ShortReview string   `bson:"shortReview" json:"shortReview"`
	TitleText   string   `bson:"titleText" json:"title_text"`
	Title       string   `bson:"title" json:"title"`
	Images      []string `bson:"images" json:"images"`
	IconName    string   `bson:"icon" json:"icon"`
	Categories  []string `bson:"categories" json:"categories"`
	Position    string   `bson:"position" json:"position"`
}
type userBookmarks struct {
	UserId    bson.ObjectId `json:"user_id" bson:"user_id"`
	Bookmarks []bookmarkTy  `json:"bookmarks" bson:"bookmarks"`
}
type bookmarksTy struct {
	Bookmarks []bookmarkTy
}
type processUrlFinished struct {
	Url        []*url.URL
	Attributes attributesTy
}
type attributesTy struct {
	imgSrcs     map[int]string
	title       string
	description string
	keywords    []string
}

var imgSrcs = make(map[int]string)
var attributes attributesTy

var usersCollection *mgo.Collection
var bookmarkCollection *mgo.Collection
var favIconsGridFs *mgo.GridFS
var processedFinishedChannel = make(chan processUrlFinished)
var tempImageGridFs *mgo.GridFS

func main() {
	attributes.imgSrcs = make(map[int]string)
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
	favIconsGridFs = db.GridFS("favicons")
	tempImageGridFs = db.GridFS("temp")
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/drueckMich", pressMeHandler)
	http.HandleFunc("/Url", urlAjaxHandler)
	http.HandleFunc("/registrate", registrateHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/deleteAccount", deleteAccountHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/gridGetIcon", getIconFromGrid)
	http.ListenAndServe(":4242", nil)
}
func check_ResponseToHTTP(err error, writer http.ResponseWriter) {
	if err != nil {
		fmt.Fprintln(writer, err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}
func getIconFromGrid(writer http.ResponseWriter, request *http.Request) {
	iconName := request.URL.Query().Get("fileName")
	fmt.Println("iconNamein GetRequest:", iconName)
	if len(iconName) != 0 {
		fmt.Println("iconName:", iconName)
		img, err := favIconsGridFs.Open(iconName)
		check(err)
		writer.Header().Add("Content-Type", img.ContentType())
		_, err = io.Copy(writer, img)
		check_ResponseToHTTP(err, writer)
		err = img.Close()
		check_ResponseToHTTP(err, writer)
	}

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

	attributes = getAllAttributes(docZeiger)
	processUrlFinished.Attributes = attributes

	getAndSaveFavicon(pageUrl, attributes.title)
	// Alle relativen SRC-URLs in absolute URLs wandeln:
	// https://golang.org/pkg/net/url/#example_URL_Parse

	// Zunächst die pageUrl (raw-url) in eine URL-structure wandeln:
	u, err := url.Parse(pageUrl)
	if err != nil {
		fmt.Println(err)
	}

	// Nun alle URLs aus der Map im Kontext der pageUrl u in
	// absolute URLS konvertieren:

	for _, wert := range attributes.imgSrcs {
		absURL, err := u.Parse(wert)
		if err != nil {
			fmt.Println(err)
		}

		//ToDo extract gps information from images

		processUrlFinished.Url = append(processUrlFinished.Url, absURL)

	}

	processedFinishedChannel <- processUrlFinished

}

func getAllAttributes(node *html.Node) attributesTy {

	if node.Type == html.ElementNode {
		switch node.Data {
		case "img":

			for _, img := range node.Attr {
				if img.Key == "src" {
					// mit nächstem int-Index auf Map pushen:
					if len(attributes.imgSrcs) == 0 {
						attributes.imgSrcs[0] = img.Val
					} else {
						attributes.imgSrcs[len(attributes.imgSrcs)] = img.Val
					}
				}
			}
			break
		case "title":
			attributes.title = node.FirstChild.Data
			break
		case "meta":

			for _, meta := range node.Attr {

				switch meta.Val {

				case "description":

					attributes.description = node.Attr[1].Val
					break

				case "keywords":

					keywords := strings.Split(node.Attr[1].Val, " ")
					for _, keyword := range keywords {
						attributes.keywords = append(attributes.keywords, keyword)
					}
				}
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {

		getAllAttributes(child)
	}
	return attributes
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
	fmt.Println("urlAjaxHandler")
	var bookmarks userBookmarks
	var imgUrl = processUrlFinished{}
	var bookmark bookmarkTy

	Url := request.URL.Query().Get("url")

	oldCookie, _ := request.Cookie("pressMe")
	Id := oldCookie.Value
	docSelector := bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value)}
	bookmark.URL = Url
	go getAndProcessPage(Url)
	//toDo find a better name
	imgUrl = <-processedFinishedChannel

	attributes = attributesTy{}
	attributes.imgSrcs = make(map[int]string)
	go extractPosition(imgUrl.Url)
	exits, err := bookmarkCollection.Find(bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value)}).Count()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("exits: ", exits)
	if exits == 1 {
		for _, img := range imgUrl.Url {
			bookmark.Images = append(bookmark.Images, img.String())
			fmt.Println("imgUrl:", img.String())

		}
		bookmark.ShortReview = imgUrl.Attributes.description
		bookmark.Categories = imgUrl.Attributes.keywords
		bookmark.IconName = imgUrl.Attributes.title
		bookmark.TitleText = imgUrl.Attributes.title
		bookmark.Title = imgUrl.Attributes.title
		docUpdate := bson.M{"$addToSet": bson.M{"bookmarks": bookmark}}
		err = bookmarkCollection.Update(docSelector, docUpdate)
		if err != nil {
			fmt.Println("updateCollectionError")
			fmt.Println(err)
		}

	} else {
		var item bookmarkTy
		for _, img := range imgUrl.Url {
			item.Images = append(item.Images, img.String())
			fmt.Println("imgUrl:", img.String())

		}
		item.Categories = imgUrl.Attributes.keywords
		item.URL = Url
		item.IconName = imgUrl.Attributes.title
		item.TitleText = imgUrl.Attributes.title
		item.Categories = imgUrl.Attributes.keywords
		item.ShortReview = imgUrl.Attributes.description
		bookmarks.Bookmarks = append(bookmarks.Bookmarks, item)
		bookmarks.UserId = bson.ObjectIdHex(Id)
		fmt.Println(bookmarks)
		err := bookmarkCollection.Insert(bookmarks)
		check(err)
	}

}

func extractPosition(urls []*url.URL) {

	for _, url := range urls {
		res, err := http.Get(url.String())
		check(err)
		file, err := tempImageGridFs.Create("tmp")
		_, err = io.Copy(file, res.Body)
		check(err)
		err = file.Close()
		check(err)
		file, err = tempImageGridFs.Open("tmp")
		check(err)

		x, err := exif.Decode(file)
		if err != nil {
			fmt.Println("Datei enthält keine exif-Daten, Programm wird beendet:")

		} else {
			latitude, longitude, _ := x.LatLong()
			fmt.Println("Geografische Breite: ", latitude)
			fmt.Println("Geografische Länge: ", longitude)

			breite, _ := x.Get("ImageWidth")
			hoehe, _ := x.Get("ImageLength")
			fmt.Println("\nBreite des Bildes: ", breite)
			fmt.Println("Höhe des Bildes: ", hoehe)
		}

	}
}

func getAndSaveFavicon(Url string, title string) {
	fmt.Println("getAndSaveFavicon")
	faviconUrl := "https://www.google.com/s2/favicons?domain=" + Url
	res, err := http.Get(faviconUrl)
	if err != nil {
		fmt.Println(err)
	}

	gridFile, err := favIconsGridFs.Create(title)
	if err != nil {
		fmt.Println("error while creating:", err)
	}
	_, err = io.Copy(gridFile, res.Body)

	if err != nil {
		fmt.Println("error whily copy", err)
	}
	err = gridFile.Close()
	if err != nil {
		fmt.Println("error close")
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
				IconName:    doc1.IconName,
				Images:      doc1.Images,
				Categories:  doc1.Categories,
				Position:    doc1.Position,
			}
			entries.Bookmarks = append(entries.Bookmarks, item)
		}

	}

	return entries
}
