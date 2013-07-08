package processing

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing/monitor"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"container/list"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

var fileChan chan *FileId
var subChan chan *project.Submission

func init() {
	fileChan = make(chan *FileId)
	subChan = make(chan *project.Submission)
}

type FileId struct {
	id    bson.ObjectId
	subid bson.ObjectId
}

func AddFile(file *project.File) {
	fileChan <- &FileId{file.Id, file.SubId}
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
	subs := make(map[bson.ObjectId]chan bson.ObjectId)
	for {
		select {
		case sub := <-subChan:
			if ch, ok := subs[sub.Id]; ok {
				close(ch)
				delete(subs, sub.Id)
			} else {
				subs[sub.Id] = make(chan bson.ObjectId)
				proc := NewProcessor(sub, subs[sub.Id])
				go proc.Process()
			}
		case fileIds := <-fileChan:
			if ch, ok := subs[fileIds.subid]; ok {
				ch <- fileIds.id
			} else {
				util.Log(fmt.Errorf("No channel found for submission: %q", fileIds.subid))
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
		matcher := bson.M{project.SUBID: subId, project.NUM: count}
		file, err := db.GetFile(matcher, bson.M{project.ID: 1})
		if err != nil {
			util.Log(err)
			return
		}
		file.SubId = sub.Id
		AddFile(file)
		count++
	}
	EndSubmission(sub)
}

type Processor struct {
	sub     *project.Submission
	recv    chan bson.ObjectId
	tests   []*TestRunner
	dir     string
	jpfPath string
}

func NewProcessor(sub *project.Submission, recv chan bson.ObjectId) *Processor {
	return &Processor{sub: sub, recv: recv, dir: filepath.Join(os.TempDir(), sub.Id.Hex())}
}

//ProcessSubmission processes a new submission.
//It listens for incoming files on fileChan and processes them.
func (this *Processor) Process() {
	monitor.Busy(this.sub.Id)
	util.Log("Processing submission", this.sub)
	defer os.RemoveAll(this.dir)
	err := this.Setup()
	if err != nil {
		util.Log(err)
	}
	var fId bson.ObjectId
	files := list.New()
	receiving, busy := true, false
	errChan := make(chan error)
	for receiving || busy {
		select {
		case fId, receiving = <-this.recv:
			if receiving {
				files.PushBack(fId)
				if !busy {
					go func() { errChan <- nil }()
				}
			}
		case err := <-errChan:
			busy = false
			if err != nil {
				util.Log(err)
			}
			if e := files.Front(); e != nil {
				busy = true
				fId := files.Remove(e).(bson.ObjectId)
				go this.goFile(fId, errChan)
			}
		}
	}
	util.Log("Processed submission", this.sub)
	monitor.Done(this.sub.Id)
}

func (this *Processor) Setup() error {
	var err error
	this.tests, err = SetupTests(this.sub.ProjectId, this.dir)
	if err != nil {
		return err
	}
	jpfFile, err := db.GetJPF(bson.M{project.PROJECT_ID: this.sub.ProjectId}, nil)
	if err != nil {
		return err
	}
	this.jpfPath = filepath.Join(this.dir, jpfFile.Name)
	return util.SaveFile(this.jpfPath, jpfFile.Data)
}

func (this *Processor) goFile(fId bson.ObjectId, errChan chan error) {
	file, err := db.GetFile(bson.M{project.ID: fId}, nil)
	if err == nil {
		err = this.ProcessFile(file)
	}
	errChan <- err
}

//ProcessFile processes a file according to its type.
func (this *Processor) ProcessFile(file *project.File) error {
	util.Log("Processing file", file.Id)
	switch file.Type {
	case project.ARCHIVE:
		err := this.extract(file)
		if err != nil {
			return err
		}
		db.RemoveFileByID(file.Id)
	case project.SRC:
		analyser := &Analyser{proc: this, file: file}
		err := analyser.Eval()
		if err != nil {
			return err
		}
	}
	util.Log("Processed file", file.Id)
	return nil
}

//ProcessArchive extracts files from an archive and processes them.
func (this *Processor) extract(archive *project.File) error {
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
		foundFile, err := db.GetFile(matcher, nil)
		if err != nil {
			file.SubId = archive.SubId
			file.Data = data
			err = db.AddFile(file)
			if err != nil {
				return err
			}
		} else {
			file = foundFile
		}
		err = this.ProcessFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}

type Analyser struct {
	proc   *Processor
	file   *project.File
	target *tool.TargetInfo
}

//Evaluate evaluates a source or compiled file by attempting to run tests and tools on it.
func (this *Analyser) Eval() error {
	err := this.buildTarget()
	if err != nil {
		return err
	}
	err = this.compile()
	if err != nil {
		return err
	}
	this.file, err = db.GetFile(bson.M{project.ID: this.file.Id}, nil)
	if err != nil {
		return err
	}
	for _, test := range this.proc.tests {
		err = test.Run(this.file, this.proc.dir)
		if err != nil {
			return err
		}
	}
	return this.RunTools()
}

//ExtractFile saves a file to filesystem.
//It returns file info used by tools & tests.
func (this *Analyser) buildTarget() error {
	matcher := bson.M{project.ID: this.proc.sub.ProjectId}
	p, err := db.GetProject(matcher, nil)
	if err != nil {
		return err
	}
	this.target = tool.NewTarget(this.file.Name, p.Lang, this.file.Package, this.proc.dir)
	return util.SaveFile(this.target.FilePath(), this.file.Data)
}

//Compile compiles a java source file and saves the results thereof.
//It returns true if compiled successfully.
func (this *Analyser) compile() error {
	comp := javac.NewJavac(this.target.Dir)
	res, err := comp.Run(this.file.Id, this.target)
	if err != nil {
		return err
	}
	util.Log("Compile result", res)
	return AddResult(res)
}

//RunTools runs all available tools on a file, skipping previously run tools.
func (this *Analyser) RunTools() error {
	/*fb := findbugs.NewFindBugs()
	if _, ok := this.file.Results[fb.GetName()]; ok {
		return nil
	}
	res, _ := fb.Run(this.file.Id, this.target)
	util.Log("Findbugs result", res)
	if err != nil {
		return err
	}
	err := AddResult(res)
	if err != nil {
		return err
	}*/
	j := jpf.NewJPF(this.proc.jpfPath)
	if _, ok := this.file.Results[j.GetName()]; ok {
		return nil
	}
	res, err := j.Run(this.file.Id, this.target)
	util.Log("JPF result", res)
	if err != nil {
		return err
	}
	return AddResult(res)
}

//AddResult adds a tool result to the db.
//It updates the associated file's list of results to point to this new result.
func AddResult(res tool.Result) error {
	matcher := bson.M{project.ID: res.GetFileId()}
	change := bson.M{db.SET: bson.M{project.RES + "." + res.Name(): res.GetId()}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		return err
	}
	return db.AddResult(res)
}
