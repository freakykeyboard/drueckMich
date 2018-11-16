package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

//get working Directory

var t = template.Must(template.ParseFiles(filepath.Join("./", "template", "head.html"),
	filepath.Join("./", "template", "login.html"), filepath.Join("./", "template", "bookmarks.html"),
	filepath.Join("./", "template", "end.html")))

func main() {

	http.HandleFunc("/drückMich", pressMeHandler)
	http.HandleFunc("/Url", urlAjaxHandler)
	http.ListenAndServe(":4242", nil)
}
func urlAjaxHandler(writer http.ResponseWriter, request *http.Request) {
	url := request.URL.Query().Get("url")
	fmt.Println(url)

}
func pressMeHandler(writer http.ResponseWriter, request *http.Request) {

}
