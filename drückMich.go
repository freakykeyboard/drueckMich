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
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var t = template.Must(template.ParseFiles(filepath.Join("tpl", "head.html"),
	filepath.Join("tpl", "login.html"), filepath.Join("tpl", "bookmarks.html"),
	filepath.Join("tpl", "end.html"),
	filepath.Join("tpl", "newCategoryModal.html"), filepath.Join("tpl", "removeCategoryModal.html"),
	filepath.Join("tpl", "registrationModal.html"), filepath.Join("tpl", "addCategoryModal.html")))

type userTy struct {
	Username            string   `bson:"username"`
	Password            string   `bson:"password"`
	AvailableCategories []string `bson:"available_categories" json:"available_categories"`
}
type readUserTy struct {
	ID                  bson.ObjectId `bson:"_id"`
	Username            string        `bson:"username"json:"username"`
	Password            string        `bson:"password" json:"password"`
	AvailableCategories []string      `bson:"available_categories" json:"available_categories"`
}
type GeoJsonTy struct {
	Type        string    `json:"-"`
	Coordinates []float64 `json:"coordinates",bson:"coordinates"`
}
type messageTy struct {
	ShowModal bool
	Message   string
}

type bookmarkTy struct {
	UserId           bson.ObjectId `json:"user_id" bson:"user_id"`
	URL              string        `bson:"url" json:"url"`
	ShortReview      string        `bson:"shortReview" json:"shortReview"`
	Title            string        `bson:"title" json:"title"`
	Images           []string      `bson:"images" json:"images"`
	IconName         string        `bson:"icon" json:"icon"`
	WVRCategories    []string      `bson:"wvr_categories" json:"wvr_categories"`
	CustomCategories []string      `bson:"custom_categories" json:"custom_categories"`
	Keywords         []string      `bson:"keywords" json:"keywords"`
	Location         GeoJsonTy     `bson:"location",json:"location"`
	Lat              float64       `bson:"lat" json:"lat"`
	Long             float64       `bson:"long" json:"long"`
	CreationDate     time.Time     `bson:"creation_date",json:"creation_date"`
}
type readBookmarkTy struct {
	ID               bson.ObjectId `bson:"_id" json:"id"`
	UserId           bson.ObjectId `json:"user_id" bson:"user_id"`
	URL              string        `bson:"url" json:"url"`
	ShortReview      string        `bson:"shortReview" json:"shortReview"`
	Title            string        `bson:"title" json:"title"`
	Images           []string      `bson:"images" json:"images"`
	IconName         string        `bson:"icon" json:"icon"`
	WVRCategories    []string      `bson:"wvr_categories" json:"wvr_categories"`
	CustomCategories []string      `bson:"custom_categories" json:"custom_categories"`
	Keywords         []string      `bson:"keywords" json:"keywords"`
	Coordinates      GeoJsonTy     `bson:"coordinates",json:"coordinates"`
	Lat              float64       `bson:"lat" json:"lat"`
	Long             float64       `bson:"long" json:"long"`
	CreationDate     time.Time     `bson:"creation_date",json:"creation_date"`
}
type coordinates struct {
	Lat  float64 `json:"lat" bson:"lat"`
	Long float64 `json:"lon" bson:"lon"`
}

type dataTy struct {
	AvailableCategories []string         `json:"available_categories"`
	Bookmarks           []readBookmarkTy `json:"bookmarks"`
}

type attributesTy struct {
	imgSrcs     map[int]string
	title       string
	description string
	keywords    []string
	Mux         sync.Mutex
}

type channelData struct {
	Url         string
	Docselector bson.M
}

type categories struct {
	Url        string
	Categories []string
}
type importData struct {
	Url   string
	Icon  string
	Title string
}

var data []importData
var aTagCounter int
var attributes attributesTy
var usersCollection *mgo.Collection
var bookmarkCollection *mgo.Collection
var favIconsGridFs *mgo.GridFS

var tempImageGridFs *mgo.GridFS
var coordinatesChannel = make(chan coordinates, 2)
var dataChannel = make(chan channelData, 2)
var categoriesChannel = make(chan categories)
var service *visualrecognitionv3.VisualRecognitionV3
var serviceErr error

var orderCookieName string
var sessionCookieName string

func main() {

	byteArray, err := ioutil.ReadFile("apiKey")
	check(err)
	apiKey := string(byteArray)
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
	bookmarkCollection = db.C("bookmarks")
	favIconsGridFs = db.GridFS("favicons")
	tempImageGridFs = db.GridFS("temp")

	http.Handle("/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/drueckMich", pressMeHandler)
	http.HandleFunc("/Url", urlAjaxHandler)
	http.HandleFunc("/registrate", registrateHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/deleteAccount", deleteAccountHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/gridGetIcon/", getIconFromGrid)
	http.HandleFunc("/newCategory", newCategoryHandler)
	http.HandleFunc("/setSortProperties", sortPropertiesHandler)
	http.HandleFunc("/addCategoryToBookmark", addCategoryToBookmark)
	http.HandleFunc("/removeCategory", removeCategory)
	http.HandleFunc("/geospatial", geoSpatialhandler)
	http.HandleFunc("/upload", upload)

	err := http.ListenAndServe(":4242", nil)
	check(err)
}
func pressMeHandler(writer http.ResponseWriter, request *http.Request) {
	var Id string
	sessionCookie, _ := request.Cookie("pressMe")
	var user readUserTy
	orderCookie, _ := request.Cookie(orderCookieName)
	if request.Method == "POST" {
		//benutzer mnöchte sich anmelden
		userName := request.PostFormValue("username")
		password := request.PostFormValue("password")

		exits, _ := usersCollection.Find(bson.M{"username": userName, "password": password}).Count()
		//user exists?
		if exits == 1 {
			err := usersCollection.Find(bson.M{"username": userName, "password": password}).One(&user)
			if err != nil {
				fmt.Println(err)
			}

			if err != nil {
				log.Fatal(err)
			}
			Id = user.ID.Hex()

			setCookie := http.Cookie{
				Name:  sessionCookieName,
				Value: Id,
				Path:  "/",
			}
			http.SetCookie(writer, &setCookie)
			exits, err = bookmarkCollection.Find(bson.M{"user_id": user.ID}).Count()
			check(err)
			if exits >= 1 {

				err = t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries("", Id))
				if err != nil {
					log.Fatal(err)
				}
			} else {
				err = t.ExecuteTemplate(writer, "bookmarks", nil)
			}

		} else {
			var errMessage messageTy
			errMessage.Message = "Benutzer existiert nicht"
			err := t.ExecuteTemplate(writer, "login", errMessage)
			check(err)
		}
		//benutzer ist angemeldet
	} else if request.Method == "GET" {
		if sessionCookie != nil {

			exits, err := bookmarkCollection.Find(bson.M{"user_id": bson.ObjectIdHex(sessionCookie.Value)}).Count()
			check(err)
			if exits >= 1 {
				Id = sessionCookie.Value
				if orderCookie != nil {
					err = t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries(orderCookie.Value, Id))
					check(err)
				} else {
					err = t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries("", Id))
					check(err)
				}
			} else {
				err = t.ExecuteTemplate(writer, "bookmarks", nil)
				check(err)
			}

		} else {
			err := t.ExecuteTemplate(writer, "login", nil)
			check(err)
		}

	}

}

//benutzer möchte koordinaten filtern
func geoSpatialhandler(writer http.ResponseWriter, request *http.Request) {
	latitude := request.PostFormValue("latitude")
	longitude := request.PostFormValue("longitude")
	docs := geoSpatialQuery(latitude, longitude)
	_, err := fmt.Fprint(writer, docs)
	check(err)

}

func upload(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie(sessionCookieName)
	check(err)
	id := bson.ObjectIdHex(cookie.Value)
	reader, err := request.MultipartReader()

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// Jeden "part" in den Ordner ./upDownL kopieren:
	for {
		part, err := reader.NextPart()
		if err == io.EOF {

			break
		}

		// falls part.FileName() leer, überspringen:
		if part.FileName() == "" {
			continue
		}

		dst, err := os.Create("./files/" + part.FileName())
		defer dst.Close()
		check(err)
		//datei analysieren
		go analyzeImport(part.FileName(), id)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

	}

}

func addCategoryToBookmark(writer http.ResponseWriter, request *http.Request) {
	var jsonString string
	orderMethodCookie, _ := request.Cookie(orderCookieName)
	cookie, err := request.Cookie(sessionCookieName)
	check(err)
	id := cookie.Value
	url := request.PostFormValue("url")
	category := request.PostFormValue("category")
	docSelector := bson.M{"user_id": bson.ObjectIdHex(id), "url": url}
	docUpdate := bson.M{"$addToSet": bson.M{"custom_categories": category}}
	err = bookmarkCollection.Update(docSelector, docUpdate)
	check(err)
	//ist der Cookie vorhanden in dem gespeichert ist ob und wie sortiert werden soll?
	if orderMethodCookie != nil {

		bytestring, err := json.Marshal(getBookmarksEntries(orderMethodCookie.Value, id))
		if err != nil {
			fmt.Println(err)
		}
		jsonString = string(bytestring)
	} else {
		bytestring, err := json.Marshal(getBookmarksEntries("", id))
		if err != nil {
			fmt.Println(err)
		}
		jsonString = string(bytestring)
	}

	fmt.Fprint(writer, jsonString)

}
func removeCategory(writer http.ResponseWriter, request *http.Request) {
	var bookmark bookmarkTy
	var jsonString string
	orderMethodCookie, _ := request.Cookie(orderCookieName)
	cookie, err := request.Cookie(sessionCookieName)
	check(err)
	id := cookie.Value
	url := request.PostFormValue("url")
	category := request.PostFormValue("category")
	docSelector := bson.M{"user_id": bson.ObjectIdHex(id), "url": url}
	err = bookmarkCollection.Find(docSelector).One(&bookmark)
	check(err)

	if bookmark.URL == url {
		for j := len(bookmark.CustomCategories) - 1; j >= 0; j-- {
			if bookmark.CustomCategories[j] == category {
				bookmark.CustomCategories = append(bookmark.CustomCategories[:j],
					bookmark.CustomCategories[j+1:]...)

			}
		}
	}

	err = bookmarkCollection.Update(docSelector, bookmark)

	check(err)
	if orderMethodCookie != nil {
		fmt.Println(orderMethodCookie.Value)
		bytestring, err := json.Marshal(getBookmarksEntries(orderMethodCookie.Value, id))
		if err != nil {
			fmt.Println(err)
		}
		jsonString = string(bytestring)
	} else {
		bytestring, err := json.Marshal(getBookmarksEntries("", id))
		if err != nil {
			fmt.Println(err)
		}
		jsonString = string(bytestring)
	}

	_, err = fmt.Fprint(writer, jsonString)
	check(err)
}

/*
Benutzer möchte Lesezeichen sortieren
der DatenTyp von orderBy sind Integer
0 -> alphabetisch nach Titeln sortieren
1-> nach Erstellungsdatum sortieren
*/

func sortPropertiesHandler(writer http.ResponseWriter, request *http.Request) {

	var newCookie http.Cookie

	orderBy := request.PostFormValue("orderBy")
	cookie, err := request.Cookie(sessionCookieName)
	check(err)
	Id := cookie.Value

	newCookie = http.Cookie{
		Name:  orderCookieName,
		Value: "orderBy=" + orderBy,
		Path:  "/",
	}

	http.SetCookie(writer, &newCookie)
	var jsonString string

	bytestring, err := json.Marshal(getBookmarksEntries(newCookie.Value, Id))
	if err != nil {
		fmt.Println(err)
	}

	jsonString = string(bytestring)

	_, err = fmt.Fprint(writer, jsonString)
	check(err)
}

func getIconFromGrid(writer http.ResponseWriter, request *http.Request) {
	iconName := request.URL.Query().Get("fileName")

	if len(iconName) != 0 {
		img, err := favIconsGridFs.Open(iconName)
		//für den Fall dass kein FavIcon gefunden wurde
		if err != nil {
			http.Error(writer, err.Error(), http.StatusNotFound)
		} else {
			writer.Header().Add("Content-Type", img.ContentType())
			_, err = io.Copy(writer, img)

			err = img.Close()

		}
	}

}

func updateHandler(writer http.ResponseWriter, request *http.Request) {
	var jsonString string
	var Id string
	cookie, _ := request.Cookie("pressMe")
	orderMethodCookie, _ := request.Cookie(orderCookieName)
	//werte werden gesendet wenn vorhanden
	latitude := request.PostFormValue("latitude")
	longitude := request.PostFormValue("longitude")
	//prüfen ob werte vorhanden, wenn ja Umkreissuche ausführen
	if latitude != "" && longitude != "" {
		_, err := fmt.Fprint(writer, geoSpatialQuery(latitude, longitude))
		check(err)
		//keine Umkreisusche
	} else {
		Id = cookie.Value
		//überprüfen ob die Lesezeichen sortiert werden sollen
		if orderMethodCookie != nil {
			bytestring, err := json.Marshal(getBookmarksEntries(orderMethodCookie.Value, Id))
			if err != nil {
				fmt.Println(err)
			}
			jsonString = string(bytestring)
			//keine Sortierung
		} else {
			bytestring, err := json.Marshal(getBookmarksEntries("", Id))
			if err != nil {
				fmt.Println(err)
			}
			jsonString = string(bytestring)
		}

		_, err := fmt.Fprint(writer, jsonString)
		check(err)
	}

}

func newCategoryHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("newCategory")
	cookie, err := request.Cookie("pressMe")
	check(err)
	var Id string
	if cookie != nil {
		catName := request.PostFormValue("catName")
		Id = cookie.Value
		docSelector := bson.M{"_id": bson.ObjectIdHex(Id)}
		docUpdate := bson.M{"$addToSet": bson.M{"available_categories": catName}}
		err = usersCollection.Update(docSelector, docUpdate)
		check(err)
		_, err := fmt.Fprint(writer, "sucess")
		check(err)
	} else {

	}
}

//alle UserDaten bis auf die Bilder von den Favicosn werden gelöscht
func deleteAccountHandler(writer http.ResponseWriter, request *http.Request) {
	var message messageTy
	oldCookie, _ := request.Cookie("pressMe")
	docSelector := bson.M{"_id": bson.ObjectIdHex(oldCookie.Value)}
	err := usersCollection.Remove(docSelector)
	check(err)
	docSelector = bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value)}
	_, err = bookmarkCollection.RemoveAll(docSelector)
	message.Message = "Account gelöscht"
	err = t.ExecuteTemplate(writer, "login", message)
	check(err)
}
func logoutHandler(writer http.ResponseWriter, request *http.Request) {
	oldCookie, _ := request.Cookie("pressMe")

	//Cookie löschen
	newCookie := http.Cookie{
		Name:   sessionCookieName,
		MaxAge: -1,
	}
	http.SetCookie(writer, &newCookie)
	//zurück zu r login seite
	err := t.ExecuteTemplate(writer, "login", nil)
	check(err)
}
func urlAjaxHandler(_ http.ResponseWriter, request *http.Request) {
	var bookmark bookmarkTy
	var result readBookmarkTy
	Url := request.URL.Query().Get("url")
	oldCookie, _ := request.Cookie("pressMe")
	//wenn sessionCookie existiert Url verarbeiten
	//ansonsten verwerfen
	if oldCookie != nil {
		docSelector := bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value), "url": Url}
		exits, _ := bookmarkCollection.Find(docSelector).Count()
		//das Lesezeichen soll nur einmal angelegt werden
		//wennn es noch nicht existiert -> Lesezeichen anlegen
		if exits == 0 {
			bookmark.UserId = bson.ObjectIdHex(oldCookie.Value)
			bookmark.URL = Url
			bookmark.Location = GeoJsonTy{"Point", []float64{0, 0}}
			err := bookmarkCollection.Insert(bookmark)

			if err != nil {
				fmt.Print(err)
			}
			//für den weiteren Verarbeitungsprozess wird die Url benötigt
			//der DocSelector um die Ergebnisse in der DB abzuspeichern
			var channelData = channelData{
				Url:         Url,
				Docselector: docSelector,
			}

			go getAndProcessPage()
			dataChannel <- channelData

			//auslesen des angelegten Lesezeichens um das Estellungsdatum aus der Id zu erstellen
			//könnte auch mit Time.Now beim anlegen passieren
			err = bookmarkCollection.Find(docSelector).One(&result)
			docUpdate := bson.M{"$set": bson.M{"creation_date": result.ID.Time()}}
			err = bookmarkCollection.Update(docSelector, docUpdate)
			check(err)
		} else {
			//verwerfen
		}
	}

}
func registrateHandler(writer http.ResponseWriter, request *http.Request) {
	//ToDo check if Password is 'Stong' enough
	userName := request.PostFormValue("username")
	password := request.PostFormValue("password")
	userExists, _ := usersCollection.Find(bson.M{"username": userName}).Count()
	if userExists == 0 {
		userDoc := userTy{userName, password, nil}

		var errMessage messageTy
		//wenn wahr, wird das Modale fenster angezeigt
		errMessage.ShowModal = true
		errMessage.Message = "Benutzer erstellt"

		err := usersCollection.Insert(userDoc)
		check(err)
		err = t.ExecuteTemplate(writer, "login", errMessage)
		check(err)
	} else {
		var errMessage messageTy
		errMessage.Message = "Benutzer existiert schon"
		errMessage.ShowModal = true
		err := t.ExecuteTemplate(writer, "login", errMessage)
		check(err)

	}
}

//analyse der Website
func getAndProcessPage() {
	channelData := <-dataChannel
	docSelector := channelData.Docselector
	pageUrl := channelData.Url

	var imgUrls []*url.URL

	// HTTP-GET Request senden:
	//timeout auf 30 sekunden , da es zu timeouts kam
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Get(pageUrl)
	if err != nil {
		//ToDo errorhandling
	} else {
		byteArrayPage, err := ioutil.ReadAll(res.Body)
		check(err)
		err = res.Body.Close()
		check(err)
		if err != nil {
			fmt.Println(err)
		}
		// Empfangene Seite parsen, in doc-tree wandeln:
		docZeiger, err := html.Parse(strings.NewReader(string(byteArrayPage)))
		if err != nil {
			fmt.Println(err)
		}

		//Locken des globalen Struct types um zu verhindern, dass 2 oder mehrere Goroutine ihn benutzen und die Werte überschrieben werden.
		//kann nicht lokal in der Funktion getAllAttributes angelegt werden, da sie sich rekursiv aufruft und sonst die werte überschrieben werden
		attributes.Mux.Lock()
		//alle Werte aufd Ursprung setzen
		attributes.description, attributes.title, attributes.keywords = "", "", []string{}
		attributes.imgSrcs = make(map[int]string)
		attributes = getAllAttributes(docZeiger)

		u, err := url.Parse(pageUrl)
		if err != nil {
			fmt.Println(err)
		}

		for _, wert := range attributes.imgSrcs {
			absURL, err := u.Parse(wert)
			if err != nil {
				fmt.Println(err)
			}
			//todo check if this is necessary
			imgUrls = append(imgUrls, absURL)

			docUpdate := bson.M{"$addToSet": bson.M{"images": absURL.String()}}
			err = bookmarkCollection.Update(docSelector, docUpdate)
			check(err)

		}
		docUpdate := bson.M{"$set": bson.M{"keywords": attributes.keywords, "title": attributes.title, "icon": attributes.title, "shortReview": attributes.description}}
		err = bookmarkCollection.Update(docSelector, docUpdate)

		//-----------------------------------------------------------------------------------------------------------------------------------------------------------------------
		//extrahieren des FavIcons und abspeichern im GridFs

		faviconUrl := "https://www.google.com/s2/favicons?domain=" + pageUrl
		res, err := http.Get(faviconUrl)
		if err != nil {
			fmt.Println(err)

		} else {
			mimeType := res.Header.Get("Content-Type")
			//name der datei ist der title der Webseite
			gridFile, err := favIconsGridFs.Create(attributes.title)
			attributes.Mux.Unlock()
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
		//aufruf einer weiteren goroutine um eine eventuelle Position aus den Metadaten auszulesen
		go extractPosition(imgUrls)
		coordinates := <-coordinatesChannel

		geojson := GeoJsonTy{"Point", []float64{coordinates.Long, coordinates.Lat}}
		docUpdate = bson.M{"$set": bson.M{"location": geojson, "lat": coordinates.Lat, "long": coordinates.Long}}
		err = bookmarkCollection.Update(docSelector, docUpdate)
		check(err)
		// ...createIndex( { location: "2dsphere" } ) existiert nicht im golang API, daher mit EnsureIndex:
		index := mgo.Index{
			Key: []string{"$2dsphere:location"},
			//		Bits: 26,
		}
		err = bookmarkCollection.EnsureIndex(index)
		if err != nil {
			panic(err)
		}
		//watson kategorien finden
		go classesRecoginition(imgUrls)
		categories := <-categoriesChannel

		docUpdate = bson.M{"$addToSet": bson.M{"wvr_categories": bson.M{"$each": categories.Categories}}}

		err = bookmarkCollection.Update(docSelector, docUpdate)
		check(err)
	}

}
func getAllAttributes(node *html.Node) attributesTy {
	if node.Type == html.ElementNode {
		switch node.Data {
		case "img":

			for _, img := range node.Attr {
				if img.Key == "src" {
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

			for i, meta := range node.Attr {

				switch meta.Val {

				case "description":
					//das nächste Attribut Element ist der Text der Beschreibung
					//Todo is this working with i+1?
					attributes.description = node.Attr[i+1].Val
					break

				case "keywords":
					//im nächsten Attribut element sind die keywords angeben
					//Todo is this working with i+1?
					keywords := strings.Split(node.Attr[i+1].Val, " ")
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
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func extractPosition(urls []*url.URL) {
	var coordinates = coordinates{}

	for i, url := range urls {
		res, err := http.Get(url.String())
		if err != nil {
			fmt.Println(err.Error())
		}

		x, err := exif.Decode(res.Body)
		if err != nil {

			if (len(urls) - 1) == i {

				coordinatesChannel <- coordinates
			}
		} else {

			latitude, longitude, _ := x.LatLong()
			fmt.Println(latitude, longitude)
			coordinates.Lat = latitude
			coordinates.Long = longitude

			coordinatesChannel <- coordinates
			//wen koordianten gefunden wurden for-Scheleife abbrechen
			return
		}

	}
}

func getBookmarksEntries(orderMethod string, Id string) dataTy {
	var docs []readBookmarkTy
	var user userTy

	docSelector := bson.M{"user_id": bson.ObjectIdHex(Id)}

	err := usersCollection.Find(bson.M{"_id": bson.ObjectIdHex(Id)}).One(&user)
	check(err)
	if len(orderMethod) > 0 {
		parts := strings.Split(orderMethod, "=")

		if parts[1] == "0" {
			err = bookmarkCollection.Find(docSelector).Sort("title").All(&docs)

		} else if parts[1] == "1" {

			err = bookmarkCollection.Find(docSelector).Sort("creation_date").All(&docs)

		}

	} else {
		err := bookmarkCollection.Find(docSelector).All(&docs)
		check(err)
	}

	return dataTy{AvailableCategories: user.AvailableCategories, Bookmarks: docs}
}
func classesRecoginition(urls []*url.URL) {
	var categories = categories{}
	for _, url := range urls {
		res, _ := http.Get(url.String())
		fmt.Println(res.Header)
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
			fmt.Println(responseErr)

		} else {
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
					fmt.Println("urlOfImageCategoryFound:", url.String())
					categories.Url = url.String()
				}
			}
			categoriesChannel <- categories
			//wenn Kategorien für ein bIld gefunden abbrecvhen
			return
		}

	}

}
func analyzeImport(fileName string, id bson.ObjectId) {

	fmt.Println("analyzeImport")

	filePath := path.Join("./files", fileName)

	tempFile, err := os.Open(filePath)
	defer tempFile.Close()
	check(err)
	byteArrayPage, err := ioutil.ReadAll(tempFile)
	check(err)
	docZeiger, err := html.Parse(strings.NewReader(string(byteArrayPage)))
	if err != nil {
		fmt.Println(err)
	}
	aTagCounter = 0
	data = getUrl(docZeiger)
	for i := range data {
		var bookmark bookmarkTy
		bookmark.UserId = id
		bookmark.Location = GeoJsonTy{"Point", []float64{0, 0}}
		bookmark.URL = data[i].Url
		bookmark.Title = data[i].Title
		err = bookmarkCollection.Insert(bookmark)
		check(err)
		docSelector := bson.M{"user_id": id, "url": data[i].Url}
		var channelData = channelData{
			Url:         data[i].Url,
			Docselector: docSelector,
		}
		go getAndProcessPage()
		dataChannel <- channelData

	}

}

func getUrl(node *html.Node) []importData {

	if node.Type == html.ElementNode {

		switch node.Data {

		case "a":
			fmt.Println("a")
			data = append(data, importData{})

			data[aTagCounter].Title = node.LastChild.Data
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					data[aTagCounter].Url = attr.Val
				} else if attr.Key == "icon" {

					data[aTagCounter].Icon = attr.Val
				}
			}
			aTagCounter++

		}

	} else if node.Type == html.TextNode {

	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {

		getUrl(child)
	}
	return data
}

//wird eventuell von der updateMethode oder dem geospatiaHandler aufgerufen
func geoSpatialQuery(latitude string, longitude string) string {
	var docs []readBookmarkTy

	long64, _ := strconv.ParseFloat(longitude, 64)
	lat64, _ := strconv.ParseFloat(latitude, 64)

	fmt.Println("latitude:", latitude, " longitude:", longitude)
	err := bookmarkCollection.Find(bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{long64, lat64},
				},
				"$minDistance": 0,
				"$maxDistance": 7000,
			},
		},
	}).All(&docs)
	if err != nil {
		panic(err)
	}

	data := dataTy{nil, docs}
	bytestring, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	return string(bytestring)

}
