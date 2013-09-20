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

package javac

import (
	"bytes"
	"encoding/gob"
	"labix.org/v2/mgo/bson"
	"strconv"
)

var (
	//Used to check the result of compilation.
	compSuccess  = []byte("Compiled successfully")
	compWarning  = []byte("warning")
	compWarnings = []byte("warnings")
	compError    = []byte("error")
	compErrors   = []byte("errors")
)

type (
	//Report contains the result of a Java compilation.
	Report struct {
		Id bson.ObjectId "_id"
		//Type stores the type of result:
		//Success, Warnings or Errors.
		Type CompileType "type"
		//Count is the number of errors or warnings encountered.
		Count int "count"
		//Data is what was generated by compilation.
		Data []byte "data"
	}
	//CompileType tells us what type of result compilation gave us.
	CompileType int
)

const (
	//The different types of compilation that we can have.
	SUCCESS CompileType = iota
	WARNINGS
	ERRORS
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) *Report {
	data = bytes.TrimSpace(data)
	return &Report{
		Id:    id,
		Type:  getType(data),
		Count: calcCount(data),
		Data:  data,
	}
}

//Errors tells us if there were errors during compilation.
func (this *Report) Errors() bool {
	return this.Type == ERRORS
}

//Success tells us if compilation finished with no errors or warnings.
func (this *Report) Success() bool {
	return this.Type == SUCCESS
}

//Warnings tells us if there were warnings during compilation.
func (this *Report) Warnings() bool {
	return this.Type == WARNINGS
}

//Header generates a string which briefly describes the compilation.
func (this *Report) Header() (header string) {
	if this.Success() {
		header = string(this.Data)
		return
	} else {
		header = strconv.Itoa(this.Count) + " "
		if this.Warnings() {
			header += "Warning"
		} else if this.Errors() {
			header += "Error"
		}
		if this.Count > 1 {
			header += "s"
		}
	}
	return
}

//calcCount extracts the number of errors or warnings which occurred
//during compilation.
func calcCount(data []byte) (n int) {
	split := bytes.Split(data, []byte("\n"))
	if len(split) < 1 {
		return
	}
	split = bytes.Split(bytes.TrimSpace(split[len(split)-1]), []byte(" "))
	if len(split) < 1 {
		return
	}
	n, _ = strconv.Atoi(string(split[0]))
	return
}

//getType extracts the type of compilation.
func getType(data []byte) (tipe CompileType) {
	if bytes.Equal(data, compSuccess) {
		tipe = SUCCESS
	} else if bytes.HasSuffix(data, compWarning) || bytes.HasSuffix(data, compWarnings) {
		tipe = WARNINGS
	} else if bytes.HasSuffix(data, compError) || bytes.HasSuffix(data, compErrors) {
		tipe = ERRORS
	}
	return
}
