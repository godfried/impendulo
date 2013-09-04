package jpf

import (
	//"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)

var listenersFile = "listeners.json"
var searchesFile = "searches.json"

//Listener is a JPF Listener.
type Class struct {
	Name    string
	Package string
}

//Listeners retrieves all JPF Listeners.
func Listeners() ([]*Class, error) {
	return GetClasses("listeners", listenersFile)
}


//Listeners retrieves all JPF Listeners.
func Searches() ([]*Class, error) {
	return GetClasses("searches", searchesFile)
}

func GetClasses(tipe, fname string) (classes []*Class, err error) {
	var data []byte
	path := filepath.Join(util.BaseDir(), fname)
	classes, err = loadClasses(path)
	if err == nil {
		return
	}
	data, err = findClasses(tipe, path)
	if err != nil {
		return
	}
	classes, err = readClasses(data)
	return
}


//FindListeners searches for JPF Listeners in the jpf-core directory tree.
func findClasses(tipe, fname string) (found []byte, err error) {
	target := tool.NewTarget("JPFFinder.java", "java", "finder",
		config.Config(config.JPF_FINDER_DIR))
	cp := filepath.Join(config.Config(config.JPF_HOME), "build", "main") +
		":" + target.Dir + ":" + config.Config(config.GSON_JAR)
	comp := javac.New(cp)
	_, err = comp.Run(bson.NewObjectId(), target)
	if err != nil {
		return
	}
	args := []string{config.Config(config.JAVA), "-cp", cp,
		target.Executable(), tipe, fname}
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(fname)
	if err == nil{
		found = util.ReadBytes(resFile)
	} else if execRes.Err != nil {
		err = execRes.Err
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run finder: %q.",
			string(execRes.StdErr))
	}
	return
}

func readClasses(data []byte) (classes []*Class, err error) {
	err = json.Unmarshal(data, &classes)
	return
}

func loadClasses(fname string) (vals []*Class, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	data := util.ReadBytes(f)
	vals, err = readClasses(data)
	return
}
