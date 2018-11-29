package main

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
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
	Username         string       `bson:"username"`
	Password         string       `bson:"password"`
	CustomCategories []string     `bson:"custom_categories" json:"custom_categories"`
	Bookmarks        []bookmarkTy `bson:"bookmarks"`
}
type readUserTy struct {
	ID               bson.ObjectId `bson:"_id"`
	Username         string        `bson:"username"`
	Password         string        `bson:"password"`
	CustomCategories []string      `bson:"custom_categories" json:"custom_categories"`
	Bookmarks        []bookmarkTy  `bson:"bookmarks"`
}

type messageTy struct {
	Message string
}

type bookmarkTy struct {
	URL         string      `bson:"url" json:"url"`
	ShortReview string      `bson:"shortReview" json:"shortReview"`
	TitleText   string      `bson:"titleText" json:"title_text"`
	Title       string      `bson:"title" json:"title"`
	Images      []string    `bson:"images" json:"images"`
	IconName    string      `bson:"icon" json:"icon"`
	Categories  []string    `bson:"categories" json:"categories"`
	Coordinates coordinates `bson:"position" json:"position"`
}
type coordinates struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lon float64 `json:"lon" bson:"lon"`
}
type userBookmarks struct {
	UserId    bson.ObjectId `json:"user_id" bson:"user_id"`
	Bookmarks []bookmarkTy  `json:"bookmarks" bson:"bookmarks"`
}
type bookmarksTy struct {
	Bookmarks []bookmarkTy `json:"bookmarks"`
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
var favIconsGridFs *mgo.GridFS
var processedFinishedChannel = make(chan processUrlFinished)
var coordinatesChannel = make(chan coordinates)
var tempImageGridFs *mgo.GridFS

func main() {
	attributes.imgSrcs = make(map[int]string)
	cookieName = "pressMe"

	dbSession, _ := mgo.Dial("localhost")
	defer dbSession.Close()
	dbSession.SetSafe(&mgo.Safe{})
	db := dbSession.DB("drückMich")
	usersCollection = db.C("users")

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
	cookie, _ := request.Cookie("pressMe")
	Id = cookie.Value
	//ToDo send json instead of executing template
	/*bytestring, err := json.Marshal(getBookmarksEntries())
	if err != nil {
		fmt.Println(err)
	}
	jsonString := string(bytestring)
	fmt.Fprint(writer, jsonString)*/
	t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries())
}
func getAndProcessPage(pageUrl string, docSelector bson.M) {
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

			doc1 := userTy{userName, password, nil, []bookmarkTy{}}
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
	var user readUserTy

	var imgUrl = processUrlFinished{}
	var bookmark bookmarkTy

	Url := request.URL.Query().Get("url")

	oldCookie, _ := request.Cookie("pressMe")

	docSelector := bson.M{"_id": bson.ObjectIdHex(oldCookie.Value)}
	err := usersCollection.Find(docSelector).One(&user)
	check(err)
	bookmark.URL = Url
	docUpdate := bson.M{"$addToSet": bson.M{"bookmarks": bookmark}}
	err = usersCollection.Update(docSelector, docUpdate)
	check(err)
	go getAndProcessPage(Url, docSelector)
	//toDo find a better name
	imgUrl = <-processedFinishedChannel
	attributes = attributesTy{}
	attributes.imgSrcs = make(map[int]string)

	err = usersCollection.Find(docSelector).One(&user)
	check(err)
	fmt.Println("user:")
	fmt.Printf("%+v\n", user)
	for _, img := range imgUrl.Url {
		for i := range user.Bookmarks {
			if user.Bookmarks[i].URL == Url {
				user.Bookmarks[i].Images = append(user.Bookmarks[i].Images, img.String())
				user.Bookmarks[i].Categories = imgUrl.Attributes.keywords
				user.Bookmarks[i].URL = Url
				user.Bookmarks[i].IconName = imgUrl.Attributes.title
				user.Bookmarks[i].TitleText = imgUrl.Attributes.title
				user.Bookmarks[i].Categories = imgUrl.Attributes.keywords
				user.Bookmarks[i].ShortReview = imgUrl.Attributes.description
			}

		}
	}
	fmt.Println("user:")
	fmt.Printf("%+v\n", user)
	err = usersCollection.Update(docSelector, user)
	check(err)
	go extractPosition(imgUrl.Url)
	coordinates := <-coordinatesChannel
	err = usersCollection.Find(docSelector).One(&user)
	check(err)
	fmt.Println("user:")
	fmt.Printf("%+v\n", user)

	for i := range user.Bookmarks {
		if user.Bookmarks[i].URL == Url {
			user.Bookmarks[i].Coordinates = coordinates
		}

	}
	err = usersCollection.Update(docSelector, user)
}

func extractPosition(urls []*url.URL) {
	var coordinates = coordinates{}

	for i, url := range urls {
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
			if (len(urls) - 1) == i {

				coordinatesChannel <- coordinates
			}
		} else {

			latitude, longitude, _ := x.LatLong()
			fmt.Println("Geografische Breite: ", latitude)
			fmt.Println("Geografische Länge: ", longitude)
			coordinates.Lat = latitude
			coordinates.Lon = longitude
			breite, _ := x.Get("ImageWidth")
			hoehe, _ := x.Get("ImageLength")
			fmt.Println("\nBreite des Bildes: ", breite)
			fmt.Println("Höhe des Bildes: ", hoehe)
			coordinatesChannel <- coordinates
			return
		}
		tempImageGridFs.Remove("tmp")
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
	//ToDo change example
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
			Id = cookie.Value
			t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries())
		} else {
			t.ExecuteTemplate(writer, "login", nil)
		}

	}

}
func getBookmarksEntries() bookmarksTy {

	var doc readUserTy

	err := usersCollection.Find(bson.M{"_id": bson.ObjectIdHex(Id)}).One(&doc)
	if err != nil {
		log.Fatal(err)
	}

	return bookmarksTy{Bookmarks: doc.Bookmarks}
}
