package project

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//File stores a single file's data from a submission.
type File struct {
	Id      bson.ObjectId "_id"
	SubId   bson.ObjectId "subid"
	Name    string        "name"
	Package string        "package"
	Type    string        "type"
	Time    int64         "time"
	Data    []byte        "data"
	Results bson.M        "results"
}

func (this *File) TypeName() string {
	return "file"
}

func (this *File) String() string {
	return "Type: project.File; Id: " + this.Id.Hex() +
		"; SubId: " + this.SubId.Hex() + "; Name: " + this.Name +
		"; Package: " + this.Package + "; Type: " + this.Type +
		"; Time: " + util.Date(this.Time)
}

func (this *File) Equals(that *File) bool {
	if reflect.DeepEqual(this, that) {
		return true
	}
	return that != nil &&
		this.String() == that.String() &&
		bytes.Equal(this.Data, that.Data)
}

func (this *File) Same(that *File) bool {
	return this.Id == that.Id
}

//CanProcess returns whether a file is meant to be processed.
func (this *File) CanProcess() bool {
	return this.Type == SRC || this.Type == ARCHIVE
}

//NewFile
func NewFile(subId bson.ObjectId, info map[string]interface{}, data []byte) (file *File, err error) {
	id := bson.NewObjectId()
	file = &File{Id: id, SubId: subId, Data: data}
	//Non essential fields
	file.Type, err = util.GetString(info, TYPE)
	if err != nil && util.IsCastError(err) {
		return
	}
	//Essential fields
	file.Name, err = util.GetString(info, NAME)
	if err != nil {
		return
	}
	file.Package, err = util.GetString(info, PKG)
	if err != nil {
		return
	}
	file.Time, err = util.GetInt64(info, TIME)
	return
}

//NewArchive
func NewArchive(subId bson.ObjectId, data []byte) *File {
	id := bson.NewObjectId()
	return &File{
		Id:    id,
		SubId: subId,
		Data:  data,
		Type:  ARCHIVE,
	}
}

//ParseName retrieves file metadata encoded in a file name.
//These file names must have the format:
//[[<package descriptor>"_"]*<file name>"_"]<time in nanoseconds>
//"_"<file number in current submission>"_"<modification char>
//Where values between '[]' are optional, '*' indicates 0 to many,
//values inside '""' are literals and values inside '<>'
//describe the contents at that position.
func ParseName(name string) (file *File, err error) {
	elems := strings.Split(name, "_")
	if len(elems) < 3 {
		err = fmt.Errorf("Encoded name %q does not have enough parameters.", name)
		return
	}
	file = new(File)
	file.Id = bson.NewObjectId()
	mod := elems[len(elems)-1]
	nextIndex := 3
	if len(elems[len(elems)-2]) > 10 {
		nextIndex = 2
	}
	timeString := elems[len(elems)-nextIndex]
	if len(timeString) == 13 {
		file.Time, err = strconv.ParseInt(timeString, 10, 64)
		if err != nil {
			err = fmt.Errorf(
				"%s in name %s could not be parsed as an int.",
				timeString, name)
			return
		}
	} else if timeString[0] == '2' && len(timeString) == 17 {
		var t time.Time
		t, err = util.CalcTime(timeString)
		if err != nil {
			return
		}
		file.Time = util.GetMilis(t)
	} else {
		err = fmt.Errorf(
			"Unknown time format %s in %s.",
			timeString, name)
		return
	}
	if len(elems) > nextIndex {
		nextIndex++
		pos := len(elems) - nextIndex
		file.Name = elems[pos]
		for i := 0; i < pos; i++ {
			file.Package += elems[i]
			if i < pos-1 {
				file.Package += "."
			}
			if isOutFolder(elems[i]) {
				file.Package = ""
			}
		}
	}
	if strings.HasSuffix(file.Name, JSRC) {
		file.Type = SRC
	} else if mod == "l" {
		file.Type = LAUNCH
	} else {
		err = fmt.Errorf("Unsupported file type in name %s", name)
	}
	return
}

//isOutFolder
func isOutFolder(arg string) bool {
	return arg == SRC_DIR || arg == BIN_DIR
}
