// http_dropbox.go
package main

import (
	"net/http"
)

func DropboxHttp() {
	HttpHandleFunc("/Dropbox/", nil, true)
	http.Handle("/Dropbox", http.RedirectHandler("/Dropbox/", 301))
}

//func Dummy(w http.ResponseWriter, rew *http.Request) {}

func Dropbox(w http.ResponseWriter, rew *http.Request) {

}
