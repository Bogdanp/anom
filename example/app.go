package app

import (
	"html/template"
	"net/http"
	"strconv"

	"google.golang.org/appengine"

	"github.com/Bogdanp/anom"
)

// Post is the model struct for blog posts.
type Post struct {
	anom.Meta

	Title   string
	Content string
}

func init() {
	http.HandleFunc("/", listPosts)
	http.HandleFunc("/view/", viewPost)
	http.HandleFunc("/delete/", deletePost)
}

func listPosts(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)

	if req.Method == http.MethodPost {
		p := &Post{
			Title:   req.PostFormValue("title"),
			Content: req.PostFormValue("content"),
		}
		assert(anom.Put(ctx, p))
		http.Redirect(rw, req, "/", http.StatusFound)
		return
	}

	var posts []*Post
	keys, err := anom.Query("Post").GetAll(ctx, &posts)
	assert(err)

	for i, p := range posts {
		p.Key = keys[i]
	}

	renderTemplate(rw, "list.tmpl", posts)
}

func viewPost(rw http.ResponseWriter, req *http.Request) {
	var p Post

	ctx := appengine.NewContext(req)
	sid := req.URL.Path[len("/view/"):]
	id, err := strconv.ParseInt(sid, 10, 64)
	assert(err)
	assert(anom.Get(ctx, &p, anom.WithIntID(ctx, id)))

	renderTemplate(rw, "post.tmpl", &p)
}

func deletePost(rw http.ResponseWriter, req *http.Request) {
	var p Post

	ctx := appengine.NewContext(req)
	sid := req.URL.Path[len("/delete/"):]
	id, err := strconv.ParseInt(sid, 10, 64)
	assert(err)
	assert(anom.Get(ctx, &p, anom.WithIntID(ctx, id)))
	assert(anom.Delete(ctx, &p))

	http.Redirect(rw, req, "/", http.StatusFound)
}

func renderTemplate(rw http.ResponseWriter, f string, d interface{}) {
	t, err := template.ParseFiles(f)
	assert(err)
	assert(t.Execute(rw, d))
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
