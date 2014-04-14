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

package pmd

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"

	"github.com/godfried/impendulo/tool"

	"html/template"

	"labix.org/v2/mgo/bson"

	"sort"
	"strings"
)

type (
	//Report is the result of running PMD on a Java source file.
	Report struct {
		Id      bson.ObjectId
		Version string  `xml:"version,attr"`
		Files   []*File `xml:"file"`
		Errors  int
	}

	//File defines a source file analysed by PMD and all of the errors
	//found in it.
	File struct {
		Name       string     `xml:"name,attr"`
		Violations Violations `xml:"violation"`
	}

	//Violations
	Violations []*Violation

	//Violation describes an error detected by PMD.
	Violation struct {
		Id          bson.ObjectId
		Begin       int          `xml:"beginline,attr"`
		End         int          `xml:"endline,attr"`
		Rule        string       `xml:"rule,attr"`
		RuleSet     string       `xml:"ruleset,attr"`
		Url         template.URL `xml:"externalInfoUrl,attr"`
		Priority    int          `xml:"priority,attr"`
		Description string       `xml:",innerxml"`
		//The locations where the error was detected.
		Starts, Ends []int
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport generates a new Report from XML generated by PMD.
func NewReport(id bson.ObjectId, data []byte) (*Report, error) {
	var r *Report
	if e := xml.Unmarshal(data, &r); e != nil {
		return nil, tool.NewXMLError(e, "pmd/pmdResult.go")
	}
	r.Id = id
	r.Errors = 0
	for _, f := range r.Files {
		r.Errors += len(f.Violations)
		f.CompressViolations()
	}
	return r, nil
}

//Success is true if no errors were found.
func (r *Report) Success() bool {
	return r.Errors == 0
}

//String
func (r *Report) String() string {
	s := fmt.Sprintf("Report{ Errors: %d\n.", r.Errors)
	if r.Files != nil {
		s += "Files: \n"
		for _, f := range r.Files {
			s += f.String()
		}
	}
	return s + "}\n"
}

//File retrieves a File whose name ends with the provided name.
func (r *Report) File(name string) *File {
	for _, f := range r.Files {
		if strings.HasSuffix(f.Name, name) {
			return f
		}
	}
	return nil
}

//CompressViolations packs all Violations of the same
//type into a single Violation by storing their location seperately.
func (f *File) CompressViolations() {
	indices := make(map[string]int)
	c := make(Violations, 0, len(f.Violations))
	for _, v := range f.Violations {
		i, ok := indices[v.Rule]
		if !ok {
			//Only store if the Violation type has not been stored yet.
			v.Starts = make([]int, 0, len(f.Violations))
			v.Ends = make([]int, 0, len(f.Violations))
			v.Id = bson.NewObjectId()
			c = append(c, v)
			i = len(c) - 1
			indices[v.Rule] = i
		}
		//Add Violation location.
		c[i].Starts = append(c[i].Starts, v.Begin)
		c[i].Ends = append(c[i].Ends, v.End)
	}
	sort.Sort(c)
	f.Violations = c
}

//String
func (f *File) String() string {
	return fmt.Sprintf("File{ Name: %s\n}\n.", f.Name)
}

func (v Violations) Len() int {
	return len(v)
}

func (v Violations) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Violations) Less(i, j int) bool {
	return v[i].Rule < v[j].Rule
}
