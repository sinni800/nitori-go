package main

import (
	"crypto"
	"crypto/rand"
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/url"
	//"net/http/httputil"
	//"strconv"
	"os"
	//"bufio"
	"flag"
	"math/big"
	//"time"
	"path/filepath"
	//"strings"
	"io/ioutil"
)

var fHttplisten *string = flag.String("httplisten", "0.0.0.0:1339", "HTTP Listen address")
var fHttp *bool = flag.Bool("http", false, "Listen on http?")
var fHttpslisten *string = flag.String("httpslisten", "0.0.0.0:1340", "HTTPS Listen address")
var fHttps *bool = flag.Bool("https", false, "Listen on https?")
var fHttpsCertFile *string = flag.String("httpscert", "sslcert.pem", "Https certificate PEM file")
var fHttpsKeyFile *string = flag.String("httpskey", "sslprivkey.pem", "Https private key PEM file")
var fHttphostname *string = flag.String("httphostname", "natori.com", "HTTP hostname (used for some URLs)")
var Httplisten string
var Http bool
var Httpslisten string
var Https bool
var HttpsCertFile string
var HttpsKeyFile string
var Httphostname string

func init() {
	flag.Parse()
	Httplisten = *fHttplisten
	Http = *fHttp
	Httpslisten = *fHttpslisten
	Https = *fHttps
	HttpsCertFile = *fHttpsCertFile
	HttpsKeyFile = *fHttpsKeyFile
	Httphostname = *fHttphostname
}

func HttpHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func handleHttpFuncs() {

	HttpHandleFunc("/", func(w http.ResponseWriter, rew *http.Request) {
		Template("").Execute(w, nil)
	})

	HttpHandleFunc("/kawaii", Kawaiiface)
	HttpHandleFunc("/dna", DNA)
	HttpHandleFunc("/joke", Joke)
	HttpHandleFunc("/rape", Rape)
	HttpHandleFunc("/kill", Kill)
	HttpHandleFunc("/factoid", Factoid)
	HttpHandleFunc("/yomama", Yomama)
	HttpHandleFunc("/Gelbooru/Random", GelRandom)

	HttpHandleFunc("/FileUpload", GridFSFile)
	HttpHandleFunc("/FileDelete", GridFSDelete)
	HttpHandleFunc("/Files/", GridFSDL)
	HttpHandleFunc("/Files", GridFSDL)
	HttpHandleFunc("/Auth", Auth)
	HttpHandleFunc("/CreateUser", CreateUser)

	HttpHandleFunc("/plugins", PluginList)
	HttpHandleFunc("/plugins/edit", PluginEdit)
	HttpHandleFunc("/plugins/load", PluginLoad)
	HttpHandleFunc("/plugins/unload", PluginUnload)
	HttpHandleFunc("/plugins/delete", PluginDelete)

	u, _ := url.Parse("http://metalgearsonic.de/scripts/ace/") //I just host it here :)
	http.Handle("/scripts/ace/", http.StripPrefix("/scripts/ace/", NewSingleHostReverseProxy(u)))

}

func PluginList(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}
	plugins := make(map[string]map[string]interface{})
	filepath.Walk("plugins", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			_, filename := filepath.Split(path)
			plugins[filename] = make(map[string]interface{})
			plugins[filename]["file"] = true
			plugins[filename]["name"] = filename

		}
		return nil
	})

	for key, value := range pluginfuncs {
		if _, ok := plugins[key]; !ok {
			plugins[key] = make(map[string]interface{})
		}

		plugins[key]["loaded"] = true
		funcs := make([]string, 0, len(value))
		for key1, _ := range value {
			funcs = append(funcs, key1)
		}
		plugins[key]["functions"] = funcs
		plugins[key]["name"] = key
	}
	Template(TemplatePluginlist).Execute(w, plugins)
}

func PluginEdit(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}
	rew.ParseForm()
	name := rew.Form.Get("name")

	if rew.Method == "GET" {
		content, err := ioutil.ReadFile("plugins\\" + name)
		if err != nil {
			content = []byte("")
		}
		Template(TemplatePluginedit).Execute(w, map[string]string{"name": name, "content": string(content)})
	} else if rew.Method == "POST" {
		content := rew.FormValue("content")
		filename := rew.FormValue("filename")

		if filename != name {
			err := os.Remove("plugins\\" + name)
			if err != nil {
				fmt.Println("Could not delete old file:", name, filename, err.Error())
			}
			name = filename
		}

		if content != "" {
			ioutil.WriteFile("plugins\\"+name, []byte(content), 0)
		}
		w.Header().Add("Location", "/plugins")
		w.WriteHeader(303)
	}

}

func PluginDelete(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}
	rew.ParseForm()
	name := rew.FormValue("name")
	err := os.Remove("plugins\\" + name)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Location", "/plugins")
		w.WriteHeader(303)
	}
}

func PluginLoad(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}
	rew.ParseForm()
	name := rew.FormValue("name")
	err := loadpluginname(name)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Location", "/plugins")
		w.WriteHeader(303)
	}
}

func PluginUnload(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}
	rew.ParseForm()
	name := rew.FormValue("name")
	err := unloadpluginname(name)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(500)
	} else {
		w.Header().Add("Location", "/plugins")
		w.WriteHeader(303)
	}
}

func CreateUser(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}
	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	if rew.Method == "POST" {
		var Users *mgo.Collection = MongoDB.C("users")
		md5 := crypto.MD5.New()
		md5.Write([]byte(rew.FormValue("password")))
		passwordhash := fmt.Sprintf("%x", md5.Sum(nil))
		err := Users.Insert(bson.M{"username": rew.FormValue("username"), "password": passwordhash})
		if err == nil {
			io.WriteString(w, "Success.")
		} else {
			io.WriteString(w, err.Error())
		}
	} else if rew.Method == "GET" {
		Template(TemplateCreateUser).Execute(w, nil)
	}
}

func HttpAuthenticate(resp http.ResponseWriter, rew *http.Request) bool {
	if !CheckAuthCookie(rew) {
		resp.Header().Add("Location", "/Auth?return="+rew.RequestURI)
		resp.WriteHeader(303)
		return false
	} else {
		return true
	}
}

func CheckAuthCookie(rew *http.Request) bool {
	namecookie, err := rew.Cookie("loginuser")
	if err != nil {
		return false
	}

	pwdcookie, err := rew.Cookie("loginpwd")
	if err != nil {
		return false
	}

	return AuthenticateHashedPW(namecookie.Value, pwdcookie.Value)
}

func Auth(w http.ResponseWriter, rew *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	rew.ParseForm()
	if rew.Method == "POST" {
		md5 := crypto.MD5.New()
		md5.Write([]byte(rew.FormValue("password")))
		passwordhash := fmt.Sprintf("%x", md5.Sum(nil))

		if Authenticate(rew.FormValue("username"), rew.FormValue("password")) {
			w.Header().Add("Set-Cookie", "loginuser="+rew.FormValue("username")+"")
			w.Header().Add("Set-Cookie", "loginpwd="+passwordhash+"")
			if rew.Form.Get("return") != "" {
				w.Header().Add("Location", rew.Form.Get("return"))
			} else {
				w.Header().Add("Location", "/")
			}
			w.WriteHeader(303)
		} else {
			Template(TemplateLogin).Execute(w, "User or Password wrong")
		}
	} else if rew.Method == "GET" {
		Template(TemplateLogin).Execute(w, nil)
	}

}

func GridFSDL(w http.ResponseWriter, rew *http.Request) {
	auth := CheckAuthCookie(rew)

	if len(rew.URL.Path) < 7 {
		w.Header().Add("Content-Type", "text/html; charset=UTF-8")
		io.WriteString(w, "<!DOCTYPE html>\r\n"+
			"<html><head></head><body>")

		coll := MongoDB.C("fs.files")

		var result bson.M

		iter := coll.Find(nil).Iter()

		for {
			success := iter.Next(&result)

			if success {
				io.WriteString(w, "<a href=\"/Files/"+result["filename"].(string)+"\">"+result["filename"].(string)+"</a> ")
				if auth {
					io.WriteString(w, `<a href="/FileDelete?file=`+result["filename"].(string)+`">Delete</a> <br />`)
				}
			} else {
				break
			}
		}

		io.WriteString(w, "</body></html>")
	} else {
		if rew.URL.Path[7:] != "" {
			var GridFS *mgo.GridFS = MongoDB.GridFS("fs")
			GridFile, err := GridFS.Open(rew.URL.Path[7:])
			if err == nil {

				http.ServeContent(w, rew, GridFile.Name(), GridFile.UploadDate(), GridFile)

			} else {
				http.NotFound(w, rew)
			}
		}
	}
}

func GridFSDelete(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}

	rew.ParseForm()
	if rew.Form.Get("file") != "" {
		GridFS := MongoDB.GridFS("fs")
		err := GridFS.Remove(rew.Form.Get("file"))
		if err != nil {
			io.WriteString(w, "could not delete: "+err.Error())
		} else {
			io.WriteString(w, "File deleted.")
		}
	}
}

func GridFSFile(w http.ResponseWriter, rew *http.Request) {
	if !HttpAuthenticate(w, rew) {
		return
	}
	if rew.Method == "POST" {
		fmt.Println("METHOD WAS POST")
		GridFS := MongoDB.GridFS("fs")
		var file *mgo.GridFile
		var filename string

		rew.ParseMultipartForm(500000)
		formfileheaderarr := rew.MultipartForm.File["file"]
		formfileheader := formfileheaderarr[0]
		formfile, err := formfileheader.Open()

		//formfile, formfileheader, err := rew.FormFile("file")
		if err == nil {
			if rew.FormValue("filename") == "" {
				filename = formfileheader.Filename
			} else {
				filename = rew.FormValue("filename")
			}
			fmt.Println(filename)

			file, err = GridFS.Create(filename)

			if err == nil {

				_, err = io.Copy(file, formfile)
				if err == nil {
					file.SetContentType(formfileheader.Header.Get("Content-Type"))
					err = file.Close()
					if err == nil {
						w.Header().Add("Content-Type", "text/html; charset=UTF-8")
						io.WriteString(w, "<!DOCTYPE html>\r\n"+
							"<html><head></head><body>File uploaded, get here: <a href=\"/Files/"+filename+"\">"+filename+"</a></body></html>")
					}
				}
			}

		}

		if err != nil {
			io.WriteString(w, "Error occured: "+err.Error())
		}

		//GridFS.Create(
		//io.Copy()
		//rew.FormFile("file").
	} else if rew.Method == "GET" {
		w.Header().Add("Content-Type", "text/html; charset=UTF-8")
		io.WriteString(w, "<!doctype html>\r\n"+
			"<html><head></head><body><form enctype=\"multipart/form-data\" action=\"/FileUpload\" method=\"POST\">File: <input type=\"file\" name=\"file\" /><br />Filename: <input type=\"text\" name=\"filename\" /><br /><input type=\"submit\" name=\"submit\" value=\"Submit\" /></form></body></html>")
	}

}

func Joke(w http.ResponseWriter, rew *http.Request) {
	coll := MongoDB.C("joke")
	var result []bson.M
	coll.Find(nil).All(&result)
	Template(TemplateSingleColumnTable).Execute(w, result)
}

func Rape(w http.ResponseWriter, rew *http.Request) {
	coll := MongoDB.C("rape")
	var result []bson.M
	coll.Find(nil).All(&result)
	Template(TemplateSingleColumnTable).Execute(w, result)
}

func Kill(w http.ResponseWriter, rew *http.Request) {
	coll := MongoDB.C("kill")
	var result []bson.M
	coll.Find(nil).All(&result)
	Template(TemplateSingleColumnTable).Execute(w, result)
}

func Factoid(w http.ResponseWriter, rew *http.Request) {
	coll := MongoDB.C("factoid")
	var result []bson.M
	coll.Find(nil).All(&result)
	Template(TemplateSingleColumnTable).Execute(w, result)
}

func Yomama(w http.ResponseWriter, rew *http.Request) {
	coll := MongoDB.C("yomama")
	var result []bson.M
	coll.Find(nil).All(&result)
	Template(TemplateSingleColumnTable).Execute(w, result)
}

func DNA(w http.ResponseWriter, rew *http.Request) {
	coll := MongoDB.C("dna")
	var result []bson.M
	coll.Find(nil).All(&result)
	Template(TemplateDNATable).Execute(w, result)

}

func Kawaiiface(w http.ResponseWriter, rew *http.Request) {
	coll := MongoDB.C("kawaiiface")
	var result []bson.M
	coll.Find(nil).All(&result)
	Template(TemplateKawaiiTable).Execute(w, result)
}

func GelRandom(w http.ResponseWriter, rew *http.Request) {
	entries, _ := GelbooruGet("")
	if entries != nil {

		postnum, _ := rand.Int(rand.Reader, big.NewInt(int64(len(entries))))
		io.WriteString(w, "<a href=\""+entries[int(postnum.Int64())].Url+"\"><img src=\""+entries[int(postnum.Int64())].File_url+"\" /></a>")
	} else {
		io.WriteString(w, "Error")
	}
}
