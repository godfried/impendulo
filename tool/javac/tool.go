package javac

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

//Javac is a tool.Tool used to compile Java source files.
type Tool struct {
	cmd string
	cp  string
}

//New creates a new Javac instance. cp is the classpath used when compiling.
func New(cp string) *Tool {
	return &Tool{
		cmd: config.GetConfig(config.JAVAC),
		cp:  cp,
	}
}

func (this *Tool) GetLang() string {
	return "java"
}

func (this *Tool) GetName() string {
	return NAME
}

func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	args := []string{this.cmd, "-cp", this.cp + ":" + ti.Dir,
		"-implicit:class", "-Xlint", ti.FilePath()}
	//Compile the file.
	execRes := tool.RunCommand(args, nil)
	if execRes.Err != nil {
		if !tool.IsEndError(execRes.Err) {
			err = execRes.Err
		} else {
			//Unsuccessfull compile.
			res = NewResult(fileId, execRes.StdErr)
			err = &CompileError{ti.FullName(), string(execRes.StdErr)}
		}
	} else if execRes.HasStdErr() {
		//Compiler warnings.
		res = NewResult(fileId, execRes.StdErr)
	} else {
		res = NewResult(fileId, compSuccess)
	}
	return
}

//CompileError is used to indicate that compilation failed.
type CompileError struct {
	name string
	msg  string
}

func (this *CompileError) Error() string {
	return fmt.Sprintf("Could not compile %q due to: %q.", this.name, this.msg)
}

//IsCompileError checks whether an error is a CompileError.
func IsCompileError(err error) (ok bool) {
	_, ok = err.(*CompileError)
	return
}
