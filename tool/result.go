package tool

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"github.com/godfried/impendulo/util"
)

//Result describes a tool or test's results for a given file.
type Result struct {
	Id     bson.ObjectId "_id"
	FileId bson.ObjectId "fileid"
	Name   string        "name"
	HTML   bool          "html"
	Time   int64         "time"
	Data   []byte        "data"
}

func (this *Result) Equals(that *Result) bool {
	return reflect.DeepEqual(this, that)
}

func (this *Result) String() string {
	return "File: " + this.FileId.String() + "; Name: " + this.Name + "; Output:" + this.Output()
}

func (this *Result) Output() string {
	return string(this.Data)
}

//NewResult
func NewResult(fileId bson.ObjectId, tool Tool, data []byte) *Result {
	return &Result{bson.NewObjectId(), fileId, tool.GetName(), tool.GenHTML(), util.CurMilis(), data}
}
