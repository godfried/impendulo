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

package web

import (
	"code.google.com/p/gorilla/pat"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"net/http"
	"strings"
)

type (
	//Args represents arguments passed to html templates or to template.Execute.
	Args   map[string]interface{}
	Getter func(r *http.Request, c *Context) (Args, string, error)
)

var (
	getters map[string]Getter
)

//Getters retrieves all getters
func Getters() map[string]Getter {
	if getters == nil {
		getters = defaultGetters()
	}
	return getters
}

//defaultGetters loads the default getters.
func defaultGetters() map[string]Getter {
	return map[string]Getter{
		"configview": configView, "editdbview": editDBView,
		"loadproject": loadProject, "loadsubmission": loadSubmission,
		"loadfile": loadFile, "loaduser": loadUser,
		"displayresult": displayResult, "getfiles": getFiles, "displaychildresult": displayChildResult,
		"getsubmissionschart": getSubmissionsChart, "getsubmissions": getSubmissions,
	}
}

//GenerateGets loads post request handlers and adds them to the router.
func GenerateGets(r *pat.Router, gets map[string]Getter, views map[string]string) {
	for n, f := range gets {
		h := f.CreateGet(n, views[n])
		p := "/" + n
		r.Add("GET", p, Handler(h)).Name(n)
	}
}

func (g Getter) CreateGet(name, view string) Handler {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		a, m, e := g(r, c)
		if m != "" {
			c.AddMessage(m, e != nil)
		}
		if e != nil {
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return e
		}
		t, e := util.GetStrings(a, "templates")
		if e != nil {
			c.AddMessage("Could not load page.", true)
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return e
		}
		delete(a, "templates")
		c.Browse.View = view
		if c.Browse.View == "home" {
			c.Browse.SetLevel(name)
		}
		a["ctx"] = c
		return T(append(t, getNav(c))...).Execute(w, a)
	}
}

//configView loads a tool's configuration page.
func configView(r *http.Request, c *Context) (Args, string, error) {
	t, e := GetString(r, "tool")
	if e != nil {
		t = "none"
	}
	return Args{"tool": t, "templates": []string{"configview", toolTemplate(t)}},
		"", nil
}

//getSubmissions displays a list of submissions.
func getSubmissions(r *http.Request, c *Context) (Args, string, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, "Could not load submissions.", e
	}
	var m bson.M
	if !c.Browse.IsUser {
		m = bson.M{db.PROJECTID: c.Browse.Pid}
	} else {
		m = bson.M{db.USER: c.Browse.Uid}
	}
	s, e := db.Submissions(m, nil, "-"+db.TIME)
	if e != nil {
		return nil, "Could not load submissions.", e
	}
	t := make([]string, 1)
	if c.Browse.IsUser {
		t[0] = "usersubmissionresult"
	} else {
		t[0] = "projectsubmissionresult"
	}
	return Args{"subRes": s, "templates": t}, "", nil
}

//getFiles diplays information about files.
func getFiles(r *http.Request, c *Context) (Args, string, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, "Could not retrieve files.", e
	}
	m := bson.M{db.SUBID: c.Browse.Sid, db.OR: [2]bson.M{bson.M{db.TYPE: project.SRC}, bson.M{db.TYPE: project.TEST}}}
	f, e := db.FileInfos(m)
	if e != nil {
		return nil, "Could not retrieve files.", e
	}
	return Args{"fileInfo": f, "templates": []string{"fileresult"}}, "", nil
}

//editDBView
func editDBView(r *http.Request, c *Context) (Args, string, error) {
	editing, e := GetString(r, "editing")
	if e != nil {
		editing = "Project"
	}
	t := []string{"editdbview", "edit" + strings.ToLower(editing)}
	return Args{"editing": editing, "templates": t},
		"", nil
}

func loadProject(r *http.Request, c *Context) (Args, string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return nil, m, e
	}
	p, e := db.Project(bson.M{db.ID: pid}, nil)
	if e != nil {
		return nil, "Could not find project.", e
	}
	return Args{"editing": "Project", "project": p,
		"templates": []string{"editdbview", "editproject"}}, "", nil
}

func loadUser(r *http.Request, c *Context) (Args, string, error) {
	n, m, e := getUserId(r)
	if e != nil {
		return nil, m, e
	}
	u, e := db.User(n)
	if e != nil {
		return nil, fmt.Sprintf("Could not find user %s.", n), e
	}
	return Args{"editing": "User", "user": u, "templates": []string{"editdbview", "edituser"}}, "", nil
}

func loadSubmission(r *http.Request, c *Context) (Args, string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return nil, m, e
	}
	sid, m, e := getSubId(r)
	if e != nil {
		return nil, m, e
	}
	s, e := db.Submission(bson.M{db.ID: sid}, nil)
	if e != nil {
		return nil, "Could not find submission.", e
	}
	return Args{"editing": "Submission", "projectId": pid, "submission": s,
		"templates": []string{"editdbview", "editsubmission"}}, "", nil
}

func loadFile(r *http.Request, c *Context) (Args, string, error) {
	pid, m, e := getProjectId(r)
	if e != nil {
		return nil, m, e
	}
	sid, m, e := getSubId(r)
	if e != nil {
		return nil, m, e
	}
	fid, m, e := getFileId(r)
	if e != nil {
		return nil, m, e
	}
	f, e := db.File(bson.M{db.ID: fid}, nil)
	if e != nil {
		return nil, "Could not find file.", e
	}
	return Args{"editing": "File", "projectId": pid, "submissionId": sid, "file": f,
		"templates": []string{"editdbview", "editfile"}}, "", nil
}

//displayResult displays a tool's result.
func displayResult(r *http.Request, c *Context) (Args, string, error) {
	a, e := _displayResult(r, c)
	if e != nil {
		return nil, "Could not load results.", e
	}
	return a, "", nil
}

func _displayResult(r *http.Request, c *Context) (Args, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, e
	}
	if c.Browse.childResult() {
		return _displayChildResult(r, c)
	}
	fs, e := Snapshots(c.Browse.Sid, c.Browse.File, c.Browse.Type)
	if e != nil {
		return nil, e
	}
	cf, e := getFile(fs[c.Browse.Current].Id)
	if e != nil {
		return nil, e
	}
	rs, e := analysisNames(c.Browse.Pid, c.Browse.Type)
	if e != nil {
		return nil, e
	}
	cr, e := GetResult(c.Browse.Result, cf.Id)
	if e != nil {
		return nil, e
	}
	nf, e := getFile(fs[c.Browse.Next].Id)
	if e != nil {
		return nil, e
	}
	nr, e := GetResult(c.Browse.Result, nf.Id)
	if e != nil {
		return nil, e
	}
	t := []string{"analysisview", "pager", "srcanalysis", ""}
	if !isError(cr) || isError(nr) {
		t[3] = cr.Template()
	} else {
		t[3] = nr.Template()
	}
	return Args{
		"files": fs, "currentFile": cf, "currentResult": cr, "results": rs,
		"nextFile": nf, "nextResult": nr, "templates": t,
	}, nil
}

func displayChildResult(r *http.Request, c *Context) (Args, string, error) {
	a, e := _displayChildResult(r, c)
	if e != nil {
		return nil, "Could not load results.", e
	}
	return a, "", nil
}

func _displayChildResult(r *http.Request, c *Context) (Args, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, e
	}
	parentFiles, e := Snapshots(c.Browse.Sid, c.Browse.File, c.Browse.Type)
	if e != nil {
		return nil, e
	}
	if cur, ce := getCurrent(r, len(parentFiles)-1); ce == nil {
		c.Browse.Current = cur
	}
	if next, ne := getNext(r, len(parentFiles)-1); ne == nil {
		c.Browse.Next = next
	}
	currentFile, e := getFile(parentFiles[c.Browse.Current].Id)
	if e != nil {
		return nil, e
	}
	results, e := analysisNames(c.Browse.Pid, c.Browse.Type)
	if e != nil {
		return nil, e
	}
	nextFile, e := getFile(parentFiles[c.Browse.Next].Id)
	if e != nil {
		return nil, e
	}
	childFiles, e := Snapshots(c.Browse.Sid, c.Browse.ChildFile, c.Browse.ChildType)
	if e != nil {
		return nil, e
	}
	currentChild, e := db.File(bson.M{db.ID: childFiles[c.Browse.CurrentChild].Id}, nil)
	if e != nil {
		return nil, e
	}
	var hc, hn string
	var ic, in bson.ObjectId
	switch c.Browse.ChildType {
	case project.SRC:
		ic = currentFile.Id
		in = nextFile.Id
		hc = currentChild.Id.Hex()
		hn = currentChild.Id.Hex()
	case project.TEST:
		ic = currentChild.Id
		in = currentChild.Id
		hc = currentFile.Id.Hex()
		hn = nextFile.Id.Hex()
	}
	currentChildResult, e := GetChildResult(c.Browse.Result, hc, ic)
	if e != nil {
		return nil, e
	}
	nextChildResult, e := GetChildResult(c.Browse.Result, hn, in)
	if e != nil {
		return nil, e
	}
	t := []string{"analysisview", "pager", "testanalysis", ""}
	if !isError(currentChildResult) || isError(nextChildResult) {
		t[3] = currentChildResult.Template()
	} else {
		t[3] = nextChildResult.Template()
	}
	return Args{
		"files": parentFiles, "childFiles": childFiles, "currentFile": currentFile, "nextFile": nextFile,
		"results": results, "childFile": currentChild, "currentChildResult": currentChildResult,
		"nextChildResult": nextChildResult, "templates": t,
	}, nil
}

//getSubmissionsChart displays a chart of submissions.
func getSubmissionsChart(r *http.Request, c *Context) (Args, string, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, "Could not load chart.", e
	}
	var m bson.M
	if !c.Browse.IsUser {
		m = bson.M{db.PROJECTID: c.Browse.Pid}
	} else {
		m = bson.M{db.USER: c.Browse.Uid}
	}
	s, e := db.Submissions(m, nil, "-"+db.TIME)
	if e != nil {
		return nil, "Could not load chart.", e
	}
	d := SubmissionChart(s)
	t := make([]string, 1)
	if c.Browse.IsUser {
		t[0] = "usersubmissionchart"
	} else {
		t[0] = "projectsubmissionchart"
	}
	return Args{"chart": d, "templates": t}, "", nil
}
