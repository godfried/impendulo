package tool

import (
	"bytes"
	"fmt"
	"labix.org/v2/mgo/bson"
	"os/exec"
	"io"
)

const (
	JUNIT    = "junit"
	JAVAC    = "javac"
	FINDBUGS = "findbugs"
	LINT4J   = "lint4j"
	JPF      = "jpf"
)

type Tool interface {
	GetName() string
	GetLang() string
	Run(fileId bson.ObjectId, target *TargetInfo) (Result, error)
}

//RunCommand executes a given external command.
func RunCommand(args []string) ([]byte, []byte, error) {
	return RunInputCommand(args, nil)
}

func RunInputCommand(args []string, stdin io.Reader) ([]byte, []byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("Encountered error %q executing command %q", err, args)
	}
	err = cmd.Wait()
	if err != nil {
		err = fmt.Errorf("Encountered error %q executing command %q", err, args)
	}
	return stdout.Bytes(), stderr.Bytes(), err
}