// main
package main

import (
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var js *otto.Otto
var pluginfuncs map[string]pluginmap
var corefuncs map[string]func(sender, source, line string, parameters []string)

func init() {
	pluginfuncs = make(map[string]pluginmap)
	corefuncs = make(map[string]func(sender, source, line string, parameters []string))
}

func main() {

	js = otto.New()
	RegisterJSFuncs()

	filepath.Walk("plugins", func(path string, info os.FileInfo, err error) error {
		if strings.ToLower(filepath.Ext(path)) == ".js" {
			loadplugin(path)
		}
		return nil
	})

	js.Set("Subscribe", nil)
	//testit()
	handleHttpFuncs()

	if Http {
		go func() {
			err := http.ListenAndServe(Httplisten, nil)
			if err != nil {
				panic(err.Error())
			}
		}()
	}
	if Https {
		go func() {
			err := http.ListenAndServeTLS(Httpslisten, HttpsCertFile, HttpsKeyFile, nil)
			if err != nil {
				panic(err.Error())
			}
		}()
	}

	connectDB()
	RegisterIRCCoreFuncs()
	InitIRC()
	Irccon.Loop()

}

func raise(evname, sender, source, line string) {

	var args []string = make([]string, 0, 0)

	if len(strings.SplitN(line, " ", 2)) == 2 {
		args = Args(line)
	}

	ottoargs, err := js.ToValue(args)

	if err != nil {
		panic(err.Error())
	}

	if val, ok := corefuncs[evname]; ok {
		val(sender, source, line, args)
	}

	for _, val1 := range pluginfuncs {
		if val2, ok := val1[evname]; ok {

			_, err := val2.Call(otto.UndefinedValue(), sender, source, line, ottoargs)
			if err != nil {
				Irccon.Privmsg(source, "err: "+err.Error())
			}
		}
	}
}

func raiseRegex(sender, source, line string) {

	raiseIfRegex := func(sender, source, line, key string) (bool, [][]string) {
		if strings.HasPrefix(key, "regex:") {
			regex, _ := regexp.Compile(key[len("regex:"):])
			if regex.MatchString(line) {
				matches := regex.FindAllStringSubmatch(line, -1)
				return true, matches
			}
			return false, nil
		} else {
			return false, nil
		}
	}

	for key, val := range corefuncs {
		if ok, matches := raiseIfRegex(sender, source, line, key); ok {
			for _, match := range matches {
				val(sender, source, line, match)
			}

		}
	}

	for _, val1 := range pluginfuncs {
		for key2, val2 := range val1 {
			if ok, matches := raiseIfRegex(sender, source, line, key2); ok {
				for _, match := range matches {
					ottomatch, err := js.ToValue(match)
					if err != nil {
						panic(err.Error())
					}
					_, err = val2.Call(otto.UndefinedValue(), sender, source, line, ottomatch)
					if err != nil {
						Irccon.Privmsg(source, "err: "+err.Error())
					}
				}
			}
		}
	}
}

type pluginmap map[string]otto.Value

func (m pluginmap) Subscribe(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) >= 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsFunction() {
		name := call.ArgumentList[0].String()
		f := call.ArgumentList[1]
		if strings.HasPrefix(name, "regex:") {
			_, err := regexp.Compile(name[len("regex:"):])
			if err != nil {
				Irccon.Privmsg(Channel, "Regex error: "+err.Error()+" - Regex subscribe was not successful")
				return otto.FalseValue()
			}
		}
		m[name] = f
		return otto.TrueValue()
	} else {
		return otto.FalseValue()
	}
}

func unloadpluginname(name string) error {
	if _, ok := pluginfuncs[name]; ok {
		delete(pluginfuncs, name)
		return nil
	} else {
		return errors.New("Plugin is not loaded")
	}
}

func loadpluginname(name string) error {
	return loadplugin("plugins\\" + name)
}

func loadplugin(name string) error {
	b, err := ioutil.ReadFile(name)
	_, filename := filepath.Split(name)

	if _, ok := pluginfuncs[filename]; ok {
		delete(pluginfuncs, filename)
	}

	if err != nil {
		return err
	}

	m := make(pluginmap)
	js.Set("Subscribe", m.Subscribe)
	_, err = js.Run(string(b))

	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		pluginfuncs[filename] = m
		return nil
	}

}
