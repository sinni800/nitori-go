package main

import (
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (i *instance) raise(evname string, src messagesource, line string, authed bool) {

	var args []string = make([]string, 0, 0)

	//if len(strings.SplitN(line, " ", 2)) == 2 {

	args = Args(line)

	//}

	ottosrc, _ := i.js.ToValue(src)
	ottoargs, err := i.js.ToValue(args)

	if err != nil {
		panic(err.Error())
	}

	if val, ok := i.Corefuncs["#"+evname]; authed && ok {
		val(src, line, args)
	} else if val, ok := i.Corefuncs[evname]; ok {
		val(src, line, args)
	}

	for _, val1 := range i.Pluginfuncs {
		if val2, ok := val1["#"+evname]; authed && ok {
			_, err := val2.Call(otto.UndefinedValue(), ottosrc, line, ottoargs)
			if err != nil {
				i.Irc.Privmsg(src.Source, "err: "+err.Error())
			}
		} else if val2, ok := val1[evname]; ok {
			_, err := val2.Call(otto.UndefinedValue(), ottosrc, line, ottoargs)
			if err != nil {
				i.Irc.Privmsg(src.Source, "err: "+err.Error())
			}
		}
	}
}

func (i *instance) raiseRegex(src messagesource, line string, authed bool) {

	ottosrc, _ := i.js.ToValue(src)

	raiseIfRegex := func(src messagesource, line, key string) (bool, [][]string) {
		if authed && strings.HasPrefix(key, "#regex:") {
			regex, _ := regexp.Compile(key[len("#regex:"):])
			if regex.MatchString(line) {
				matches := regex.FindAllStringSubmatch(line, -1)
				return true, matches
			}
			return false, nil
		} else if strings.HasPrefix(key, "regex:") {
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

	for key, val := range i.Corefuncs {
		if ok, matches := raiseIfRegex(src, line, key); ok {
			for _, match := range matches {
				val(src, line, match)
			}

		}
	}

	for _, val1 := range i.Pluginfuncs {
		for key2, val2 := range val1 {
			if ok, matches := raiseIfRegex(src, line, key2); ok {
				for _, match := range matches {
					ottomatch, err := i.js.ToValue(match)
					if err != nil {
						panic(err.Error())
					}
					_, err = val2.Call(otto.UndefinedValue(), ottosrc, line, ottomatch)
					if err != nil {
						i.Irc.Privmsg(src.Source, "err: "+err.Error())
					}
				}
			}
		}
	}
}

func (i *instance) unloadpluginname(name string) error {
	if _, ok := i.Pluginfuncs[name]; ok {
		delete(i.Pluginfuncs, name)
		delete(i.Plugindocs, name)
		return nil
	} else {
		return errors.New("Plugin is not loaded")
	}
}

func (i *instance) loadpluginname(name string) error {
	return i.loadplugin(conf.Plugindir + string(os.PathSeparator) + name)
}

func (i *instance) loadplugin(name string) error {
	b, err := ioutil.ReadFile(name)
	_, filename := filepath.Split(name)

	if _, ok := i.Pluginfuncs[filename]; ok {
		delete(i.Pluginfuncs, filename)
		delete(i.Plugindocs, filename)
	}

	if err != nil {
		return err
	}

	m := make(map[string]otto.Value)
	i.js.Set("Subscribe", func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) >= 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsFunction() {
			name := call.ArgumentList[0].String()
			f := call.ArgumentList[1]
			if strings.HasPrefix(name, "regex:") {
				_, err := regexp.Compile(name[len("regex:"):])
				if err != nil {
					i.Irc.Privmsg(i.Irccfg.Channel, "Regex error: "+err.Error()+" - Regex subscribe was not successful")
					return otto.FalseValue()
				}
			}
			m[name] = f
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})

	md := make(map[string]string)
	i.js.Set("Doc", func(call otto.FunctionCall) otto.Value {
		if len(call.ArgumentList) >= 2 && call.ArgumentList[0].IsString() && call.ArgumentList[1].IsString() {
			name := call.ArgumentList[0].String()
			doc := call.ArgumentList[1].String()
			md[name] = doc
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})

	_, err = i.js.Run(string(b))
	i.js.Set("Subscribe", nil)
	i.js.Set("Doc", nil)

	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		i.Pluginfuncs[filename] = m
		i.Plugindocs[filename] = md
		return nil
	}

}
