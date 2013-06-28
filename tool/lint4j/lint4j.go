package lint4j

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type Lint4j struct {
	cmd string
}

func NewLint4j() *Lint4j {
	return &Lint4j{config.GetConfig(config.FINDBUGS)}
}

func (this *Lint4j) GetLang() string {
	return "java"
}

func (this *Lint4j) GetName() string {
	return tool.LINT4J
}

func (this *Lint4j) args(target string) []string {
	return []string{this.cmd, "-textui", "-low", target}
}

func (this *Lint4j) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (*tool.Result, error) {
	target := ti.GetTarget(tool.PKG_PATH)
	args := this.args(target)
	stderr, stdout, err := tool.RunCommand(args...)
	if err != nil {
		return nil, err
	}
	return tool.NewResult(fileId, this, stdout, stderr, err), nil
}
