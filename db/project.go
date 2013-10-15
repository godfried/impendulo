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

package db

import (
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type (
	//FileInfo
	FileInfo struct {
		Name  string
		Count int
	}
)

//File retrieves a file matching the given interface from the active database.
func File(matcher, selector interface{}) (ret *project.File, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(FILES)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"file", err, matcher}
	}
	return
}

//Files retrieves files matching the given interface from the active database.
func Files(matcher, selector interface{}, sort ...string) (ret []*project.File, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(FILES)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"files", err, matcher}
	}
	return
}

//FileInfos retrieves names of file information.
func FileInfos(matcher bson.M) (ret []*FileInfo, err error) {
	names, err := FileNames(matcher)
	if err != nil {
		return
	}
	ret = make([]*FileInfo, len(names))
	for i, name := range names {
		ret[i] = new(FileInfo)
		ret[i].Name = name
		matcher[project.NAME] = name
		ret[i].Count, err = Count(FILES, matcher)
		if err != nil {
			return
		}
	}
	return
}

//FileNames retrieves names of files
//matching the given interface from the active database.
func FileNames(matcher interface{}) (ret []string, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(FILES)
	err = c.Find(matcher).Distinct(project.NAME, &ret)
	if err != nil {
		err = &DBGetError{"filenames", err, matcher}
	}
	return
}

//Submission retrieves a submission matching
//the given interface from the active database.
func Submission(matcher, selector interface{}) (ret *project.Submission, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"submission", err, matcher}
	} else if ret.Status == project.UNKNOWN {
		err = UpdateStatus(ret)
	}
	return
}

//Submissions retrieves submissions matching
//the given interface from the active database.
func Submissions(matcher, selector interface{}, sort ...string) (ret []*project.Submission, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(SUBMISSIONS)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"submissions", err, matcher}
	}
	return
}

//Project retrieves a project matching
//the given interface from the active database.
func Project(matcher, selector interface{}) (ret *project.Project, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"project", err, matcher}
	}
	return
}

//Projects retrieves projects matching
//the given interface from the active database.
func Projects(matcher, selector interface{}, sort ...string) (ret []*project.Project, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(PROJECTS)
	q := c.Find(matcher)
	if len(sort) > 0 {
		q = q.Sort(sort...)
	}
	err = q.Select(selector).All(&ret)
	if err != nil {
		err = &DBGetError{"projects", err, matcher}
	}
	return
}

//RemoveFileById removes a file matching the given id from the active database.
func RemoveFileById(id interface{}) (err error) {
	file, err := File(bson.M{project.ID: id}, bson.M{project.RESULTS: 1})
	if err != nil {
		return
	}
	for _, resId := range file.Results {
		if _, ok := resId.(bson.ObjectId); !ok {
			continue
		}
		err = RemoveById(RESULTS, resId)
		if err != nil {
			return
		}
	}
	err = RemoveById(FILES, id)
	return
}

//RemoveSubmissionById removes a submission matching
//the given id from the active database.
func RemoveSubmissionById(id interface{}) (err error) {
	files, err := Files(bson.M{project.SUBID: id},
		bson.M{project.ID: 1})
	if err != nil {
		return
	}
	for _, file := range files {
		err = RemoveFileById(file.Id)
		if err != nil {
			return
		}
	}
	err = RemoveById(SUBMISSIONS, id)
	return
}

//RemoveProjectById removes a project matching
//the given id from the active database.
func RemoveProjectById(id interface{}) (err error) {
	projectMatch := bson.M{project.PROJECT_ID: id}
	idSelect := bson.M{project.ID: 1}
	subs, err := Submissions(projectMatch, idSelect)
	if err != nil {
		return
	}
	for _, sub := range subs {
		err = RemoveSubmissionById(sub.Id)
		if err != nil {
			return
		}
	}
	tests, err := JUnitTests(projectMatch, idSelect)
	if err != nil {
		return
	}
	for _, test := range tests {
		RemoveById(TESTS, test.Id)
	}
	jpfConfig, err := JPFConfig(projectMatch, idSelect)
	if err == nil {
		RemoveById(JPF, jpfConfig.Id)
	}
	pmdRules, err := PMDRules(projectMatch, idSelect)
	if err == nil {
		RemoveById(PMD, pmdRules.Id)
	}
	err = RemoveById(PROJECTS, id)
	return
}

func LastFile(sub *project.Submission) (file *project.File, err error) {
	matcher := bson.M{project.TYPE: project.SRC, project.SUBID: sub.Id}
	selector := bson.M{project.RESULTS: 1, project.TIME: 1}
	files, err := Files(matcher, selector, "-"+project.TIME)
	if err != nil {
		return
	}
	if len(files) == 0 {
		if sub.Status != project.BUSY {
			RemoveSubmissionById(sub.Id)
		}
		return
	}
	file = files[0]
	return
}

func UpdateStatus(sub *project.Submission) (err error) {
	file, err := LastFile(sub)
	if err != nil || file == nil {
		return
	}
	junitStatus := CheckJUnit(sub.ProjectId, file)
	jpfStatus := CheckJPF(sub.ProjectId, file)
	switch junitStatus {
	case project.JUNIT:
		switch jpfStatus {
		case project.UNKNOWN:
			sub.Status = project.JUNIT
		case project.JPF:
			sub.Status = project.ALL
		case project.NOTJPF:
			sub.Status = project.JUNIT_NOTJPF
		}
	case project.NOTJUNIT:
		switch jpfStatus {
		case project.UNKNOWN:
			sub.Status = project.NOTJUNIT
		case project.JPF:
			sub.Status = project.JPF_NOTJUNIT
		case project.NOTJPF:
			sub.Status = project.FAILED
		}
	case project.UNKNOWN:
		switch jpfStatus {
		case project.UNKNOWN:
			sub.Status = project.UNKNOWN
		case project.JPF:
			sub.Status = project.JPF
		case project.NOTJPF:
			sub.Status = project.NOTJPF
		}
	}
	change := bson.M{SET: bson.M{project.STATUS: sub.Status}}
	err = Update(SUBMISSIONS, bson.M{project.ID: sub.Id}, change)
	return
}

func CheckJUnit(projectId bson.ObjectId, file *project.File) project.Status {
	tests, err := JUnitTests(bson.M{project.PROJECT_ID: projectId}, bson.M{project.NAME: 1})
	if err != nil || len(tests) == 0 {
		return project.UNKNOWN
	}
	for _, test := range tests {
		name := tool.NewTarget(test.Name, "", "", "").Name
		id, ok := file.Results[name].(bson.ObjectId)
		if !ok {
			return project.NOTJUNIT
		}
		testResult, terr := JUnitResult(bson.M{project.ID: id}, nil)
		if terr != nil || !testResult.Success() {
			return project.NOTJUNIT
		}
	}
	return project.JUNIT
}

func CheckJPF(projectId bson.ObjectId, file *project.File) project.Status {
	_, err := JPFConfig(bson.M{project.PROJECT_ID: projectId}, bson.M{project.ID: 1})
	if err != nil {
		return project.UNKNOWN
	}
	res, jerr := JPFResult(bson.M{project.FILEID: file.Id}, nil)
	if jerr == nil && res.Success() {
		return project.JPF
	}
	return project.NOTJPF
}
