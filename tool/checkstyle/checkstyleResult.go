package checkstyle

import (
	"fmt"
	"encoding/xml"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/tool"
	"html/template"
	"labix.org/v2/mgo/bson"
)

const NAME = "Checkstyle"

type CheckstyleResult struct {
	Id     bson.ObjectId     "_id"
	FileId bson.ObjectId     "fileid"
	Time   int64             "time"
	Data   *CheckstyleReport "data"
}

func (this *CheckstyleResult) GetName() string {
	return NAME
}

func (this *CheckstyleResult) GetSummary() *tool.Summary {
		body := fmt.Sprintf("Errors: %d", 
		this.Data.Errors)
	return &tool.Summary{this.GetName(), body}
}

func (this *CheckstyleResult) GetId() bson.ObjectId {
	return this.Id
}

func (this *CheckstyleResult) GetFileId() bson.ObjectId {
	return this.FileId
}

func (this *CheckstyleResult) GetData() interface{} {
	return this.Data
}

func (this *CheckstyleResult) Template(current bool) string{
	if current {
		return "checkstyleCurrent"
	} else {
		return "checkstyleNext"
	}
}

func (this *CheckstyleResult) Success() bool {
	return true
}

func NewResult(fileId bson.ObjectId, data []byte) (res *CheckstyleResult, err error) {
	res = &CheckstyleResult{Id: bson.NewObjectId(), FileId: fileId, Time: util.CurMilis()}
	res.Data, err = genReport(res.Id, data)
	return
}

func genReport(id bson.ObjectId, data []byte) (res *CheckstyleReport, err error) {
	if err = xml.Unmarshal(data, &res); err != nil {
		return
	}
	res.Id = id
	res.Errors = 0
	for _, f := range res.Files{
		res.Errors += len(f.Errors)
	}
	return
}

type CheckstyleReport struct {
	Id      bson.ObjectId
	Version string  `xml:"version,attr"`
	Errors int
	Files   []*File `xml:"file"`
}
type File struct {
	Name   string   `xml:"name,attr"`
	Errors []*Error `xml:"error"`
}

type Error struct {
	Line     int           `xml:"line,attr"`
	Column   int           `xml:"column,attr"`
	Severity string        `xml:"severity,attr"`
	Message  template.HTML `xml:"message,attr"`
	Source   string        `xml:"source,attr"`
}
