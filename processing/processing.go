package processing

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing/monitor"
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
)

var subChan chan *project.Submission

func init() {
	subChan = make(chan *project.Submission)
}


//EndSubmission stops this submission's goroutine to.
func DoSubmission(sub *project.Submission) {
	subChan <- sub
}

//Serve spawns new processing routines for each submission started.
//Added files are received here and then sent to the relevant submission goroutine.
//Incomplete submissions are read from disk and reprocessed via ProcessStored.
func Serve() {
	//go monitor.Listen()
	go func() {
		stored := monitor.GetStored()
		for subId, busy := range stored {
			if busy {
				go ProcessStored(subId)
			}
		}
	}()
	subs := make(map[bson.ObjectId]bool)
	for {
		select{
		case sub := <-subChan:
			if subs[sub.Id]{
				delete(subs, sub.Id)
				//monitor.Done(sub.Id)
			} else {
				subs[sub.Id] = true 
				proc := NewProcessor(sub)
				go proc.Process()
				//monitor.Busy(sub.Id)
			} 
		}
	}
}

//ProcessStored processes incompletely processed submissions.
//It retrieves files in the submission and sends to be processed.
func ProcessStored(subId bson.ObjectId) {
	sub, err := db.GetSubmission(bson.M{project.ID: subId}, nil)
	if err != nil {
		util.Log(err)
		return
	}
	DoSubmission(sub)
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

func NewProcessor(sub *project.Submission) *Processor {
	dir := filepath.Join(os.TempDir(), sub.Id.Hex())
	return &Processor{sub: sub, rootDir: dir, srcDir: filepath.Join(dir, "src"), toolDir: filepath.Join(dir, "tools")}
}

//Process processes a new submission.
//It listens for incoming files and creates new goroutines to processes them.
func (this *Processor) Process() {
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
	files, err := db.GetFiles(bson.M{project.SUBID:this.sub.Id}, bson.M{project.ID:1, project.NUM:1},project.NUM) 
	if err != nil {
		util.Log(err)
	}
	for _, file := range files{
		file, err := db.GetFile(bson.M{project.ID: file.Id}, nil)
		if err != nil {
			util.Log(err)
		}
		err = this.ProcessFile(file)
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
	util.Log("built ", this.file.Num)
	var compileErr bool
	compileErr, err = this.compile()
	util.Log("compiled ", this.file.Num, err, compileErr)
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
		util.Log("tested ", this.file.Num, err)
		if err != nil {
			util.Log(err)
		}
	}
	this.RunTools()
	util.Log("ran tools ", this.file.Num)
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
		util.Log("ran ", tool.GetName(), this.file.Num)
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
