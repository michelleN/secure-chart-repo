package main

import (
	"fmt"
	"html/template"
	//"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/repo"
)

const indexHTMLTemplate = `
<html>
<head>
	<title>Private Helm Repository</title>
</head>
<h1>Helm Charts Repository</h1>
<ul>
{{range $name, $ver := .Index.Entries}}
  <li>{{$name}}<ul>{{range $ver}}
    <li><a href="{{index .URLs 0}}">{{.Name}}-{{.Version}}</a></li>
  {{end}}</ul>
  </li>
{{end}}
</ul>
<body>
<p>Last Generated: {{.Index.Generated}}</p>
</body>
</html>
`

const (
	username = "user"
	password = "shhhh"
)

// RepositoryServer is an HTTP handler for serving a chart repository.
type RepositoryServer struct {
	RepoPath string
}

func main() {

	err := startLocalRepo(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println(err)
	}

}

// ServeHTTP implements the http.Handler interface.
func (s *RepositoryServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	uri := r.URL.Path
	switch uri {
	case "/", "/charts/", "/charts/index.html", "/charts/index":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		s.htmlIndex(w, r)
	default:
		file := strings.TrimPrefix(uri, "/charts/")
		http.ServeFile(w, r, filepath.Join(s.RepoPath, file))
	}
}

func auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if !check(user, pass) {
			http.Error(w, "Unauthorized.", 401)
			return
		}

		fn(w, r)
	}
}

func check(u, p string) bool {
	if u == username && p == password {
		return true
	}

	return false
}

// startLocalRepo starts a web server and serves files from the given path
func startLocalRepo(path, address string) error {
	if address == "" {
		address = "127.0.0.1:8879"
	}
	s := &RepositoryServer{RepoPath: path}
	return http.ListenAndServe(address, auth(s.ServeHTTP))
}

func (s *RepositoryServer) htmlIndex(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("index.html").Parse(indexHTMLTemplate))
	// load index
	lrp := filepath.Join(s.RepoPath, "index.yaml")
	i, err := repo.LoadIndexFile(lrp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	data := map[string]interface{}{
		"Index": i,
	}
	if err := t.Execute(w, data); err != nil {
		fmt.Fprintf(w, "Template error: %s", err)
	}
}
