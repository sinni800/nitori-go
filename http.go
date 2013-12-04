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
	"sort"
	//"bufio"
	"math/big"
	//"time"
	//"strings"

)

func handleHttpFuncs() {

	HttpHandleFunc("/", func(w http.ResponseWriter, rew *http.Request) {
		err := Template("").Execute(w, nil)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	}, false)

	HttpHandleFunc("/kawaii", Kawaiiface, false)
	HttpHandleFunc("/dna", DNA, false)
	HttpHandleFunc("/show1", Showsinglecolumn, false)
	HttpHandleFunc("/showtable", ShowTable, false)
	HttpHandleFunc("/Gelbooru/Random", GelRandom, false)

	HttpHandleFunc("/Auth", Auth, false)
	HttpHandleFunc("/Logout", Logout, false)
	HttpHandleFunc("/CreateUser", CreateUser, true)

	HttpHandleFunc("/docs", Documentation, false)

	HttpHandleFunc("/IRC", HttpIRC, false)
	HttpHandleFunc("/IRC/Log", HttpIRCLog, false)

	FileHttp()
	PluginHttp()

	u, _ := url.Parse("http://metalgearsonic.de/scripts/ace/") //I just host it here :)
	http.Handle("/scripts/ace/", http.StripPrefix("/scripts/ace/", NewSingleHostReverseProxy(u)))

}

func HttpHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), auth bool) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, rew *http.Request) {
		if auth && HttpAuthenticate(w, rew) {
			handler(w, rew)
		} else if !auth {
			handler(w, rew)
		} else {
			return
		}
	})
}

func HttpHandle(pattern string, handler http.Handler, auth bool) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, rew *http.Request) {
		if auth && HttpAuthenticate(w, rew) {
			handler.ServeHTTP(w, rew)
		} else if !auth {
			handler.ServeHTTP(w, rew)
		} else {
			return
		}
	})
}

func HttpIRC(w http.ResponseWriter, rew *http.Request) {
	Template(TemplateIRCLogs).Execute(w, conf.Instances)
}

func HttpIRCLog(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()

	inst := rew.FormValue("inst")
	channel := rew.FormValue("chan")

	if i, ok := conf.Instances[inst]; ok {
		if log, ok2 := i.Irclog[channel]; ok2 {
			out := make(map[string]interface{})
			out["log"] = log.Out()
			out["chan"] = channel
			out["inst"] = i
			Template(TemplateIRCLatest).Execute(w, out)
			return
		}
	}

	http.NotFound(w, rew)
}

func Documentation(w http.ResponseWriter, rew *http.Request) {
	if val, ok := conf.Instances[conf.DocInstance]; ok {
		Template(TemplateDocs).Execute(w, val.Plugindocs)
	} else {
		w.WriteHeader(500)
		w.Write([]byte("No doc instance defined or doc instance does not exist."))
	}
}

func CreateUser(w http.ResponseWriter, rew *http.Request) {
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

func Logout(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Set-Cookie", "loginuser=; expires=Thu, 01 Jan 1970 00:00:00 GMT")
	w.Header().Add("Set-Cookie", "loginpwd=; expires=Thu, 01 Jan 1970 00:00:00 GMT")
	w.Header().Add("Location", "/")
	w.WriteHeader(303)
}

func Showsinglecolumn(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if req.Form.Get("c") != "" {
		coll := MongoDB.C(req.Form.Get("c"))
		var result []bson.M
		coll.Find(nil).All(&result)
		Template(TemplateSingleColumnTable).Execute(w, result)
	} else {
		http.NotFound(w, req)
	}
}

func ShowTable(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if req.Form.Get("c") != "" {
		coll := MongoDB.C(req.Form.Get("c"))
		var stuff []bson.M
		coll.Find(nil).All(&stuff)
		cols1 := make(map[string]bool)
		
		for _, doc := range stuff {
			for key, _ := range doc {
				cols1[key] = true
			}
		}
		
		cols := make([]string, 0, len(cols1))
		
		for key, _ := range cols1 {
			cols = append(cols, key)
		}
		
		sort.Strings(cols)
		
		result := map[string]interface{} {"Data": stuff, "Columns": cols}
		Template(TemplateMongoTable).Execute(w, result)
	} else {
		http.NotFound(w, req)
	}
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
