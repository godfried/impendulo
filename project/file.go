//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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

type (
	//File stores a single file's data from a submission.
	File struct {
		Id      bson.ObjectId "_id"
		SubId   bson.ObjectId "subid"
		Name    string        "name"
		Package string        "package"
		Type    string        "type"
		Time    int64         "time"
		Data    []byte        "data"
		Results bson.M        "results"
	}
)

//String
func (this *File) String() string {
	return "Type: project.File; Id: " + this.Id.Hex() +
		"; SubId: " + this.SubId.Hex() + "; Name: " + this.Name +
		"; Package: " + this.Package + "; Type: " + this.Type +
		"; Time: " + util.Date(this.Time)
}

//Equals
func (this *File) Equals(that *File) bool {
	if reflect.DeepEqual(this, that) {
		return true
	}
	return that != nil &&
		this.String() == that.String() &&
		bytes.Equal(this.Data, that.Data)
}

//Same
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
	file.Type, err = util.GetString(info, TYPE)
	if err != nil && util.IsCastError(err) {
		return
	}
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
