package main

import (
	"encoding/json"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/watson-developer-cloud/go-sdk/core"
	"github.com/watson-developer-cloud/go-sdk/visualrecognitionv3"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var t = template.Must(template.ParseFiles(filepath.Join("./", "tpl", "head.html"),
	filepath.Join("./", "tpl", "login.html"), filepath.Join("./", "tpl", "bookmarks.html"),
	filepath.Join("./", "tpl", "end.html"), filepath.Join("./", "tpl", "registrate.html"),
	filepath.Join("./", "tpl", "modal.html")))

type userTy struct {
	Username            string       `bson:"username"`
	Password            string       `bson:"password"`
	AvailableCategories []CategoryTy `bson:"available_categories" json:"available_categories"`
	Bookmarks           []bookmarkTy `bson:"bookmarks" json:"bookmarks"`
}
type readUserTy struct {
	ID                  bson.ObjectId `bson:"_id"`
	Username            string        `bson:"username"json:"username"`
	Password            string        `bson:"password" json:"password"`
	AvailableCategories []string      `bson:"available_categories" json:"available_categories"`
	Bookmarks           []bookmarkTy  `bson:"bookmarks" json:"bookmarks"`
}

type messageTy struct {
	Message string
}

type bookmarkTy struct {
	URL              string   `bson:"url" json:"url"`
	ShortReview      string   `bson:"shortReview" json:"shortReview"`
	Title            string   `bson:"title" json:"title"`
	Images           []string `bson:"images" json:"images"`
	IconName         string   `bson:"icon" json:"icon"`
	WVRCategories    []string `bson:"wvrcategories" json:"wvr_categories"`
	CustomCategories []string `bson:"customcategories" json:"custom_categories"`
	Keywords         []string `bsonm:"keywords" json:"keywords"`
	Lat              float64  `bson:"lat" json:"lat"`
	Long             float64  `bson:"long" json:"long"`
}
type coordinates struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lon float64 `json:"lon" bson:"lon"`
}

type data struct {
	AvailableCategories []CategoryTy
	Bookmarks           []bookmarkTy `json:"bookmarks"`
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
type channelData struct {
	Url         string
	Docselector bson.M
}
type CategoryTy struct {
	Category string
}
type categories struct {
	Categories []string
}
type cookieTy struct {
	OrderMethod, Orderby string `json:"order_method"`
	customOrder          bool   `json:"custom_order"`
}
type formDataTy struct {
	Orderby string
}

var attributes attributesTy

var usersCollection *mgo.Collection
var favIconsGridFs *mgo.GridFS

var tempImageGridFs *mgo.GridFS
var coordinatesChannel = make(chan coordinates, 2)
var dataChannel = make(chan channelData, 2)
var categoriesChannel = make(chan categories)
var service *visualrecognitionv3.VisualRecognitionV3
var serviceErr error
var Id string
var orderCookieName string
var sessionCookieName string

func main() {
	byteArray, _ := ioutil.ReadFile("apiKey")
	apiKey := string(byteArray)
	attributes.imgSrcs = make(map[int]string)
	sessionCookieName = "pressMe"
	orderCookieName = "orderMethod"

	service, serviceErr = visualrecognitionv3.NewVisualRecognitionV3(&visualrecognitionv3.VisualRecognitionV3Options{
		URL:       "https://gateway.watsonplatform.net/visual-recognition/api",
		Version:   "2018-03-19",
		IAMApiKey: apiKey,
	})
	if serviceErr != nil {
		panic(serviceErr)
	}
	dbSession, err := mgo.Dial("localhost")
	check(err)
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
	http.HandleFunc("/gridGetIcon/", getIconFromGrid)
	http.HandleFunc("/newCategory", newCategoryHandler)
	http.HandleFunc("/setSortProperties", sortPropertiesHandler)
	http.ListenAndServe(":4242", nil)
}

func sortPropertiesHandler(writer http.ResponseWriter, request *http.Request) {
	//ToDo
	var newCookie http.Cookie

	orderBy := request.PostFormValue("orderBy")
	newCookie = http.Cookie{
		Name:  orderCookieName,
		Value: "orderBy=" + orderBy,
		Path:  "/",
	}

	http.SetCookie(writer, &newCookie)

}

func newCategoryHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("newCategory")
	cookie, err := request.Cookie("pressMe")
	check(err)
	var Id string
	var user readUserTy
	if cookie != nil {

		catName := request.PostFormValue("catName")

		Id = cookie.Value

		docSelector := bson.M{"_id": bson.ObjectIdHex(Id)}
		err := usersCollection.Find(docSelector).One(&user)
		check(err)
		user.AvailableCategories = append(user.AvailableCategories, catName)
		err = usersCollection.Update(docSelector, user)
		check(err)
		fmt.Fprint(writer, "sucess")
	} else {

	}
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

//todo no app crash when cookie exits but no entry in db
func getIconFromGrid(writer http.ResponseWriter, request *http.Request) {
	iconName := request.URL.Query().Get("fileName")

	if len(iconName) != 0 {
		img, err := favIconsGridFs.Open(iconName)
		check(err)
		//by now no Content-Type  is saved in DB
		writer.Header().Add("Content-Type", img.ContentType())
		_, err = io.Copy(writer, img)
		check_ResponseToHTTP(err, writer)
		err = img.Close()
		check_ResponseToHTTP(err, writer)
	}

}

func updateHandler(writer http.ResponseWriter, request *http.Request) {
	//ToDo no cookie
	var jsonString string
	cookie, _ := request.Cookie("pressMe")
	orderMethodCookie, _ := request.Cookie(orderCookieName)

	Id = cookie.Value
	if orderMethodCookie != nil {
		fmt.Println(orderMethodCookie.Value)
		bytestring, err := json.Marshal(getBookmarksEntries(orderMethodCookie.Value))
		if err != nil {
			fmt.Println(err)
		}
		jsonString = string(bytestring)
	} else {
		bytestring, err := json.Marshal(getBookmarksEntries(""))
		if err != nil {
			fmt.Println(err)
		}
		jsonString = string(bytestring)
	}

	fmt.Fprint(writer, jsonString)
}
func getAndProcessPage() {
	channelData := <-dataChannel
	docSelector := channelData.Docselector
	pageUrl := channelData.Url
	fmt.Println("getAndProcessPage")
	var user readUserTy
	var imgUrls []*url.URL

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
	attributes = attributesTy{}
	attributes.imgSrcs = make(map[int]string)
	attributes = getAllAttributes(docZeiger)

	go getAndSaveFavicon(pageUrl, attributes.title)
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

		imgUrls = append(imgUrls, absURL)

		err = usersCollection.Find(docSelector).One(&user)
		check(err)

		for i := range user.Bookmarks {
			if user.Bookmarks[i].URL == pageUrl {
				user.Bookmarks[i].Images = append(user.Bookmarks[i].Images, absURL.String())
				user.Bookmarks[i].Keywords = attributes.keywords
				user.Bookmarks[i].Title = attributes.title
				user.Bookmarks[i].IconName = attributes.title
				user.Bookmarks[i].ShortReview = attributes.description
			}

		}

	}

	err = usersCollection.Update(docSelector, user)
	check(err)
	go extractPosition(imgUrls)

	coordinates := <-coordinatesChannel
	err = usersCollection.Find(docSelector).One(&user)
	check(err)

	for i := range user.Bookmarks {
		if user.Bookmarks[i].URL == pageUrl {
			user.Bookmarks[i].Lat = coordinates.Lat
			user.Bookmarks[i].Long = coordinates.Lon
		}
	}
	err = usersCollection.Update(docSelector, user)
	go classesRecoginition(imgUrls)
	categories := <-categoriesChannel

	err = usersCollection.Find(docSelector).One(&user)
	check(err)
	for i := range user.Bookmarks {
		if user.Bookmarks[i].URL == pageUrl {

			for _, category := range categories.Categories {
				user.Bookmarks[i].WVRCategories = append(user.Bookmarks[i].WVRCategories, category)
			}

		}
	}
	err = usersCollection.Update(docSelector, user)

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

	docSelector := bson.M{"_id": bson.ObjectIdHex(oldCookie.Value)}
	docUpdate := bson.M{"$set": bson.M{"is_logged_in": false}}
	err := usersCollection.Update(docSelector, docUpdate)
	if err != nil {
		fmt.Println(err)
	}

	newCookie := http.Cookie{
		Name:   sessionCookieName,
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

		userExists, _ := usersCollection.Find(bson.M{"username": userName}).Count()

		if userExists == 0 {

			doc1 := userTy{userName, password, []CategoryTy{}, []bookmarkTy{}}
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
func urlAjaxHandler(_ http.ResponseWriter, request *http.Request) {
	//ToDo check if entry alread exists
	fmt.Println("urlAjaxHandler")
	var user readUserTy
	var bookmark bookmarkTy

	Url := request.URL.Query().Get("url")
	oldCookie, _ := request.Cookie("pressMe")
	docSelector := bson.M{"_id": bson.ObjectIdHex(oldCookie.Value)}
	var channelData = channelData{
		Url:         Url,
		Docselector: docSelector,
	}
	go getAndProcessPage()
	dataChannel <- channelData
	err := usersCollection.Find(docSelector).One(&user)
	check(err)
	bookmark.URL = Url
	docUpdate := bson.M{"$addToSet": bson.M{"bookmarks": bookmark}}
	err = usersCollection.Update(docSelector, docUpdate)
	check(err)

	//toDo find a better name

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
		result := struct{ INode int }{}
		err = file.GetMeta(&result)
		if err != nil {
			panic(err)
		}

		check(err)
		file, err = tempImageGridFs.Open("tmp")
		check(err)

		x, err := exif.Decode(file)
		if err != nil {

			if (len(urls) - 1) == i {

				coordinatesChannel <- coordinates
			}
		} else {

			latitude, longitude, _ := x.LatLong()

			coordinates.Lat = latitude
			coordinates.Lon = longitude

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

	mimeType := res.Header.Get("Content-Type")
	gridFile, err := favIconsGridFs.Create(title)
	gridFile.SetContentType(mimeType)

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
	sessionCookie, _ := request.Cookie("pressMe")

	orderCookie, _ := request.Cookie(orderCookieName)
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
				Name:  sessionCookieName,
				Value: Id,
				Path:  "/",
			}
			http.SetCookie(writer, &setCookie)
			err = t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries(""))
			if err != nil {
				log.Fatal(err)
			}

		} else {
			var errMessage messageTy
			errMessage.Message = "Benutzer existiert nicht"
			t.ExecuteTemplate(writer, "login", errMessage)
		}
	} else if request.Method == "GET" {
		if sessionCookie != nil {
			Id = sessionCookie.Value
			if orderCookie != nil {
				t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries(orderCookie.Value))
			} else {
				t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries(""))
			}

		} else {
			t.ExecuteTemplate(writer, "login", nil)
		}

	}

}
func getBookmarksEntries(orderMethod string) readUserTy {
	var doc readUserTy
	fmt.Println("getBookmarkEntries")
	fmt.Println("orderMethod:", orderMethod)

	err := usersCollection.Find(bson.M{"_id": bson.ObjectIdHex(Id)}).One(&doc)
	if err != nil {
		log.Fatal(err)
	}
	if len(orderMethod) > 0 {
		parts := strings.Split(orderMethod, "=")

		if parts[1] == "0" {
			sort.Slice(doc.Bookmarks, func(i, j int) bool {
				return doc.Bookmarks[i].Title < doc.Bookmarks[j].Title
			})

		}

	}

	return readUserTy{AvailableCategories: doc.AvailableCategories, Bookmarks: doc.Bookmarks}
}
func classesRecoginition(urls []*url.URL) {
	var categories = categories{}
	for _, url := range urls {
		// Optionen für die Klassifizierung festlegen:
		classifyOptions := service.NewClassifyOptions()
		classifyOptions.URL = core.StringPtr(url.String())

		// Schwellwert für den "Verlässlichkeitsscore":
		classifyOptions.Threshold = core.Float32Ptr(0.6)

		classifyOptions.ClassifierIds = []string{"default"}
		//	classifyOptions.ClassifierIds = []string{"default", "food", "explicit"}

		// Ausgabesprache definieren:
		sprache := new(string)
		*sprache = "de"
		classifyOptions.AcceptLanguage = sprache

		// Classify Dienst aufrufen:
		response, responseErr := service.Classify(classifyOptions)
		if responseErr != nil {
			log.Println(responseErr)
		}

		// Ergebnisdaten aufbereiten:
		classifyResult := service.GetClassifyResult(response)

		if classifyResult != nil {

			// Einzelne Datenelemente aus dem Ergebnis extrahieren:
			classes := classifyResult.Images[0].Classifiers[0].Classes
			imageName := *classifyResult.Images[0].ResolvedURL
			_, imageName = path.Split(imageName) // path.Split NICHT string.Split !!!!!
			//typHierarchie := *classes[0].TypeHierarchy

			//fmt.Printf("\n------------------Das Image %s wurde in die Typ-Hierarchie %s eingeordnet.\n", imageName, typHierarchie)

			for _, wert := range classes {

				categories.Categories = append(categories.Categories, *wert.ClassName)
			}
		}
		categoriesChannel <- categories
		return
	}

}
