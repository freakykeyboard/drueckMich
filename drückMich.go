package main

import (
	"fmt"
	"net/http"

)

func main() {
	http.HandleFunc("/drückMich",pressMeHandler)
	http.HandleFunc("/Url",urlAjaxHandler)
	http.ListenAndServe(":4242",nil)
}
func urlAjaxHandler(writer http.ResponseWriter, request *http.Request) {
	url:=request.URL.Query().Get("url")
	fmt.Println(url)



}
func pressMeHandler(writer http.ResponseWriter, request *http.Request) {
	
}
