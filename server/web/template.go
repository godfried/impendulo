package web

import (
	"github.com/godfried/impendulo/util"
	"html/template"
	"path/filepath"
	"strings"
	"fmt"
)

var funcs = template.FuncMap{
	"reverse":     reverse,
	"projectName": projectName,
	"date": util.Date,
	"setBreaks": setBreaks,
	"address": address,
}

func address(val interface{})string{
	return fmt.Sprint(&val)
}

func setBreaks(val string) template.HTML {
	return template.HTML(strings.Replace(val, "\n", "<br>", -1))
}

var basicT = []string{filepath.Join("templates", "_base.html"), filepath.Join("templates", "index.html"), filepath.Join("templates", "messages.html")}

func T(names ...string) *template.Template {
	t := template.New("_base.html").Funcs(funcs)
	all := make([]string, len(basicT)+len(names))
	end := copy(all, basicT)
	for i, name := range names {
		all[i+end] = filepath.Join("templates", name)
	}
	t = template.Must(t.ParseFiles(all...))
	return t
}
