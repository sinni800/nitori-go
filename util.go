package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httputil"
	//"time"
	"net/url"
	"path/filepath"
	"strings"
)

func Args(s string) (result []string) {

	defer func() {
		if recover() != nil {
			result = make([]string, 0, 0)
		}
	}()

	inStr := false
	escape := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if escape {
			escape = false
			return false
		}
		switch r {
		case '\\':
			escape = true
			return false
		case ' ', '\n', '\t':
			return !inStr
		case '"':
			inStr = !inStr
			return true
		default:
			return false
		}
	})
}

func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

type Ftypes struct {
	WebAudio bool
	Audio    bool
	Video    bool
	WebVideo bool
	Flash    bool
	Image    bool
	DirectlyViewable bool
}

func Filetypes(filename string) Ftypes {
	ext := filepath.Ext(filename)
	ret := Ftypes{}
	e := strings.ToLower(ext)
	ret.Audio = StringIsOneOf(e, ".wma", ".mp3", ".aac", ".flac", ".ogg", ".opus")
	ret.WebAudio = StringIsOneOf(e, ".mp3", ".ogg", ".wav", ".opus")
	ret.Video = StringIsOneOf(e, ".mkv", ".mp4", ".wmv", ".webm", ".avi")
	ret.WebVideo = StringIsOneOf(e, ".mp4", ".webm", ".ogv")
	ret.Flash = StringIsOneOf(e, ".swf")
	ret.Image = StringIsOneOf(e, ".jpg", ".gif", ".png", ".jpeg", ".tiff", ".tif", ".webp")
	ret.DirectlyViewable = ret.Image || ret.Flash || ret.WebVideo || ret.WebAudio
	return ret
}

func StringIsOneOf(str string, things ...string) bool {
	for _, s := range things {
		if str == s {
			return true
		}
	}
	return false
}

type Post struct {
	File_url, Id, Url, Tags string
}

func GelbooruGet(tags string) (Posts []Post, err error) {
	resp, err := http.Get("http://gelbooru.com/index.php?page=dapi&s=post&q=index&tags=" + strings.Replace(tags, " ", "+", -1))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	parser := xml.NewDecoder(resp.Body)

	//var token xml.Token

	entries := make([]Post, 0, 100)

	depth := 0
	for {
		token, err := parser.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				return nil, err
			}
		}
		switch t := token.(type) {
		case xml.StartElement:
			elmt := xml.StartElement(t)
			//name := elmt.Name.Local

			if elmt.Name.Local == "post" {

				Post := new(Post)

				for i := 0; i < len(elmt.Attr); i++ {
					switch elmt.Attr[i].Name.Local {
					case "file_url":
						Post.File_url = elmt.Attr[i].Value
						break
					case "id":
						Post.Id = elmt.Attr[i].Value
						Post.Url = "http://gelbooru.com/index.php?page=post&s=view&id=" + Post.Id
						break
					case "tags":
						Post.Tags = elmt.Attr[i].Value
					}
				}

				entries = append(entries, *Post)
			}

			//printElmt(name, depth)
			depth++

		case xml.EndElement:
			depth--
			//elmt := xml.EndElement(t)
			//name := elmt.Name.Local
			//printElmt(name, depth)
		case xml.CharData:
			//bytes := xml.CharData(t)
			//printElmt("\"" + string([]byte(bytes)) + "\"", depth)
		case xml.Comment:
			//printElmt("Comment", depth)
		case xml.ProcInst:
			//printElmt("ProcInst", depth)
		case xml.Directive:
			//printElmt("Directive", depth)
		default:
			fmt.Println("Unknown")
		}
	}
	return entries, nil
}
