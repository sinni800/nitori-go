package main

import (
	"code.google.com/p/go.net/html"
	"net/http"
	"os"
	"strings"
)

func (i *instance) RegisterIRCCoreFuncs() {
	i.Corefuncs["#!exit"] = func(source messagesource, line string, parameters []string) {
		os.Exit(0)
	}

	i.Corefuncs["#!loadplugin"] = func(source messagesource, line string, parameters []string) {
		err := i.loadpluginname(parameters[0])
		if err != nil {
			i.Irc.Privmsg(source.Source, "Error occured: "+err.Error())
		} else {
			i.Irc.Privmsg(source.Source, "Plugin (re-)loaded")
		}
	}

	i.Corefuncs["#!unloadplugin"] = func(source messagesource, line string, parameters []string) {
		err := i.unloadpluginname(parameters[0])

		if err != nil {
			i.Irc.Privmsg(source.Source, "Error occured: "+err.Error())
		} else {
			i.Irc.Privmsg(source.Source, "Plugin unloaded")
		}
	}

	i.Corefuncs["#!pluginlist"] = func(source messagesource, line string, parameters []string) {

		out := ""

		for key, _ := range i.Pluginfuncs {
			out += key + ", "
		}

		i.Irc.Privmsg(source.Source, "Loaded plugins: "+out)
	}

	i.Corefuncs["!identify"] = func(source messagesource, line string, parameters []string) {
		if strings.HasPrefix(source.Source, "#") {
			if line != "" {
				i.Irc.Privmsg(source.Source, "You shouldn't have done that.")
			} else {
				i.Irc.Privmsgf(source.Source, "Usage: /msg %s !auth <password>", i.Irccfg.Nick)
			}
		} else {
			if line != "" {
				if Authenticate(source.Nick, line) {
					i.Authenticatednicks[source.Nick] = true
					i.Irc.Privmsg(source.Source, "OK.")
				} else {
					i.Irc.Privmsg(source.Source, "Password is wrong.")
				}
			} else {
				i.Irc.Privmsg(source.Source, "!auth <password> - log in to the bot")
			}
		}
	}

	/*corefuncs["!rescuejs"] = func(sender string, source string, line string, parameters []string) {
		js = otto.New()
		RegisterJSFuncs()
	}*/

	i.Corefuncs["regex:((([A-Za-z]{3,9}:(?:\\/\\/)?)(?:[-;:&=\\+\\$,\\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;:&=\\+\\$,\\w]+@)[A-Za-z0-9.-]+)((?:\\/[\\+~%\\/.\\w-_]*)?\\??(?:[-\\+=&;%@.\\w_]*)#?(?:[\\w]*))?)"] = func(source messagesource, line string, match []string) {
		//i.Irc.Privmsg(source, "URL Detected.")

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
						i.Irc.Privmsg(source.Source, "URL Title: \x02\x034"+strings.Trim(string(tok.Text()), "\r\n"))
					}
					return
				}
			}
		}
	}
}
