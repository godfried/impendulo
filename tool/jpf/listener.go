package jpf

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/util"
	"os"
	"path/filepath"
	"labix.org/v2/mgo/bson"
)

var listenersFile = "listeners.gob"

func Listeners() (listeners []*Listener, err error) {
	var data []byte
	path := filepath.Join(util.BaseDir(), listenersFile)
	listeners, err = loadListeners(path)
	if err == nil {
		return
	}
	data, err = FindListeners()
	if err != nil {
		return
	}
	listeners, err = readListeners(data)
	if err != nil {
		return
	}
	err = saveListeners(listeners, path)
	return
}

func FindListeners() ([]byte, error) {
	target := tool.NewTarget("ListenerFinder.java", "java", "listener", config.GetConfig(config.LISTENER_DIR))
	cp := filepath.Join(config.GetConfig(config.JPF_HOME), "build", "main") + ":" + target.Dir + ":" + config.GetConfig(config.GSON_JAR)
	comp := javac.NewJavac(cp)
	_, err := comp.Run(bson.NewObjectId(), target) 
	if err != nil {
		return nil, err
	}
	args := []string{config.GetConfig(config.JAVA), "-cp", cp, target.Executable()}
	stdout, stderr, err := tool.RunCommand(args)
	if err != nil {
		return nil, err
	} else if stderr != nil && len(stderr) > 0 {
		return nil, fmt.Errorf("Could not run listener finder: %q.", string(stderr))
	}
	return stdout, err
}

type Listener struct {
	Name    string
	Package string
}

func readListeners(data []byte) (listeners []*Listener, err error) {
	err = json.Unmarshal(data, &listeners)
	return
}

func saveListeners(vals []*Listener, fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating file %q", err, fname)
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(&vals)
	if err != nil {
		return fmt.Errorf("Encountered error %q while encoding map %q to file %q", err, vals, fname)
	}
	return nil
}

func loadListeners(fname string) (vals []*Listener, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(f)
	err = dec.Decode(&vals)
	return
}
