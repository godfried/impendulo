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

package findbugs

import (
	"fmt"

	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

const (
	NAME = "Findbugs"
)

type (
	//Result is a tool.ToolResult and a tool.DisplayResult.
	//It is used to store the output of running Findbugs.
	Result struct {
		Id     bson.ObjectId "_id"
		FileId bson.ObjectId "fileid"
		Name   string        "name"
		Report *Report       "report"
		GridFS bool          "gridfs"
		Type   string        "type"
	}
)

//SetReport is used to change this result's report. This comes in handy
//when putting data into/getting data out of GridFS
func (r *Result) SetReport(report tool.Report) {
	if report == nil {
		r.Report = nil
	} else {
		r.Report = report.(*Report)
	}
}

//OnGridFS
func (r *Result) OnGridFS() bool {
	return r.GridFS
}

//String
func (r *Result) String() string {
	return fmt.Sprintf("Id: %q; FileId: %q; TestName: %s; \n Report: %s", r.Id, r.FileId, r.Name, r.Report)
}

//GetName
func (r *Result) GetName() string {
	return r.Name
}

//GetId
func (r *Result) GetId() bson.ObjectId {
	return r.Id
}

//GetFileId
func (r *Result) GetFileId() bson.ObjectId {
	return r.FileId
}

//Summary
func (r *Result) Summary() *tool.Summary {
	return &tool.Summary{
		Name: r.GetName(),
		Body: fmt.Sprintf("Bugs: %d", r.Report.Summary.BugCount),
	}
}

//GetReport
func (r *Result) GetReport() tool.Report {
	return r.Report
}

//ChartVals
func (r *Result) ChartVals() []*tool.ChartVal {
	return []*tool.ChartVal{
		&tool.ChartVal{"All", float64(r.Report.Summary.BugCount), r.FileId},
		&tool.ChartVal{"Priority 1", float64(r.Report.Summary.Priority1), r.FileId},
		&tool.ChartVal{"Priority 2", float64(r.Report.Summary.Priority2), r.FileId},
		&tool.ChartVal{"Priority 3", float64(r.Report.Summary.Priority3), r.FileId},
	}
}

func (r *Result) Template() string {
	return "findbugsresult"
}

func (r *Result) GetType() string {
	return r.Type
}

//NewResult
func NewResult(fileId bson.ObjectId, data []byte) (*Result, error) {
	id := bson.NewObjectId()
	r, e := NewReport(id, data)
	if e != nil {
		return nil, e
	}
	return &Result{
		Id:     id,
		FileId: fileId,
		Name:   NAME,
		GridFS: len(data) > tool.MAX_SIZE,
		Type:   NAME,
		Report: r,
	}, nil
}
