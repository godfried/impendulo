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

package tool

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"strings"
)

type (
	ChartVal struct {
		Name   string
		Y      int
		Show   bool
		FileId bson.ObjectId
	}

	//ChartResult is used to display result data in a chart.
	ChartResult interface {
		ChartVals() []*ChartVal
		GetName() string
	}

	//ToolResult is used to store tool result data.
	ToolResult interface {
		GetType() string
		//Retrieves the result's db id.
		GetId() bson.ObjectId
		//Retrieves the file associated with the result's db id.
		GetFileId() bson.ObjectId
		//Retrieves a summary of this result.
		Summary() *Summary
		//Retrieves this result's tool name.
		GetName() string
		//True if this result is partially stored on GridFS, false otherwise.
		OnGridFS() bool
		//Retrieves the report generated by the associated tool stored in this result.
		GetReport() Report
		//Sets this result's tool report. Used mainly to move data from/to GridFS
		SetReport(Report)
	}

	//DisplayResult is used to display result reports.
	DisplayResult interface {
		GetType() string
		//GetName
		GetName() string
		//GetReport
		GetReport() Report
		//Template retrieves the name of a html template.
		//Each DisplayResult should therefore have a html file which specify how
		//its result is displayed. The return value of the DisplayResult's
		//GetReport method is passed to the template.
		//Here is an example for Javac:
		//
		//{{define "result"}}
		//{{if .Success}}
		//<h4 class="text-success">{{.Header}}</h4>
		//{{else}}
		//{{$content := setBreaks .Result}}
		//{{if .Warnings}}
		//<h4 class="text-warning">{{.Header}}</h4>
		//<p class="text-warning">{{$content}}</p>
		//{{else}}
		//<h4 class="text-danger">{{.Header}}</h4>
		//<p class="text-danger">{{$content}}</p>
		//{{end}}
		//{{end}}
		//{{end}}
		Template() string
	}

	AdditionalResult interface {
		AdditionalTemplate() string
	}
	BugResult interface {
		Bug(id string, index int) (*Bug, error)
	}

	Bug struct {
		Id               string
		ResultId, FileId bson.ObjectId
		Title            string
		Content          []interface{}
		Lines            []int
	}

	//Report is an interface which represents a tool report on a snapshot.
	Report interface{}

	//ErrorResult is a DisplayResult used to indicate that an error
	//occured when retrieving a Tool's result or running a Tool..
	ErrorResult struct {
		err  error
		name string
	}

	//CodeResult is a DisplayResult used to display a source file's code.
	CodeResult struct {
		Lang string
		Data string
		Bug  *Bug
	}

	//SummaryResult is a DisplayResult used to provide a summary of all results.
	SummaryResult struct {
		summary []*Summary
	}

	//Summary is short summary of a ToolResult's result.
	Summary struct {
		//The tool's name.
		Name string
		//The text to be displayed in the summary.
		Body string
	}
	//CompileType tells us what type of result compilation gave us.
	CompileType uint
)

const (
	//Some result names.
	NORESULT = "NoResult"
	TIMEOUT  = "Timeout"
	SUMMARY  = "Summary"
	ERROR    = "Error"
	CODE     = "Code"
	//The different types of compilation that we can have.
	SUCCESS CompileType = iota
	WARNINGS
	ERRORS
)

var (
	COMPILE_SUCCESS = []byte("Compiled successfully")
)

//NewErrorResult creates an ErrorResult. There are 3 types:
//Timeout, No result and error.
func NewErrorResult(tipe, resultName string) *ErrorResult {
	var err error
	switch tipe {
	case TIMEOUT:
		err = fmt.Errorf("A timeout occured during execution of %s.", resultName)
	case NORESULT:
		err = fmt.Errorf("No result available for %s.", resultName)
	default:
		tipe = ERROR
		err = fmt.Errorf("%s: could not retrieve result for %s.", tipe, resultName)
	}
	return &ErrorResult{
		err:  err,
		name: tipe,
	}
}

//GetName
func (this *ErrorResult) GetName() string {
	return this.name
}

func (this *ErrorResult) GetType() string {
	return ERROR
}

//GetReport
func (this *ErrorResult) GetReport() Report {
	return this.err.Error()
}

func (this *ErrorResult) AdditionalTemplate() string {
	return "emptyadditionalresult"
}

//Template
func (this *ErrorResult) Template() string {
	return "emptyresult"
}

//NewCodeResult
func NewCodeResult(lang string, data []byte) *CodeResult {
	fmt.Println(strings.TrimSpace(string(data)))
	return &CodeResult{
		Lang: strings.ToLower(lang),
		Data: strings.TrimSpace(string(data)),
	}
}

func (this *CodeResult) GetType() string {
	return CODE
}

//GetName
func (this *CodeResult) GetName() string {
	return this.GetType()
}

//GetReport
func (this *CodeResult) GetReport() Report {
	return this
}

func (this *CodeResult) Template() string {
	return "coderesult"
}

//NewSummaryResult
func NewSummaryResult() *SummaryResult {
	return &SummaryResult{
		summary: make([]*Summary, 0),
	}
}

func (this *SummaryResult) GetType() string {
	return SUMMARY
}

//GetName
func (this *SummaryResult) GetName() string {
	return SUMMARY
}

//GetReport
func (this *SummaryResult) GetReport() Report {
	return this.summary
}

//Template
func (this *SummaryResult) Template() string {
	return "summaryresult"
}

//AddSummary adds a ToolResult's summary to this SummaryResult's
//list of summaries.
func (this *SummaryResult) AddSummary(result ToolResult) {
	this.summary = append(this.summary, result.Summary())
}

func NewBug(result ToolResult, id string, content []interface{}, start, end int) *Bug {
	lines := make([]int, end-start+1)
	for i := start; i <= end; i++ {
		lines[i-start] = i
	}
	return &Bug{
		Id:       id,
		ResultId: result.GetId(),
		FileId:   result.GetFileId(),
		Title:    result.GetName() + " Violation",
		Content:  content,
		Lines:    lines,
	}
}
