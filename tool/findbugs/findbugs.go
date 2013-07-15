package findbugs

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
)

type FindBugs struct {
	cmd string
}

func NewFindBugs() *FindBugs {
	return &FindBugs{config.GetConfig(config.FINDBUGS)}
}

func (this *FindBugs) GetLang() string {
	return "java"
}

func (this *FindBugs) GetName() string {
	return tool.FINDBUGS
}

func (this *FindBugs) args(target string) []string {
	return []string{config.GetConfig(config.JAVA), "-jar", this.cmd, "-textui", "-low", "-xml:withMessages", target}
}

func (this *FindBugs) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.Result, err error) {
	target := ti.GetTarget(tool.PKG_PATH)
	args := this.args(target)
	stdout, stderr, err := tool.RunCommand(args)
	if stdout != nil {
		//var fres *FindbugsResult
		res, err = NewResult(fileId, stdout)
		//res = fres
	} else if stderr != nil && len(stderr) > 0{
		err = fmt.Errorf("Could not run findbugs: %q.", string(stderr))
	}
	return
}

func (this *FindBugs) GenHTML() bool {
	return false
}
