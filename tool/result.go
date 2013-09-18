//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package tool

import (
	"labix.org/v2/mgo/bson"
	"strings"
)

const (
	//Some result names.
	NORESULT = "NoResult"
	TIMEOUT  = "Timeout"
	SUMMARY  = "Summary"
	ERROR    = "Error"
	CODE     = "Code"
)

type (
	//GraphResult is used to display result data in a graph.
	GraphResult interface {
		AddGraphData(curMax, x float64, graphData []map[string]interface{}) (newMax float64)
		GetName() string
	}
	//ToolResult is used to store tool result data.
	ToolResult interface {
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
		//Retrieves the actual data generated by the associated tool stored in this result.
		GetData() interface{}
		//Sets this result's tool data. Used mainly to move data from/to GridFS
		SetData(interface{})
	}

	//DisplayResult is used to display result data.
	DisplayResult interface {
		GetName() string
		GetData() interface{}
	}
	//NoResult is a DisplayResult used to indicate that a
	//Tool provided no result when run.
	NoResult struct{}

	//TimeoutResult is a DisplayResult used to indicate that a
	//Tool timed out when running.
	TimeoutResult struct{}

	//ErrorResult is a DisplayResult used to indicate that an error
	//occured when retrieving a Tool's result.
	ErrorResult struct {
		err error
	}

	//CodeResult is a DisplayResult used to display a source file's code.
	CodeResult struct {
		data string
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
)

//AddCoords inserts new coordinates into data used to display a chart.
func AddCoords(chart map[string]interface{}, x, y float64) {
	if _, ok := chart["data"].([]map[string]float64); !ok {
		return
	}
	if x < 0 {
		x = float64(len(chart["data"].([]map[string]float64)))
	} else {
		x = x / 1000
	}
	chart["data"] = append(chart["data"].([]map[string]float64),
		map[string]float64{"x": x, "y": y})
}

//CreateChart initialises new chart data.
func CreateChart(name string) (chart map[string]interface{}) {
	chart = make(map[string]interface{})
	chart["name"] = name
	chart["data"] = make([]map[string]float64, 0, 100)
	return
}

//Template retrieves a html template for a DisplayResult.
//Each DisplayResult should therefore have two html files which specify how
//its result is displayed. These are called (DisplayResult name)Current.html and
//(DisplayResult name)Next.html. They are almost identical in structure but the one
//will just use .curResult and the other .nextResult when displaying data.
//This is an example (javacNext.html):
//
//{{define "nextResult"}}
//{{$result := .nextResult}}
//{{$header := $result.ResultHeader}}
//{{if $result.Success}}
//<h4 class="text-success">{{$header}}</h4>
//{{else}}
//{{$res := setBreaks $result.Result}}
//{{if $result.Warnings}}
//<h4 class="text-warning">{{$header}}</h4>
//<p class="text-warning">{{$res}}</p>
//{{else}}
//<h4 class="text-danger">{{$header}}</h4>
//<p class="text-danger">{{$res}}</p>
//{{end}}
//{{end}}
//{{end}}
func Template(name string, current bool) string {
	name = strings.ToLower(name)
	if current {
		return name + "Current"
	} else {
		return name + "Next"
	}
}

//GetName
func (this *NoResult) GetName() string {
	return NORESULT
}

//GetData
func (this *NoResult) GetData() interface{} {
	return "No output generated."
}

//GetName
func (this *TimeoutResult) GetName() string {
	return TIMEOUT
}

//GetData
func (this *TimeoutResult) GetData() interface{} {
	return "A timeout occured during execution."
}

//NewErrorResult
func NewErrorResult(err error) *ErrorResult {
	return &ErrorResult{
		err: err,
	}
}

//GetName
func (this *ErrorResult) GetName() string {
	return ERROR
}

//GetData
func (this *ErrorResult) GetData() interface{} {
	return this.err.Error()
}

//NewCodeResult
func NewCodeResult(data []byte) *CodeResult {
	return &CodeResult{
		data: strings.TrimSpace(string(data)),
	}
}

//GetName
func (this *CodeResult) GetName() string {
	return CODE
}

//GetData
func (this *CodeResult) GetData() interface{} {
	return this.data
}

//NewSummaryResult
func NewSummaryResult() *SummaryResult {
	return &SummaryResult{
		summary: make([]*Summary, 0),
	}
}

//GetName
func (this *SummaryResult) GetName() string {
	return SUMMARY
}

//GetData
func (this *SummaryResult) GetData() interface{} {
	return this.summary
}

//AddSummary adds a ToolResult's summary to this SummaryResult's
//list of summaries.
func (this *SummaryResult) AddSummary(result ToolResult) {
	this.summary = append(this.summary, result.Summary())
}
