package tool

import (
	"bytes"
	"fmt"
	"labix.org/v2/mgo/bson"
	"os/exec"
	"reflect"
)

const (
	JUNIT    = "junit"
	JAVAC    = "javac"
	FINDBUGS = "findbugs"
	LINT4J   = "lint4j"
	JPF   = "jpf"
)

type Tool interface {
	GenHTML() bool
	GetName() string
	GetLang() string
	Run(fileId bson.ObjectId, target *TargetInfo) (*Result, error)
}

//Tool is a generic tool specification.
type GenericTool struct {
	name     string
	lang     string
	exec     string
	preamble []string
	flags    []string
	args     map[string]string
	target   TargetSpec
}

//GetArgs sets up tool arguments for execution.
func (this *GenericTool) GetArgs(target string) (args []string) {
	args = make([]string, len(this.preamble)+len(this.flags)+(len(this.args)*2)+2)
	for i, p := range this.preamble {
		args[i] = p
	}
	args[len(this.preamble)] = this.exec
	start := len(this.preamble) + 1
	stop := start + len(this.flags)
	for j := start; j < stop; j++ {
		args[j] = this.flags[j-start]
	}
	cur := stop
	stop += len(this.args) * 2
	for k, v := range this.args {
		args[cur] = k
		args[cur+1] = v
		cur += 2
	}
	args[stop] = target
	return args
}

func (this *GenericTool) AddArgs(args map[string]string) {
	this.args = args
}

func (this *GenericTool) GenHTML() bool {
	return false
}

func (this *GenericTool) Equals(that Tool) bool {
	return reflect.DeepEqual(this, that)
}

func (this *GenericTool) Run(fileId bson.ObjectId, ti *TargetInfo) (*Result, error) {
	target := ti.GetTarget(this.target)
	args := this.GetArgs(target)
	stdout, stderr, err := RunCommand(args...)
	if err != nil {
		return nil, err
	}
	if stderr != nil && len(stderr) > 0 {
		return NewResult(fileId, this, stderr), nil
	}
	return NewResult(fileId, this, stdout), nil
}

func (this *GenericTool) GetName() string {
	return this.name
}

func (this *GenericTool) GetLang() string {
	return this.lang
}

//RunCommand executes a given external command.
func RunCommand(args ...string) ([]byte, []byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("Encountered error %q executing command %q", err, args)
	}
	err = cmd.Wait()
	if err != nil{
		return nil, nil, fmt.Errorf("Encountered error %q executing command %q", err, args)
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}
