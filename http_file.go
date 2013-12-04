package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"path/filepath"
)

import (
	//"errors"
	"os"
	"path"
	"strings"
	"time"
)

func FileHttp() {
	HttpHandleFunc("/FileUpload", FileUpload, false)
	HttpHandleFunc("/FileDelete", FileDelete, false)
	HttpHandleFunc("/FileZip", FileZip, false)
	HttpHandleFunc("/FileCreateFolder", FileCreateFolder, false)
	//HttpHandle("/FilesAlt/", http.StripPrefix("/Files", http.HandlerFunc(FileHtmlHandler)), false)

	HttpHandle("/Files/", http.StripPrefix("/Files", http.HandlerFunc(Files)), false)
	HttpHandle("/Files/Grid/", http.StripPrefix("/Files/Grid/", http.HandlerFunc(GridFSDL)), false)
	HttpHandle("/Files/Grid", http.StripPrefix("/Files/Grid", http.HandlerFunc(GridFSDL)), false)
}

func FileCreateFolder(w http.ResponseWriter, rew *http.Request) {
	if rew.Method == "POST" {
		auth := CheckAuthCookie(rew)
		rew.ParseForm()
		prefix := rew.FormValue("prefix")
		foldername := rew.FormValue("foldername")
		base := rew.FormValue("path")
		println(prefix)
		if prefix == "" {
			http.Redirect(w, rew, "/Files", 303)
			return
		}

		if rew.Form.Get("prefix") == "Grid" {

		} else if prefix != "" {
			if val, ok := conf.FileSystemPrefixes[prefix]; ok {
				if !(auth || (!auth && !val.CreateFolderNeedsAuth)) {
					HttpAuthenticate(w, rew)
					return
				}
				p := path.Clean(string(val.Dir) + filepath.FromSlash(path.Clean(base)) + string(os.PathSeparator) + path.Clean(foldername))
				os.Mkdir(p, os.ModeDir)
				http.Redirect(w, rew, "/Files/"+prefix+"/"+base+"/", 303)
			}

		} else {

		}
	}
}

func FileZip(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()
	prefix := rew.FormValue("prefix")
	zippath := rew.FormValue("path")
	auth := CheckAuthCookie(rew)

	if rew.Form.Get("prefix") == "Grid" {
		return
	} else if prefix != "" {
		if val, ok := conf.FileSystemPrefixes[prefix]; ok {
			if !(auth || (!auth && !val.ZipNeedsAuth)) {
				HttpAuthenticate(w, rew)
				return
			}
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(filepath.Clean(zippath))+`.zip"`)
			zipp := zip.NewWriter(w)
			err := filepath.Walk(filepath.Clean(string(val.Dir)+zippath), func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					newpath, _ := filepath.Rel(filepath.Clean(string(val.Dir)+zippath), path)
					header, _ := zip.FileInfoHeader(info)
					header.Name = newpath
					w, _ := zipp.CreateHeader(header)
					f, err := os.Open(path)
					if err != nil {
						return nil
					}
					io.Copy(w, f)
					f.Close()
					return nil
				} else {
					return nil
				}
			})
			if err == nil {
				zipp.Close()
			}
		}

	}

	//rew.Pars
}

func FileUpload(w http.ResponseWriter, rew *http.Request) {
	if rew.Method == "POST" {

		err := rew.ParseMultipartForm(10000000)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		prefix := rew.FormValue("prefix")

		if prefix == "" {
			http.Redirect(w, rew, "/Files", 303)
			return
		}

		auth := CheckAuthCookie(rew)

		if rew.Form.Get("prefix") == "Grid" {
			GridFSFileUpload(w, rew)
		} else if prefix != "" {
			if val, ok := conf.FileSystemPrefixes[prefix]; ok {
				if !(auth || (!auth && !val.UploadNeedsAuth)) {
					HttpAuthenticate(w, rew)
					return
				}
				var filename string

				formfileheaderarr := rew.MultipartForm.File["file"]
				formfileheader := formfileheaderarr[0]
				formfile, err := formfileheader.Open()
				length, _ := io.Copy(ioutil.Discard, formfile)
				formfile.Seek(0, 0)
				//formfile, formfileheader, err := rew.FormFile("file")
				if err == nil {
					if rew.FormValue("filename") == "" {
						filename = formfileheader.Filename
					} else {
						filename = rew.FormValue("filename")
					}
				}

				base := rew.FormValue("path")
				unzip := rew.FormValue("unzip")
				errored := false
				if unzip == "on" {

					if !(auth || (!auth && !val.ZipNeedsAuth)) {
						HttpAuthenticate(w, rew)
						return
					}

					p := string(val.Dir) + string(os.PathSeparator) + filepath.FromSlash(path.Clean(base))
					z, err := zip.NewReader(formfile, length)

					if err != nil {
						http.Error(w, "Can't read ZIP file: "+err.Error(), 500)
						return
					}

					for _, file := range z.File {
						os.MkdirAll(p+string(os.PathSeparator)+filepath.Dir(filepath.FromSlash(file.Name)), os.ModeDir)

						if !file.FileHeader.FileInfo().IsDir() {
							f, err := os.Create(p + string(os.PathSeparator) + filepath.FromSlash(file.Name))

							f2, _ := file.Open()
							if err != nil {
								errored = true
								w.Write([]byte(file.Name + ": " + err.Error() + "\r\n"))
								break
							}

							io.Copy(f, f2)

							f.Close()
							f2.Close()
						}
					}
				} else {
					p := path.Clean(string(val.Dir) + filepath.FromSlash(path.Clean(base)) + string(os.PathSeparator) + path.Clean(filename))
					f, err := os.Create(p)
					if err != nil {
						http.Error(w, err.Error(), 500)
						errored = true
						return
					}
					err = nil
					defer f.Close()
					_, err = io.Copy(f, formfile)

					if err != nil {
						http.Error(w, err.Error(), 500)
						errored = true
						return
					}
				}

				if !errored {

					http.Redirect(w, rew, "/Files/"+prefix+"/"+base+"/", 303)
				}
			}
		} else {

		}
	} /*else if rew.Method == "GET" {
		out := make(map[string]interface{})
		path := rew.Form.Get("path")
		if prefix != "Grid" {
			out["path"] = path
		}
		out["prefix"] = prefix
		Template(TemplateFileUpload).Execute(w, out)
	}*/
}

func FileDelete(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()
	auth := CheckAuthCookie(rew)

	prefix := rew.Form.Get("prefix")
	name := rew.Form.Get("file")
	if prefix == "Grid" {
		GridFSDelete(w, rew)
	} else if prefix == "" {
		http.Redirect(w, rew, "/Files", 303)
	} else {
		if val, ok := conf.FileSystemPrefixes[prefix]; ok {
			if !(auth || (!auth && !val.DeleteNeedsAuth)) {
				HttpAuthenticate(w, rew)
				return
			}
			p := string(val.Dir) + string(os.PathSeparator) + filepath.FromSlash(path.Clean(name))
			stat, err := os.Stat(p)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			if stat.IsDir() {
				os.RemoveAll(p)
				http.Redirect(w, rew, "/Files/"+prefix+"/"+path.Dir(path.Clean(name))+"/", 303)
				return
			} else {
				err := os.Remove(p)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}

				http.Redirect(w, rew, "/Files/"+prefix+"/"+path.Dir(path.Clean(name))+"/", 303)
			}
		} else {
			http.Redirect(w, rew, "/Files", 303)
		}
	}
}

func Files(w http.ResponseWriter, rew *http.Request) {
	rew.ParseForm()
	p := strings.Split(rew.URL.Path, "/")
	prefix := ""

	if rew.URL.Path == "/" {
		out := make(map[string]interface{})
		auth := CheckAuthCookie(rew)
		out["auth"] = auth
		out["fs"] = conf.FileSystemPrefixes
		fsrights := make(map[string]bool)
		for key, val := range conf.FileSystemPrefixes {
			//fsrights[key] = make(map[string]bool)
			fsrights[key] = auth || (!auth && !val.ListNeedsAuth) || (!auth && !val.UploadNeedsAuth)
		}
		out["fsrights"] = fsrights

		Template(TemplateFilesPrefixList).Execute(w, out)
		return
	}
	if len(p) > 1 && p[1] != "" {
		
		if _, ok := rew.Form["html"]; ok {
			hname := rew.FormValue("h")
			if hname == "" {
				ft := Filetypes(rew.URL.Path)
				
				if ft.WebAudio {
					hname = "audio"
				} else if ft.Flash {
					hname = "flash"
				} else if ft.Image {
					hname = "image"
				} else if ft.WebVideo {
					hname = "video"
				}
			} 
			
			Template("handler_" + hname + ".html").Execute(w, "/Files" + rew.URL.Path)
			return
		}
		
		
		
		prefix = p[1]
		if val, ok := conf.FileSystemPrefixes[prefix]; ok {
			ok = false
			auth := CheckAuthCookie(rew)
			ok := auth || (!auth && !val.ReadNeedsAuth) || (!auth && !val.ListNeedsAuth) || (!auth && !val.UploadNeedsAuth)
			if ok {
				serveFile(w, rew, val, strings.TrimPrefix(rew.URL.Path, "/"+prefix), false, prefix)
			} else {
				// HttpAuthenticate has already sent an error.
				return
			}
		} else {
			http.Error(w, "Prefix not found", 404)
		}
	} else {

	}

}

func GridFSDL(w http.ResponseWriter, rew *http.Request) {
	auth := CheckAuthCookie(rew)

	if rew.URL.Path == "" {
		coll := MongoDB.C("fs.files")

		var result []bson.M = make([]bson.M, 0, 0)

		coll.Find(nil).All(&result)

		for _, val := range result {
			val["IsDir"] = false
		}

		Template(TemplateGridFiles).Execute(w, bson.M{"files": result, "auth": auth, "prefix": "Grid", "path": "/"})
	} else {
		var GridFS *mgo.GridFS = MongoDB.GridFS("fs")
		GridFile, err := GridFS.Open(rew.URL.Path)
		if err == nil {
			http.ServeContent(w, rew, GridFile.Name(), GridFile.UploadDate(), GridFile)
		} else {
			http.NotFound(w, rew)
		}
	}
}

func GridFSDelete(w http.ResponseWriter, rew *http.Request) {
	if auth := HttpAuthenticate(w, rew); auth {
		rew.ParseForm()
		if rew.Form.Get("file") != "" {
			GridFS := MongoDB.GridFS("fs")
			err := GridFS.Remove(rew.Form.Get("file"))
			if err != nil {
				http.Error(w, "could not delete: "+err.Error(), 500)
			} else {
				http.Redirect(w, rew, "/Files/Grid/", 303)
			}
		}
	}
}

func GridFSFileUpload(w http.ResponseWriter, rew *http.Request) {
	if auth := HttpAuthenticate(w, rew); auth {
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
							Template(TemplateFileUploaded).Execute(w, "Grid/"+filename)
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
		}
	}
}

func dirList(w http.ResponseWriter, f http.File, fs Filesystemhttp, name string, prefix string, auth bool) {
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//fmt.Fprintf(w, "<pre>\n")

	//finfo, _ := f.Stat()

	out := make(map[string]interface{})
	out["prefix"] = prefix
	out["path"] = name
	out["auth"] = auth

	authRead := auth || (!auth && !fs.ReadNeedsAuth)
	authList := auth || (!auth && !fs.ListNeedsAuth)
	authZip := auth || (!auth && !fs.ZipNeedsAuth)
	authDelete := auth || (!auth && !fs.DeleteNeedsAuth)
	authUpload := auth || (!auth && !fs.UploadNeedsAuth)
	authCreateFolder := auth || (!auth && !fs.CreateFolderNeedsAuth)

	out["authRead"] = authRead
	out["authList"] = authList
	out["authZip"] = authZip
	out["authDelete"] = authDelete
	out["authUpload"] = authUpload
	out["authCreateFolder"] = authCreateFolder

	if authList {
		files := make([]os.FileInfo, 0, 50)

		for {
			dirs, err := f.Readdir(100)
			if err != nil || len(dirs) == 0 {
				break
			}
			for _, d := range dirs {
				name := d.Name()
				if d.IsDir() {
					name += "/"
				}
				files = append(files, d)
			}
		}
		out["files"] = files
	}

	Template(TemplateFiles).Execute(w, out)

	//fmt.Fprintf(w, "</pre>\n")
}

// modtime is the modification time of the resource to be served, or IsZero().
// return value is whether this request is now complete.
func checkLastModified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	if modtime.IsZero() {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}

// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs Filesystemhttp, name string, redirect bool, prefix string) {
	const indexPage = "/index.html"

	auth := CheckAuthCookie(r)
	authRead := auth || (!auth && !fs.ReadNeedsAuth)

	// redirect .../index.html to .../
	// can't use Redirect() because that would make the path absolute,
	// which would be a problem running under StripPrefix
	if strings.HasSuffix(r.URL.Path, indexPage) {
		localRedirect(w, r, "./")
		return
	}

	f, err := fs.Dir.Open(name)
	if err != nil {
		// TODO expose actual error?
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	d, err1 := f.Stat()
	if err1 != nil {
		// TODO expose actual error?
		http.NotFound(w, r)
		return
	}

	if redirect {
		// redirect to canonical path: / at end of directory url
		// r.URL.Path always begins with /
		url := r.URL.Path
		if d.IsDir() {
			if url[len(url)-1] != '/' {
				localRedirect(w, r, path.Base(url)+"/")
				return
			}
		} else {
			if url[len(url)-1] == '/' {
				localRedirect(w, r, "../"+path.Base(url))
				return
			}
		}
	}

	// use contents of index.html for directory, if present
	if d.IsDir() {
		index := name + indexPage
		ff, err := fs.Dir.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		}
	}

	// Still a directory? (we didn't find an index.html file)
	if d.IsDir() {
		if checkLastModified(w, r, d.ModTime()) {
			return
		}
		dirList(w, f, fs, name, prefix, auth)
		return
	}

	if authRead {
		// serverContent will check modification time
		http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	}

	//http.ServeContent(w, r, d.Name(), d.ModTime(), d.Size(), f)
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}
