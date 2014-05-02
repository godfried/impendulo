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
	"github.com/godfried/impendulo/user"

	"net/http"
	"strings"
)

type (
	Perm int
)

const (
	OUT user.Permission = 42
)

var (
	viewRoutes  map[string]string
	permissions map[string]user.Permission
	out         = []string{
		"registerview", "register", "login",
	}
	none = []string{
		"index", "", "homeview", "projectresult",
		"userresult", "displayresult", "displaychildresult",
		"getfiles", "favicon.ico", "getsubmissions", "submissionschartview",
		"static", "userchart", "projectchart",
	}
	student = []string{
		"projectdownloadview", "skeleton.zip",
		"intloladownloadview", "intlola.zip",
		"archiveview", "submitarchive", "logout",
	}
	teacher = []string{
		"skeletonview", "addskeleton", "projectview",
		"addproject", "runtoolsview", "runtools", "configview",
	}
	admin = []string{
		"projectdeleteview", "deleteproject", "userdeleteview",
		"deleteuser", "resultsdeleteview", "deleteresults",
		"deleteskeletons", "skeletondeleteview",
		"importdataview", "exportdataview",
		"importdata", "exportdb.zip", "statusview",
		"evaluatesubmissionsview", "evaluatesubmissions", "logs",
		"editdbview", "loadproject", "editproject", "loaduser",
		"edituser", "loadsubmission", "editsubmission", "loadfile",
		"editfile",
	}

	homeViews = []string{
		"homeview", "userresult", "projectresult",
		"userchart", "projectchart",
		"displayresult", "getfiles", "submissionschartview",
		"getsubmissions", "displaychildresult",
	}
	submitViews = []string{
		"skeletonview", "archiveview", "projectview",
		"configview",
	}
	registerViews = []string{"registerview"}
	downloadViews = []string{"projectdownloadview", "intloladownloadview"}
	deleteViews   = []string{"projectdeleteview", "userdeleteview", "resultsdeleteview", "skeletondeleteview"}
	statusViews   = []string{"statusview"}
	toolViews     = []string{"runtoolsview", "evaluatesubmissionsview"}
	dataViews     = []string{
		"importdataview", "exportdataview", "editdbview",
		"loadproject", "loadsubmission", "loadfile", "loaduser",
	}
)

//Views loads all views.
func Views() map[string]string {
	if viewRoutes != nil {
		return viewRoutes
	}
	viewRoutes = make(map[string]string)
	for _, n := range homeViews {
		viewRoutes[n] = "home"
	}
	for _, n := range submitViews {
		viewRoutes[n] = "submit"
	}
	for _, n := range registerViews {
		viewRoutes[n] = "register"
	}
	for _, n := range downloadViews {
		viewRoutes[n] = "download"
	}
	for _, n := range deleteViews {
		viewRoutes[n] = "delete"
	}
	for _, n := range statusViews {
		viewRoutes[n] = "status"
	}
	for _, n := range toolViews {
		viewRoutes[n] = "tool"
	}
	for _, n := range dataViews {
		viewRoutes[n] = "data"
	}
	return viewRoutes
}

//Permissions loads all permissions.
func Permissions() map[string]user.Permission {
	if permissions != nil {
		return permissions
	}
	permissions = toolPermissions()
	for _, n := range none {
		permissions[n] = user.NONE
	}
	for _, n := range out {
		permissions[n] = OUT
	}
	for _, n := range student {
		permissions[n] = user.STUDENT
	}
	for _, n := range teacher {
		permissions[n] = user.TEACHER
	}
	for _, n := range admin {
		permissions[n] = user.ADMIN
	}
	return permissions
}

//GenerateViews is used to load all the basic views used by our web app.
func GenerateViews(r *pat.Router, views map[string]string) {
	for n, v := range views {
		r.Add("GET", "/"+n, Handler(LoadView(n, v))).Name(n)
	}
}

//LoadView loads a view so that it is accessible in our web app.
func LoadView(n, v string) Handler {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		c.Browse.View = v
		if c.Browse.View == "home" {
			c.Browse.SetLevel(n)
		}
		return T(getNav(c), n).Execute(w, map[string]interface{}{"ctx": c})
	}
}

//CheckAccess verifies that a user is allowed access to a url.
func CheckAccess(p string, c *Context, ps map[string]user.Permission) error {
	//Retrieve the location they are requesting
	n := p
	if strings.HasPrefix(n, "/") {
		if len(n) > 1 {
			n = n[1:]
		} else {
			n = ""
		}
	}
	if i := strings.Index(n, "/"); i != -1 {
		n = n[:i]
	}
	if i := strings.Index(n, "?"); i != -1 {
		n = n[:i]
	}
	//Get the permission and check it.
	v, ok := ps[n]
	if !ok {
		return fmt.Errorf("could not find request %s", n)
	}
	if m := checkPermission(c, v); m != "" {
		return fmt.Errorf(m, p)
	}
	return nil
}

func checkPermission(c *Context, p user.Permission) string {
	//Check permission levels.
	switch p {
	case user.NONE:
	case OUT:
		if c.LoggedIn() {
			return "cannot access %s when logged in"
		}
	case user.STUDENT:
		if !c.LoggedIn() {
			return "you need to be logged in to access %s"
		}
	case user.ADMIN, user.TEACHER:
		u, e := c.Username()
		if e != nil {
			return "you need to be logged in to access %s"
		} else if !checkUserPermission(u, p) {
			return "you have insufficient permissions to access %s"
		}
	default:
		return "unknown url %s"
	}
	return ""
}

//checkUserPermission verifies that a user has the specified permission level.
func checkUserPermission(uname string, p user.Permission) bool {
	u, e := db.User(uname)
	return e == nil && u.Access >= p
}
