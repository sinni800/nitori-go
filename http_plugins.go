package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func PluginHttp() {
	HttpHandleFunc("/plugins", PluginList, true)
	HttpHandleFunc("/plugins/edit", PluginEdit, true)
	HttpHandleFunc("/plugins/load", PluginLoad, true)
	HttpHandleFunc("/plugins/unload", PluginUnload, true)
	HttpHandleFunc("/plugins/delete", PluginDelete, true)
}

func PluginList(w http.ResponseWriter, rew *http.Request) {
	stuff := make(map[string]interface{})
	plugins := make(map[string]map[string]interface{})
	stuff["plugins"] = plugins
	instances := make([]string, 0, 0)

	for _, i := range conf.Instances {

		instances = append(instances, i.Name)
	}

	stuff["instances"] = instances

	filepath.Walk(conf.Plugindir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			_, filename := filepath.Split(path)
			plugins[filename] = make(map[string]interface{})
			plugins[filename]["file"] = true
			plugins[filename]["name"] = filename
			plugins[filename]["loaded"] = make(map[string]bool)
		}
		return nil
	})

	for _, i := range conf.Instances {
		for key, value := range i.Pluginfuncs {
			if _, ok := plugins[key]; !ok {
				plugins[key] = make(map[string]interface{})
				plugins[key]["loaded"] = make(map[string]bool)
			}

			funcs := make([]string, 0, len(value))

			for key1, _ := range value {
				funcs = append(funcs, key1)
			}

			plugins[key]["functions"] = funcs
			plugins[key]["name"] = key

			plugins[key]["loaded"].(map[string]bool)[i.Name] = true
		}
	}

	err := Template(TemplatePluginlist).Execute(w, stuff)

	if err != nil {
		fmt.Println("template error: ", err.Error())
	}
}

func PluginEdit(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()

	name := rew.Form.Get("name")
	oldname := rew.FormValue("oldname")

	if rew.Method == "GET" {
		content, err := ioutil.ReadFile(conf.Plugindir + string(os.PathSeparator) + name)
		if err != nil {
			content = []byte("")
		}
		Template(TemplatePluginedit).Execute(w, map[string]string{"name": name, "content": string(content)})
	} else if rew.Method == "POST" {
		content := rew.FormValue("content")
		filename := rew.FormValue("filename")
		submit := rew.FormValue("submit")

		if filename != name {
			err := os.Remove(conf.Plugindir + string(os.PathSeparator) + oldname)
			if err != nil {
				fmt.Println("Could not delete old file:", oldname, filename, err.Error())
			}
			name = filename
		}

		if content != "" {
			ioutil.WriteFile(conf.Plugindir+string(os.PathSeparator)+filename, []byte(content), 0)
		}

		if submit == "Save and reload plugin" {
			var err error
			for _, i := range conf.Instances {
				i.unloadpluginname(oldname)
				err = i.loadpluginname(filename)
				if err != nil {
					break
				}
			}

			if err != nil {
				Template(TemplatePluginedit).Execute(w, map[string]string{"name": filename, "content": string(content), "error": err.Error()})
				return
			} else {
				w.Header().Add("Location", "/plugins/edit?name="+filename)
			}
		} else {
			w.Header().Add("Location", "/plugins")
		}

		w.WriteHeader(303)
	}

}

func PluginDelete(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()
	name := rew.FormValue("name")
	err := os.Remove(conf.Plugindir + string(os.PathSeparator) + name)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Location", "/plugins")
		w.WriteHeader(303)
	}
}

func PluginLoad(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()
	name := rew.FormValue("name")
	inst := rew.FormValue("instance")

	if inst == "all" {
		for _, i := range conf.Instances {
			err := i.loadpluginname(name)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				break
			} else {
				w.Header().Add("Location", "/plugins")
				w.WriteHeader(303)
			}
		}
	} else {
		if i, ok := conf.Instances[inst]; ok {
			err := i.loadpluginname(name)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
			} else {
				w.Header().Add("Location", "/plugins")
				w.WriteHeader(303)
			}
		} else {
			w.WriteHeader(500)
			w.Write([]byte("instance not found"))
		}
	}
}

func PluginUnload(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()
	name := rew.FormValue("name")
	inst := rew.FormValue("instance")
	if inst == "all" {
		for _, i := range conf.Instances {
			i.unloadpluginname(name)
		}
		w.Header().Add("Location", "/plugins")
		w.WriteHeader(303)
	} else {
		if i, ok := conf.Instances[inst]; ok {
			err := i.unloadpluginname(name)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
			} else {
				w.Header().Add("Location", "/plugins")
				w.WriteHeader(303)
			}
		} else {
			w.WriteHeader(500)
			w.Write([]byte("instance not found"))
		}
	}
}
