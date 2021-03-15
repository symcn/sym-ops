package debug

import (
	"net/http"
	"net/http/pprof"
	"sort"
	"text/template"

	"k8s.io/klog/v2"
)

var indexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>SymOps Debug Console</title>
</head>
<style>
#endpoints {
  font-family: "Trebuchet MS", Arial, Helvetica, sans-serif;
  border-collapse: collapse;
}

#endpoints td, #endpoints th {
  border: 1px solid #ddd;
  padding: 8px;
}

#endpoints tr:nth-child(even){background-color: #f2f2f2;}

#endpoints tr:hover {background-color: #ddd;}

#endpoints th {
  padding-top: 12px;
  padding-bottom: 12px;
  text-align: left;
  background-color: black;
  color: white;
}
</style>
<body>
<br/>
<table id="endpoints">
<tr><th>Endpoint</th><th>Description</th></tr>
{{range .}}
	<tr>
	<td><a href='{{.Href}}'>{{.Name}}</a></td><td>{{.Help}}</td>
	</tr>
{{end}}
</table>
<br/>
</body>
</html>
`))

var (
	debugHandlers = map[string]string{}
)

// InitDebug initializes the debug handlers and adds a debug in-memory registry.
func InitDebug(mux *http.ServeMux, enableProfiling bool) {
	mux.HandleFunc("/debug", debugHandler)

	if enableProfiling {
		addDebugHandler(mux, "/debug/pprof/", "Displays pprof index", pprof.Index)
		addDebugHandler(mux, "/debug/pprof/cmdline", "The command line invocation of the current program", pprof.Cmdline)
		addDebugHandler(mux, "/debug/pprof/profile", "CPU profile", pprof.Profile)
		addDebugHandler(mux, "/debug/pprof/symbol", "Symbol looks up the program counters listed in the request", pprof.Symbol)
		addDebugHandler(mux, "/debug/pprof/trace", "A trace of execution of the current program", pprof.Trace)
	}
}

func addDebugHandler(mux *http.ServeMux, path string, help string, handler func(http.ResponseWriter, *http.Request)) {
	debugHandlers[path] = help
	mux.HandleFunc(path, handler)
}

// debugHandler lists all the supported debug endpoints
func debugHandler(w http.ResponseWriter, req *http.Request) {
	type debugEndpoint struct {
		Name string
		Href string
		Help string
	}
	var deps []debugEndpoint

	for k, v := range debugHandlers {
		deps = append(deps, debugEndpoint{
			Name: k,
			Href: k,
			Help: v,
		})
	}

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})
	if err := indexTmpl.Execute(w, deps); err != nil {
		klog.Errorf("Error in rendering index template %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}
