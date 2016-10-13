package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	fn := p.Title + ".txt"
	return ioutil.WriteFile(path.Join(pDir, fn), p.Body, 0600)
}

const (
	pDir = "pages"
	tDir = "templates"
)

var (
	templates = template.Must(template.ParseFiles(path.Join(tDir, "view.html"), path.Join(tDir, "edit.html")))
	validPath = regexp.MustCompile("^/(view|edit|save)/([a-zA-Z0-9]+)$")
)

func loadPage(title string) (*Page, error) {
	fn := title + ".txt"
	body, err := ioutil.ReadFile(path.Join(pDir, fn))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/home", http.StatusFound)
}

func main() {
	addr := "127.0.0.1:8080"
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", rootHandler)
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Printf("Wiki running [%s]...\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
