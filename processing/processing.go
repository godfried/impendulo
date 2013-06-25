package processing

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing/monitor"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/java"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

var fileChan chan *project.File
var subChan chan *project.Submission

func init() {
	fileChan = make(chan *project.File)
	subChan = make(chan *project.Submission)
}

func AddFile(file *project.File) {
	fileChan <- file
}

func StartSubmission(sub *project.Submission) {
	subChan <- sub
}

func EndSubmission(sub *project.Submission) {
	subChan <- sub
}

//Serve spawns new processing routines for each new submission received on subChan.
//New files are received on fileChan and then sent to the relevant submission process.
//Incomplete submissions are read from disk and reprocessed using the ProcessStored function.
func Serve() {
	// Start handlers
	go monitor.Listen()
	go func() {
		stored := monitor.GetStored()
		for subId, busy := range stored {
			if busy {
				go ProcessStored(subId)
			}
		}
	}()
	subs := make(map[bson.ObjectId]chan *project.File)
	for {
		select {
		case sub := <-subChan:
			if ch, ok := subs[sub.Id]; ok {
				close(ch)
				delete(subs, sub.Id)
			} else {
				subs[sub.Id] = make(chan *project.File)
				go ProcessSubmission(sub, subs[sub.Id])
			}
		case file := <-fileChan:
			if ch, ok := subs[file.SubId]; ok {
				ch <- file
			} else {
				util.Log(fmt.Errorf("No channel found for submission: %q", file.SubId))
			}
		}
	}
}

//ProcessStored processes an incompletely processed submission.
//It retrieves files in the submission from the db and sends them on fileChan to be processed.
func ProcessStored(subId bson.ObjectId) {
	sub, err := db.GetSubmission(bson.M{project.ID: subId}, nil)
	if err != nil {
		util.Log(err)
		return
	}
	total, err := db.Count(db.FILES, bson.M{project.SUBID: subId})
	if err != nil {
		util.Log(err)
		return
	}
	StartSubmission(sub)
	count := 0
	for count < total {
		matcher := bson.M{project.SUBID: subId, project.INFO + "." + project.NUM: count}
		file, err := db.GetFile(matcher, nil)
		if err != nil {
			util.Log(err)
			return
		}
		AddFile(file)
		count++
	}
	EndSubmission(sub)
}

//ProcessSubmission processes a new submission.
//It listens for incoming files on fileChan and processes them.
func ProcessSubmission(sub *project.Submission, rcvFile chan *project.File) {
	monitor.Busy(sub.Id)
	util.Log("Processing submission", sub)
	dir := filepath.Join(os.TempDir(), sub.Id.Hex())
	//defer os.RemoveAll(dir)
	tests, err := SetupTests(sub.ProjectId, dir)
	if err != nil {
		util.Log(err)
		return
	}
	for {
		file, ok := <-rcvFile
		if !ok {
			break
		}
		err := ProcessFile(file, dir, tests)
		if err != nil {
			util.Log(err)
			return
		}
	}
	util.Log("Processed submission", sub)
	monitor.Done(sub.Id)
}

//ProcessFile processes a file according to its type.
func ProcessFile(f *project.File, dir string, tests []*TestRunner) error {
	util.Log("Processing file", f.Id)
	switch f.Type{
	case project.ARCHIVE:
		err := ProcessArchive(f, dir, tests)
		if err != nil {
			return err
		}
		db.RemoveFileByID(f.Id)
	case project.SRC, project.EXEC:
		err := Evaluate(f, dir, tests, f.Type == project.SRC)
		if err != nil {
			return err
		}
	}
	util.Log("Processed file", f.Id)
	return nil
}

//ProcessArchive extracts files from an archive and processes them.
func ProcessArchive(archive *project.File, dir string, tests []*TestRunner) error {
	files, err := util.UnzipToMap(archive.Data)
	if err != nil {
		return err
	}
	for name, data := range files {
		file, err := project.ParseName(name)
		if err != nil {
			return err
		}
		matcher := bson.M{project.SUBID: archive.SubId, project.NUM: file.Num}
		file, err = db.GetFile(matcher, nil)
		if err != nil {
			file.SubId = archive.SubId
			file.Data = data
			err = db.AddFile(file)
			if err != nil {
				return err
			}
		}
		err = ProcessFile(file, dir, tests)
		if err != nil {
			return err
		}
	}
	return nil
}

//Evaluate evaluates a source or compiled file by attempting to run tests and tools on it.
func Evaluate(f *project.File, dir string, tests []*TestRunner, isSource bool) error {
	target, err := ExtractFile(f, dir)
	if err != nil {
		return err
	}
	compiled, err := Compile(f.Id, target, isSource)
	if err != nil {
		return err
	}
	if !compiled {
		return nil
	}
	f, err = db.GetFile(bson.M{project.ID: f.Id}, nil)
	if err != nil {
		return err
	}
	for _, test := range tests {
		err = test.Execute(f, target.Dir)
		if err != nil {
			return err
		}
	}
	err = RunTools(f, target)
	if err != nil {
		return err
	}
	return nil
}

//ExtractFile saves a file to filesystem.
//It returns file info used by tools & tests.
func ExtractFile(file *project.File, dir string) (*tool.TargetInfo, error) {
	matcher := bson.M{project.ID: file.SubId}
	s, err := db.GetSubmission(matcher, nil)
	if err != nil {
		return nil, err
	}
	matcher = bson.M{project.ID: s.ProjectId}
	p, err := db.GetProject(matcher, nil)
	if err != nil {
		return nil, err
	}
	ti := tool.NewTarget(file.Name, p.Lang, file.Package, dir)
	err = util.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), file.Data)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

//Compile compiles a java source file and saves the results thereof.
//It returns true if compiled successfully.
func Compile(fileId bson.ObjectId, ti *tool.TargetInfo, isSource bool) (bool, error) {
	var res *tool.Result
	var err error
	javac := java.NewJavac(ti.Dir)
	if isSource {
		res, err = javac.Run(fileId, ti)
		if err != nil {
			return false, err
		}
	} else {
		res = tool.NewResult(fileId, javac, []byte(""))
	}
	util.Log("Compile result", res)
	err = AddResult(res)
	if err != nil {
		return false, err
	}
	return true, nil
}

//AddResult adds a tool result to the db.
//It updates the associated file's list of results to point to this new result.
func AddResult(res *tool.Result) error {
	matcher := bson.M{project.ID: res.FileId}
	change := bson.M{db.SET: bson.M{project.RES + "." + res.Name: res.Id}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		return err
	}
	return db.AddResult(res)
}

//RunTools runs all available tools on a file, skipping previously run tools.
func RunTools(f *project.File, ti *tool.TargetInfo) error {
	fb := findbugs.NewFindBugs()
	if _, ok := f.Results[fb.GetName()]; ok {
		return nil
	}
	res, err := fb.Run(f.Id, ti)
	util.Log("Tool result", res)
	if err != nil {
		return err
	}
	return AddResult(res)
}
