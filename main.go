package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/rpc"
)

var (
	datafile = flag.String("data", "user.json", "user data file")
	httpAddr = flag.String("http", ":7020", "HTTP server listen address")
)

var server *Server

func main() {
	flag.Parse()
	var err error
	server, err = NewServer(*datafile)
	if err != nil {
		log.Fatal(err)
	}
	rpc.Register(server)
	rpc.HandleHTTP()
	http.HandleFunc("/", statusHandler)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	server.mu.Lock()
	defer server.mu.Unlock()
	err := statusTemplate.Execute(w, server.User)
	if err != nil {
		log.Print(err)
	}
}

var statusTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<body>
{{range $name, $u := .}}
	<h2>{{$name}}</h2>
	{{if $u.Running}}<h3>Session open</h3>{{end}}
	{{range $u.Balance}}
		<p>{{.Minutes}} {{.Kind}} minutes</p>
	{{end}}
{{end}}
</body>
</html>
`))
