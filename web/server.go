//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//Package web provides a webserver which allows for: viewing
//of results; administration of submissions, projects and tools; user management;
package web

import (
	"code.google.com/p/gorilla/pat"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"path/filepath"
	"strconv"
)

var (
	router    *pat.Router
	staticDir string
	running   bool
)

const (
	LOG_SERVER      = "webserver/server.go"
	PORT       uint = 8080
)

func init() {
	logs, e := util.LogDir()
	if e != nil {
		panic(e)
	}
	//Setup the router.
	router = pat.New()
	GenerateDownloads(router, Downloaders())
	GenerateGets(router, Getters(), Views())
	GeneratePosts(router, Posters(), IndexPosters())
	GenerateViews(router, Views())
	router.Add("GET", "/chart", getChart())
	router.Add("GET", "/tools", getTools())
	router.Add("GET", "/users", getUsers())
	router.Add("GET", "/skeletons", getSkeletons())
	router.Add("GET", "/code", getCode())
	router.Add("GET", "/static/", FileHandler(StaticDir()))
	router.Add("GET", "/static", RedirectHandler("/static/"))
	router.Add("GET", "/logs/", FileHandler(logs))
	router.Add("GET", "/logs", RedirectHandler("/logs/"))
	router.Add("GET", "/", Handler(LoadView("homeview", "home"))).Name("index")
}

func getUsers() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		projectId, msg, e := getProjectId(req)
		if e != nil {
			fmt.Fprint(w, msg)
			return
		}
		u, e := users(projectId)
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		b, e := util.JSON(map[string]interface{}{"users": u})
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		fmt.Fprint(w, string(b))
	})
}

func getTools() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		projectId, msg, e := getProjectId(req)
		if e != nil {
			fmt.Fprint(w, msg)
			return
		}
		t, e := tools(projectId)
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		b, e := util.JSON(map[string]interface{}{"tools": t})
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		fmt.Fprint(w, string(b))
	})
}

func getCode() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resultId, msg, e := getId(req, "resultid", "result")
		if e != nil {
			fmt.Fprint(w, msg)
			return
		}
		r, e := db.ToolResult(bson.M{db.ID: resultId}, bson.M{db.FILEID: 1})
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		f, e := db.File(bson.M{db.ID: r.GetFileId()}, nil)
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		b, e := util.JSON(map[string]interface{}{"code": string(f.Data)})
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		fmt.Fprint(w, string(b))
	})
}

func getSkeletons() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		projectId, msg, e := getProjectId(req)
		if e != nil {
			fmt.Fprint(w, msg)
			return
		}
		vals, e := skeletons(projectId)
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		b, e := util.JSON(map[string]interface{}{"skeletons": vals})
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		fmt.Fprint(w, string(b))
	})
}

func getChart() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		subId, msg, e := getSubId(req)
		if e != nil {
			fmt.Fprint(w, msg)
			return
		}
		n, e := GetString(req, "file")
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		r, e := GetString(req, "result")
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		switch r {
		case jacoco.NAME:
			cId, e := util.ReadId(req.FormValue("childfileid"))
			if e != nil {
				fmt.Fprint(w, e.Error())
				return
			}
			r += "-" + cId.Hex()
		case junit.NAME:
			r, _ = util.Extension(n)
			if cId, e := util.ReadId(req.FormValue("childfileid")); e == nil {
				r += "-" + cId.Hex()
			}
		}
		files, e := db.Files(bson.M{db.SUBID: subId, db.NAME: n}, bson.M{db.DATA: 0})
		c, e := LoadChart(r, files)
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		b, e := util.JSON(c)
		if e != nil {
			fmt.Fprint(w, e.Error())
			return
		}
		fmt.Fprint(w, string(b))
	})
}

//StaticDir retrieves the directory containing all the static files for the web server.
func StaticDir() string {
	if staticDir != "" {
		return staticDir
	}
	iPath, e := util.InstallPath()
	if e != nil {
		return ""
	}
	staticDir = filepath.Join(iPath, "static")
	return staticDir
}

//getRoute retrieves a route for a given name.
func getRoute(name string) string {
	u, e := router.GetRoute(name).URL()
	if e != nil {
		return "/"
	}
	return u.Path
}

//Run starts up the webserver if it is not currently running.
func Run(port uint) {
	if Active() {
		return
	}
	setActive(true)
	defer setActive(false)
	if e := http.ListenAndServe(":"+strconv.Itoa(int(port)), router); e != nil {
		util.Log(e)
	}
}

//Active is whether the server is currently running.
func Active() bool {
	return running
}

//setActive
func setActive(active bool) {
	running = active
}
