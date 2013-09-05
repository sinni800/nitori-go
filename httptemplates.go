package main

import (
	"io/ioutil"
	//"net/http"
	"html/template"
	"os"
)

var (
	TemplateSingleColumnTable = `singlecolumntable.html`
	TemplateDNATable          = `dna.html`
	TemplateKawaiiTable       = `kawaii.html`
	TemplateSkeleton          = `skeleton.html`
	TemplatePluginlist        = `pluginlist.html`
	TemplatePluginedit        = `pluginedit.html`
	TemplateLogin             = `login.html`
	TemplateCreateUser        = `createuser.html`
	TemplateFiles             = `files.html`
	TemplateGridFiles         = `files_grid.html`
	TemplateFileUpload        = `fileupload.html`
	TemplateDocs              = `docs.html`
	TemplateFilesPrefixList   = `filesprefixlist.html`
	TemplateFileUploaded      = `fileuploaded.html`
	TemplateIRCLogs           = `irclogs.html`
	TemplateIRCLatest         = `irclatest.html`
)

func init() {
}

func Template(subtmpl string) *template.Template {
	templ := template.New("main")

	skel, err := ioutil.ReadFile(conf.Web.Templatedir + string(os.PathSeparator) + "skeleton.html")

	if err != nil {
		skel = []byte(`Could not load skeleton template <br /> {{template "content" .}}`)
	}

	_, err = templ.Parse(string(skel))

	contenttmpl := templ.New("content")
	if subtmpl != "" {
		content, err := ioutil.ReadFile(conf.Web.Templatedir + string(os.PathSeparator) + subtmpl)

		if err != nil {
			contenttmpl.Parse(`Could not load subtemplate: ` + err.Error())
			return templ
		}

		_, err = contenttmpl.Parse(string(content))

		if err != nil {
			contenttmpl.Parse(`Could not load subtemplate: ` + err.Error())
		}

	} else {
		contenttmpl.Parse(` `)
	}

	return templ
}
