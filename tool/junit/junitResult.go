package junit

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"encoding/xml"
	"strings"
)

const NAME = "JUnit Test"

type JUnitResult struct {
	Id       bson.ObjectId "_id"
	FileId   bson.ObjectId "fileid"
	TestName string        "name"
	Time     int64         "time"
	Data     *TestSuite        "data"
}

func (this *JUnitResult) GetName() string {
	return this.TestName
}

func (this *JUnitResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *JUnitResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *JUnitResult) String() string {
	return "Type: tool.junit.JUnitResult; Id: " + this.Id.Hex() + "; FileId: " + this.FileId.Hex() + "; Time: " + util.Date(this.Time)
}

func (this *JUnitResult) TemplateArgs(current bool) (string, interface{}) {
	if current {
		return "junitCurrent", this.Data
	} else {
		return "junitNext", this.Data
	}
}

func (this *JUnitResult) Success() bool {
	return this.Data.Success
}

func NewResult(fileId bson.ObjectId, name string, data []byte) (res *JUnitResult, err error) {
	res = &JUnitResult{Id: bson.NewObjectId(), FileId: fileId, TestName: name, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}


type TestSuite struct {
	Id      bson.ObjectId
	Success bool
	Errors int       `xml:"errors,attr"`
	Failures int     `xml:"failures,attr"`
	Name   string `xml:"name,attr"`
	Tests  int       `xml:"tests,attr"`
	Time   float64   `xml:"time,attr"`
	Results []TestCase `xml:"testcase"`
}

type TestCase struct {
	ClassName      string `xml:"classname,attr"`
	Name string `xml:"name,attr"`
	Time   float64 `xml:"time,attr"`
	Fail Failure  `xml:"failure"`
}

type Failure struct {
	Message   string             `xml:"message,attr"`
	Type string            `xml:"type,attr"`
	Value  string `xml:",innerxml"`
}

func (this *Failure) IsFailure()bool{
	return len(strings.TrimSpace(this.Type)) > 0
}

func genReport(id bson.ObjectId, data []byte) (res *TestSuite, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		if res == nil{
			return
		} else{
			err = nil
		}
	}
	if res.Errors == 0 && res.Failures == 0 {
		res.Success = true
	} else {
		res.Success = false
	}
	res.Id = id
	return
}
