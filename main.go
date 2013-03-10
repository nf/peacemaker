package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	datafile = flag.String("data", "user.json", "user data file")
	httpAddr = flag.String("http", ":7020", "HTTP server listen address")
	password = flag.String("password", "arnie", "administrator password")
)

var server *Server

func main() {
	flag.Parse()
	var err error
	server, err = NewServer(*datafile)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/checkin", checkinHandler)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	server.mu.Lock()
	defer server.mu.Unlock()
	err := statusTemplate.Execute(w, statusData{server.User, time.Now()})
	if err != nil {
		log.Print(err)
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("password") != *password {
		http.Error(w, "invalid password", http.StatusForbidden)
		return
	}
	mins, err := strconv.Atoi(r.FormValue("minutes"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = server.setBalance(r.FormValue("username"), r.FormValue("kind"), mins)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func checkinHandler(w http.ResponseWriter, r *http.Request) {
	ok, err := server.CheckIn(r.FormValue("username"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "out of time", http.StatusForbidden)
		return
	}
	fmt.Fprint(w, "OK")
}

type statusData struct {
	User map[string]*User
	Now  time.Time
}

var statusTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<body>
{{range $name, $u := .User}}
	<h1>{{$name}}</h1>
	{{if $u.Running $.Now}}<h3>Session open</h3>{{end}}
	<ul>
	{{range $u.Balance}}
		<li>{{.}}</li>
	{{end}}
	</ul>
	<form method="POST" action="/set">
	<h3>Set Balance</h3>
	<input type="hidden" name="username" value="{{$name}}">
	<table>
	<tr>
		<td>Password</td>
		<td><input type="password" name="password"></td>
	</tr>
	<tr>
		<td>Kind</td>
		<td>
		<select name="kind">
		{{range $u.Balance}}<option value="{{.Kind}}">{{.Kind}}</option>{{end}}
		</select>
		</td>
	</tr>
	<tr>
		<td>Minutes</td>
		<td><input type="text" name="minutes" value="60"></td>
	</tr>
	<tr>
		<td></td>
		<td><input type="submit" value="Set"></td>
	</tr>
	</table>
	</form>
{{end}}
</body>
</html>
`))
