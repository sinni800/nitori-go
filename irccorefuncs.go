package main

import (
	"code.google.com/p/go.net/html"
	"net/http"
	"strings"
)

func RegisterIRCCoreFuncs() {
	corefuncs["!loadplugin"] = func(sender string, source string, line string, parameters []string) {
		err := loadpluginname(parameters[0])
		if err != nil {
			Irccon.Privmsg(source, "Error occured: "+err.Error())
		} else {
			Irccon.Privmsg(source, "Plugin (re-)loaded")
		}
	}

	corefuncs["!unloadplugin"] = func(sender string, source string, line string, parameters []string) {
		err := unloadpluginname(parameters[0])

		if err != nil {
			Irccon.Privmsg(source, "Error occured: "+err.Error())
		} else {
			Irccon.Privmsg(source, "Plugin unloaded")
		}
	}

	corefuncs["!pluginlist"] = func(sender string, source string, line string, parameters []string) {

		out := ""

		for key, _ := range pluginfuncs {
			out += key + ", "
		}

		Irccon.Privmsg(source, "Loaded plugins: "+out)
	}

	/*corefuncs["!rescuejs"] = func(sender string, source string, line string, parameters []string) {
		js = otto.New()
		RegisterJSFuncs()
	}*/

	corefuncs["regex:((([A-Za-z]{3,9}:(?:\\/\\/)?)(?:[-;:&=\\+\\$,\\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\\+\\$,\\w]+@)[A-Za-z0-9.-]+)((?:\\/[\\+~%\\/.\\w-_]*)?\\??(?:[-\\+=&;%@.\\w_]*)#?(?:[\\w]*))?)"] = func(sender string, source string, line string, match []string) {
		//Irccon.Privmsg(source, "URL Detected.")

		if strings.Contains(match[0], "gelbooru.com") || strings.Contains(match[0], "danbooru") {
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", match[0], nil)

		if err != nil {
			return //nil, err
		}

		req.Header.Add("Accept", "text/html, application/xhtml+xml")

		resp, err := client.Do(req)

		if err != nil {
			return //nil, err
		} else if resp.StatusCode != 200 {
			return
		}

		defer resp.Body.Close()

		tok := html.NewTokenizer(resp.Body)

		for {
			t := tok.Next()
			if t == html.ErrorToken {
				return
			}
			if t == html.StartTagToken {
				if name, _ := tok.TagName(); string(name) == "title" {
					if tok.Next() == html.TextToken {
						Irccon.Privmsg(source, "URL Title: \x02\x034"+strings.Trim(string(tok.Text()), "\r\n"))
					}
					return
				}
			}
		}
	}
}
