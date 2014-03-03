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

//Package javac is the OpenJDK Java compiler's implementation of an Impendulo tool.
//For more information see http://openjdk.java.net/groups/compiler/.
package javac

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type (
	//Javac is a tool.Tool used to compile Java source files.
	Tool struct {
		cmd string
		cp  string
	}
)

//New creates a new javac instance. cp is the classpath used when compiling.
func New(cp string) (tool *Tool, err error) {
	tool = &Tool{
		cp: cp,
	}
	tool.cmd, err = config.JAVAC.Path()
	return
}

//Lang is Java.
func (this *Tool) Lang() tool.Language {
	return tool.JAVA
}

//Name is Javac
func (this *Tool) Name() string {
	return NAME
}

func (this *Tool) AddCP(add string) {
	if this.cp != "" {
		this.cp += ":"
	}
	this.cp += add
}

//Run compiles the Java source file specified by t. We compile with maximum warnings and compile
//classes implicitly loaded by the source code. All compilation results will be stored (success,
//errors and warnings).
func (this *Tool) Run(fileId bson.ObjectId, t *tool.Target) (res tool.ToolResult, err error) {
	cp := this.cp
	if cp != "" {
		cp += ":"
	}
	cp += t.Dir
	args := []string{this.cmd, "-cp", cp + ":" + t.Dir,
		"-implicit:class", "-Xlint", t.FilePath()}
	//Compile the file.
	execRes := tool.RunCommand(args, nil)
	if execRes.Err != nil {
		if !tool.IsEndError(execRes.Err) {
			err = execRes.Err
		} else {
			//Unsuccessfull compile.
			res = NewResult(fileId, execRes.StdErr)
			err = tool.NewCompileError(t.FullName(), string(execRes.StdErr))
		}
	} else if execRes.HasStdErr() {
		//Compiler warnings.
		res = NewResult(fileId, execRes.StdErr)
	} else {
		res = NewResult(fileId, tool.COMPILE_SUCCESS)
	}
	return
}
