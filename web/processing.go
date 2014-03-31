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
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"

	"io/ioutil"

	"labix.org/v2/mgo/bson"

	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

//ReadFormFile reads a file's name and data from a request form.
func ReadFormFile(req *http.Request, name string) (fname string, data []byte, err error) {
	file, header, err := req.FormFile(name)
	if err != nil {
		return
	}
	fname = header.Filename
	data, err = ioutil.ReadAll(file)
	return
}

//GetInt retrieves an integer value from a request form.
func GetInt(req *http.Request, name string) (found int, err error) {
	iStr := req.FormValue(name)
	found, err = strconv.Atoi(iStr)
	return
}

//GetInt64 retrieves an integer value from a request form.
func GetInt64(req *http.Request, name string) (found int64, err error) {
	iStr := req.FormValue(name)
	found, err = strconv.ParseInt(iStr, 10, 64)
	return
}

//GetStrings retrieves a string value from a request form.
func GetStrings(req *http.Request, name string) (vals []string, err error) {
	if req.Form == nil {
		err = req.ParseForm()
		if err != nil {
			return
		}
	}
	vals = req.Form[name]
	return
}

//GetString retrieves a string value from a request form.
func GetString(req *http.Request, name string) (val string, err error) {
	val = req.FormValue(name)
	if strings.TrimSpace(val) == "" {
		err = fmt.Errorf("Invalid value for %s.", name)
	}
	return
}

//getIndex
func getIndex(req *http.Request, name string, maxSize int) (ret int, err error) {
	ret, err = GetInt(req, name)
	if err != nil {
		return
	}
	if ret > maxSize {
		ret = 0
	} else if ret < 0 {
		ret = maxSize
	}
	return
}

//getCurrent
func getCurrent(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "current", maxSize)
}

//getNext
func getNext(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "next", maxSize)
}

//getProjectId
func getProjectId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "projectid", "project")
}

//getSubId
func getSubId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "subid", "submission")
}

//getSkeletonId
func getSkeletonId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "skeletonid", "skeleton")
}

//getFileId
func getFileId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "fileid", "file")
}

func getId(req *http.Request, ident, name string) (id bson.ObjectId, msg string, err error) {
	id, err = util.ReadId(req.FormValue(ident))
	if err != nil {
		msg = fmt.Sprintf("Could not read %s.", name)
	}
	return
}

//getActiveUser
func getActiveUser(ctx *Context) (user, msg string, err error) {
	user, err = ctx.Username()
	if err != nil {
		msg = "Could not retrieve user."
	}
	return
}

//getUserId
func getUserId(req *http.Request) (userid, msg string, err error) {
	userid, err = GetString(req, "userid")
	if err != nil {
		msg = "Could not read user."
	}
	return
}

//getString.
func getString(req *http.Request, name string) (val, msg string, err error) {
	val, err = GetString(req, name)
	if err != nil {
		msg = fmt.Sprintf("Could not read %s.", name)
	}
	return
}

func getTestType(req *http.Request) (tipe junit.Type, msg string, err error) {
	val := strings.ToLower(req.FormValue("testtype"))
	switch val {
	case "default":
		tipe = junit.DEFAULT
	case "admin":
		tipe = junit.ADMIN
	case "user":
		tipe = junit.USER
	default:
		err = fmt.Errorf("Unknown test type %s.", val)
		msg = err.Error()
	}
	return
}

func ServePath(u *url.URL, src string) (servePath string, err error) {
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	cleaned := path.Clean(u.Path)
	ext, err := filepath.Rel("/"+filepath.Base(src), cleaned)
	if err != nil {
		return
	}
	servePath = filepath.Join(src, ext)
	if util.IsDir(servePath) && !strings.HasSuffix(u.Path, "/") {
		u.Path = u.Path + "/"
	}
	return
}

//getCredentials
func getCredentials(req *http.Request) (uname, pword, msg string, err error) {
	uname, msg, err = getString(req, "username")
	if err != nil {
		return
	}
	pword, msg, err = getString(req, "password")
	return
}

//Snapshots retrieves snapshots of a given file in a submission.
func Snapshots(subId bson.ObjectId, fileName string, tipe project.Type) (ret []*project.File, err error) {
	matcher := bson.M{db.SUBID: subId, db.NAME: fileName, db.TYPE: tipe}
	selector := bson.M{db.DATA: 0}
	ret, err = db.Files(matcher, selector, db.TIME)
	if err == nil && len(ret) == 0 {
		err = fmt.Errorf("No files found with name %q.", fileName)
	}
	return
}

//getFile
func getFile(id bson.ObjectId) (file *project.File, err error) {
	selector := bson.M{db.NAME: 1, db.TIME: 1}
	file, err = db.File(bson.M{db.ID: id}, selector)
	return
}

//projectName retrieves the project name associated with the project identified by id.
func projectName(idval interface{}) (name string, err error) {
	id, err := util.ReadId(idval)
	if err != nil {
		return
	}
	matcher := bson.M{db.ID: id}
	selector := bson.M{db.NAME: 1}
	proj, err := db.Project(matcher, selector)
	if err != nil {
		return
	}
	name = proj.Name
	return
}

func testedFileName(subId bson.ObjectId) (name string, err error) {
	tests, err := db.Files(bson.M{db.SUBID: subId, db.TYPE: project.TEST}, bson.M{db.DATA: 0})
	if err != nil {
		return
	}
	for _, test := range tests {
		for idstr, _ := range test.Results {
			s := strings.Split(idstr, "-")
			if len(s) < 2 {
				continue
			}
			id, ierr := util.ReadId(s[len(s)-1])
			if ierr != nil {
				continue
			}
			f, derr := db.File(bson.M{db.ID: id}, bson.M{db.NAME: 1})
			if derr != nil {
				continue
			}
			name = f.Name
			return
		}
	}
	err = fmt.Errorf("No file name found for submission %s", subId.Hex())
	return
}

func users(projectId bson.ObjectId) (names []string, err error) {
	submissions, err := db.Submissions(bson.M{db.PROJECTID: projectId}, nil)
	if err != nil {
		return
	}
	names = make([]string, 0, len(submissions))
	added := make(map[string]bool)
	for _, s := range submissions {
		if added[s.User] {
			continue
		}
		names = append(names, s.User)
		added[s.User] = true
	}
	return
}
