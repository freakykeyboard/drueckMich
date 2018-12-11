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
	"sort"
	"strconv"
	"strings"
	"time"
)

var t = template.Must(template.ParseFiles(filepath.Join("./", "tpl", "head.html"),
	filepath.Join("./", "tpl", "login.html"), filepath.Join("./", "tpl", "bookmarks.html"),
	filepath.Join("./", "tpl", "end.html"), filepath.Join("./", "tpl", "registrate.html"),
	filepath.Join("./", "tpl", "newCategoryModal.html"), filepath.Join("./", "tpl", "removeCategoryModal.html"),
	filepath.Join("./", "tpl", "registrationModal.html"), filepath.Join("./", "tpl", "addCategoryModal.html")))

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
	Message string
}
type bookmarksTy struct {
	Bookmarks []bookmarkTy
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
var bookmarks bookmarksTy
var usersCollection *mgo.Collection
var bookmarkCollection *mgo.Collection
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
	http.HandleFunc("/gridGetIcon/", getIconFromGrid)
	http.HandleFunc("/newCategory", newCategoryHandler)
	http.HandleFunc("/setSortProperties", sortPropertiesHandler)
	http.HandleFunc("/addCategoryToBookmark", addCategoryToBookmark)
	http.HandleFunc("/removeCategory", removeCategory)
	http.HandleFunc("/upload", analyzeImportedBookmarks)
	http.HandleFunc("/geospatial", geoSpatialhandler)
	http.ListenAndServe(":4242", nil)
}

func geoSpatialhandler(writer http.ResponseWriter, request *http.Request) {
	latitude := request.PostFormValue("latitude")
	longitude := request.PostFormValue("longitude")
	docs := geoSpatialQuery(latitude, longitude)
	fmt.Fprint(writer, docs)

}
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

func analyzeImportedBookmarks(writer http.ResponseWriter, request *http.Request) {
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

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		//todo is too early file cannot be found ->error

	}

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
	var jsonString string

	bytestring, err := json.Marshal(getBookmarksEntries(newCookie.Value))
	if err != nil {
		fmt.Println(err)
	}

	jsonString = string(bytestring)

	fmt.Fprint(writer, jsonString)
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
	latitude := request.PostFormValue("latitude")
	longitude := request.PostFormValue("longitude")
	if latitude != "" && longitude != "" {
		fmt.Fprint(writer, geoSpatialQuery(latitude, longitude))

	} else {
		Id = cookie.Value
		if orderMethodCookie != nil {

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

}
func analyzeImport(fileName string) {
	file, err := os.Open(fileName)
	check(err)
	//byteArrayPage, err := ioutil.ReadAll(file)
	defer file.Close()
	//docZeiger, err := html.Parse(strings.NewReader(string(byteArrayPage)))
	if err != nil {
		fmt.Println(err)
	}

}
func getAndProcessPage() {
	channelData := <-dataChannel
	docSelector := channelData.Docselector
	pageUrl := channelData.Url
	fmt.Println("getAndProcessPage")

	var imgUrls []*url.URL

	// HTTP-GET Request senden:
	//because of timeouts
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Get(pageUrl)
	if err != nil {
		fmt.Println("zeile 341")
		fmt.Println("err.Error()", err.Error())

	} else {
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

			docUpdate := bson.M{"$addToSet": bson.M{"images": absURL.String()},
				"$set": bson.M{"keywords": attributes.keywords, "title": attributes.title, "icon": attributes.title, "shortReview": attributes.description}}
			err = bookmarkCollection.Update(docSelector, docUpdate)
			check(err)

		}
		go extractPosition(imgUrls)

		coordinates := <-coordinatesChannel
		geojson := GeoJsonTy{"Point", []float64{coordinates.Long, coordinates.Lat}}
		docUpdate := bson.M{"$set": bson.M{"location": geojson, "lat": coordinates.Lat, "long": coordinates.Long}}
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
	err := usersCollection.Remove(docSelector)
	check(err)
	docSelector = bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value)}
	_, err = bookmarkCollection.RemoveAll(docSelector)
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

	userName := request.PostFormValue("username")
	password := request.PostFormValue("password")
	userExists, _ := usersCollection.Find(bson.M{"username": userName}).Count()
	if userExists == 0 {
		userDoc := userTy{userName, password, nil}

		var errMessage messageTy
		errMessage.Message = "Benutzer erstellt"

		err := usersCollection.Insert(userDoc)
		check(err)
		err = t.ExecuteTemplate(writer, "registrationModal", errMessage)
		check(err)

	} else {
		var errMessage messageTy
		errMessage.Message = "Benutzer existiert schon"

		err := t.ExecuteTemplate(writer, "registrationModal", errMessage)
		check(err)
	}
}

func urlAjaxHandler(r http.ResponseWriter, request *http.Request) {
	//ToDo check if entry alread exists
	fmt.Println("urlAjaxHandler")
	var bookmark bookmarkTy
	var result readBookmarkTy
	Url := request.URL.Query().Get("url")
	oldCookie, _ := request.Cookie("pressMe")
	if oldCookie != nil {
		docSelector := bson.M{"user_id": bson.ObjectIdHex(oldCookie.Value), "url": Url}
		var channelData = channelData{
			Url:         Url,
			Docselector: docSelector,
		}
		go getAndProcessPage()
		dataChannel <- channelData
		bookmark.UserId = bson.ObjectIdHex(oldCookie.Value)
		bookmark.URL = Url
		bookmark.Location = GeoJsonTy{"Point", []float64{0, 0}}
		err := bookmarkCollection.Insert(bookmark)

		if err != nil {
			fmt.Print(err)
		}
		err = bookmarkCollection.Find(docSelector).One(&result)
		docUpdate := bson.M{"$set": bson.M{"creation_date": result.ID.Time()}}
		err = bookmarkCollection.Update(docSelector, docUpdate)
		check(err)
	} else {
		//do Nothing
	}

}

func extractPosition(urls []*url.URL) {

	fmt.Println("extractPÜosition")
	var coordinates = coordinates{}

	for i, url := range urls {
		fmt.Println("url.String()", url.String())
		res, err := http.Get(url.String())
		if err != nil {
			fmt.Println(err.Error())
		}

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
			fmt.Println(latitude, longitude)
			coordinates.Lat = latitude
			coordinates.Long = longitude
			fmt.Println("coordinates:", coordinates)
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

	} else {
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

}
func pressMeHandler(writer http.ResponseWriter, request *http.Request) {
	//ToDo change example
	sessionCookie, _ := request.Cookie("pressMe")
	var users []readUserTy
	orderCookie, _ := request.Cookie(orderCookieName)
	if request.Method == "POST" {

		userName := request.PostFormValue("username")
		password := request.PostFormValue("password")

		exits, _ := usersCollection.Find(bson.M{"username": userName, "password": password}).Count()
		//user exists?
		if exits == 1 {
			//toDo use One
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
			exits, err = bookmarkCollection.Find(bson.M{"user_id": users[0].ID}).Count()
			check(err)
			if exits >= 1 {

				err = t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries(""))
				if err != nil {
					log.Fatal(err)
				}
			} else {
				err = t.ExecuteTemplate(writer, "bookmarks", nil)
			}

		} else {
			var errMessage messageTy
			errMessage.Message = "Benutzer existiert nicht"
			t.ExecuteTemplate(writer, "login", errMessage)
		}
	} else if request.Method == "GET" {
		if sessionCookie != nil {

			exits, err := bookmarkCollection.Find(bson.M{"user_id": bson.ObjectIdHex(sessionCookie.Value)}).Count()
			check(err)
			if exits >= 1 {
				Id = sessionCookie.Value
				if orderCookie != nil {
					t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries(orderCookie.Value))
				} else {
					t.ExecuteTemplate(writer, "bookmarks", getBookmarksEntries(""))
				}
			} else {
				t.ExecuteTemplate(writer, "bookmarks", nil)
			}

		} else {
			t.ExecuteTemplate(writer, "login", nil)
		}

	}

}
func getBookmarksEntries(orderMethod string) dataTy {
	var docs []readBookmarkTy
	var user userTy

	docSelector := bson.M{"user_id": bson.ObjectIdHex(Id)}

	err := bookmarkCollection.Find(docSelector).All(&docs)
	check(err)
	err = usersCollection.Find(bson.M{"_id": bson.ObjectIdHex(Id)}).One(&user)
	check(err)
	if len(orderMethod) > 0 {
		parts := strings.Split(orderMethod, "=")

		if parts[1] == "0" {
			sort.Slice(docs, func(i, j int) bool {
				return docs[i].Title < docs[j].Title
			})

		} else if parts[1] == "1" {
			sort.Slice(docs, func(i, j int) bool {
				return docs[i].CreationDate.Before(docs[j].CreationDate)
			})
		}

	}

	return dataTy{AvailableCategories: user.AvailableCategories, Bookmarks: docs}
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
				}
			}
			categoriesChannel <- categories
			return
		}

	}

}
