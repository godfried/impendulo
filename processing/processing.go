package processing

import (
	"container/list"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"fmt"
)

var subChan chan bson.ObjectId
var fileChan chan *ids

func init() {
	subChan = make(chan bson.ObjectId)
	fileChan = make(chan *ids)
}


type ids struct {
	fileId bson.ObjectId
	subId bson.ObjectId
}

//AddFile sends a file id to be processed.
func AddFile(file *project.File) {
	fileChan <- &ids{file.Id, file.SubId}
}

//EndSubmission stops this submission's goroutine to.
func StartSubmission(subId bson.ObjectId) {
	subChan <- subId
}

func EndSubmission(subId bson.ObjectId) {
	subChan <- subId
}

func None() interface{}{
	type e struct{}
	return e{}
}	


const MAX_PROCS = 20

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
//Incomplete submissions are read from disk and reprocessed via ProcessStored.
func Serve() {
	helpers := make(map[bson.ObjectId]*ProcHelper)
	subs := list.New()
	busy := 0
	for {
		if busy <  MAX_PROCS && subs.Len() > 0{
			subId := subs.Remove(subs.Front()).(bson.ObjectId)
			helpers[subId] = NewProcHelper(subId)
			go helpers[subId].Handle()
			busy ++
		} else if busy < 0{
			break
		}
		select{
		case subId := <-subChan:
			if helper, ok := helpers[subId]; ok{
				helper.doneChan <- None()
			} else{
				subs.PushBack(subId)
			}
		case ids := <- fileChan:
			if helper, ok := helpers[ids.subId]; ok{
				helper.serveChan <- ids.fileId
			} else{
				util.Log(fmt.Errorf("No submission %q found for file %q", ids.subId, ids.fileId))
			}
		}
	}
}

func NewProcHelper(subId bson.ObjectId) *ProcHelper{
	return &ProcHelper{subId, make(chan bson.ObjectId), make(chan interface{})}
}

type ProcHelper struct{
	subId bson.ObjectId
	serveChan chan bson.ObjectId
	doneChan chan interface{}
}

func (this *ProcHelper) Handle(){
	procChan := make(chan bson.ObjectId)
	stopChan := make(chan interface{})
	proc, err := NewProcessor(this.subId)
	if err != nil{
		util.Log(err)
	}
	go proc.Process(procChan, stopChan)
	files := list.New()
	busy, done := false, false
	for{
		if !busy{
			if files.Len() > 0{
				fId := files.Remove(files.Front()).(bson.ObjectId)
				procChan <- fId
				busy = true
			} else if done{
				stopChan <- None()
				return
			}
		}
		select{
		case fId := <- this.serveChan:
			files.PushBack(fId)
		case <- procChan:
			busy = false
		case <- this.doneChan:
			done = true
		}
	}
}

//Processor is used to process individual submissions.
type Processor struct {
	sub     *project.Submission
	tests   []*TestRunner
	rootDir string
	srcDir  string
	toolDir string
	jpfPath string
}

func NewProcessor(subId bson.ObjectId)(proc *Processor, err error) {
	sub, err := db.GetSubmission(bson.M{project.ID:subId}, nil)
	if err != nil{
		return
	}
	dir := filepath.Join(os.TempDir(), sub.Id.Hex())
	proc = &Processor{sub: sub, rootDir: dir, srcDir: filepath.Join(dir, "src"), toolDir: filepath.Join(dir, "tools")}
	return
}

//Process processes a new submission.
//It listens for incoming files and creates new goroutines to processes them.
func (this *Processor) Process(fileChan chan bson.ObjectId, doneChan chan interface{}) {
	util.Log("Processing submission", this.sub)
	defer os.RemoveAll(this.rootDir)
	err := this.SetupJPF()
	if err != nil {
		util.Log(err)
	}
	this.tests, err = SetupTests(this.sub.ProjectId, this.toolDir)
	if err != nil {
		util.Log(err)
	}
outer:
	for {
		select{
		case fId := <- fileChan:
			file, err := db.GetFile(bson.M{project.ID: fId}, nil)
			if err != nil {
				util.Log(err)
			}
			err = this.ProcessFile(file)
			if err != nil {
				util.Log(err)
			}
			fileChan <- fId
		case <- doneChan:
			break outer
		}
	}
	util.Log("Processed submission", this.sub)
}

//Setup sets up the environment needed for this Processor to function correctly.
func (this *Processor) SetupJPF() error {
	err := util.Copy(this.toolDir, config.GetConfig(config.RUNNER_DIR))
	if err != nil {
		return err
	}
	jpfFile, err := db.GetJPF(bson.M{project.PROJECT_ID: this.sub.ProjectId}, nil)
	if err != nil {
		return err
	}
	this.jpfPath = filepath.Join(this.toolDir, jpfFile.Name)
	return util.SaveFile(this.jpfPath, jpfFile.Data)
}

//ProcessFile processes a file according to its type.
func (this *Processor) ProcessFile(file *project.File) error {
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

//Analyser is used to run tools on a file.
type Analyser struct {
	proc   *Processor
	file   *project.File
	target *tool.TargetInfo
}

//Eval evaluates a source or compiled file by attempting to run tests and tools on it.
func (this *Analyser) Eval() error {
	err := this.buildTarget()
	if err != nil {
		return err
	}
	var compileErr bool
	compileErr, err = this.compile()
	if err != nil {
		return err
	} else if compileErr {
		return nil
	}
	this.file, err = db.GetFile(bson.M{project.ID: this.file.Id}, nil)
	if err != nil {
		return err
	}
	for _, test := range this.proc.tests {
		err = test.Run(this.file, this.proc.srcDir)
		if err != nil {
			util.Log(err)
		}
	}
	this.RunTools()
	return nil
}

//buildTarget saves a file to filesystem.
//It returns file info used by tools & tests.
func (this *Analyser) buildTarget() error {
	matcher := bson.M{project.ID: this.proc.sub.ProjectId}
	p, err := db.GetProject(matcher, nil)
	if err != nil {
		return err
	}
	this.target = tool.NewTarget(this.file.Name, p.Lang, this.file.Package, this.proc.srcDir)
	return util.SaveFile(this.target.FilePath(), this.file.Data)
}

//compile compiles a java source file and saves the results thereof.
//It returns true if compiled successfully.
func (this *Analyser) compile() (bool, error) {
	comp := javac.NewJavac(this.target.Dir)
	res, err := comp.Run(this.file.Id, this.target)
	compileErr := javac.IsCompileError(err)
	if err != nil && !compileErr {
		return false, err
	}
	return compileErr, AddResult(res)
}

//RunTools runs all available tools on a file, skipping previously run tools.
func (this *Analyser) RunTools() {
	tools := []tool.Tool{findbugs.NewFindBugs(), pmd.NewPMD(),
		jpf.NewJPF(this.proc.toolDir, this.proc.jpfPath),
		checkstyle.NewCheckstyle()}
	for _, tool := range tools {
		if _, ok := this.file.Results[tool.GetName()]; ok {
			continue
		}
		res, err := tool.Run(this.file.Id, this.target)
		if err != nil {
			util.Log(err)
			continue
		}
		err = AddResult(res)
		if err != nil {
			util.Log(err)
		}
	}
}

//AddResult adds a tool result to the db.
//It updates the associated file's list of results to point to this new result.
func AddResult(res tool.Result) error {
	matcher := bson.M{project.ID: res.GetFileId()}
	change := bson.M{db.SET: bson.M{project.RESULTS + "." + res.GetName(): res.GetId()}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		return err
	}
	return db.AddResult(res)
}
