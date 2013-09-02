package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"path/filepath"
)

var(
	settings map[string]string
	DEFAULT = filepath.Join(os.Getenv("GOPATH"), "src", 
	"github.com", "godfried", "impendulo", "config.txt")
)

func init(){
	settings = make(map[string]string)
	err := LoadConfigs(DEFAULT)
	if err != nil{
		fmt.Println(err)
	}
}

//LoadConfigs loads configurations from a file.
//Configurations are key-value pairs on different lines.
//Keys are seperated from the value by a '='.
func LoadConfigs(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		vals := strings.Split(scanner.Text(), "=")
		if len(vals) != 2 {
			return fmt.Errorf("Config file not formatted correctly.")
		}
		name := strings.TrimSpace(vals[0])
		value := strings.TrimSpace(vals[1])
		settings[name] = value
	}
	return scanner.Err()
}

//GetConfig attempts to retrieve the named config.
func GetConfig(name string) string {
	ret, ok := settings[name]
	if !ok {
		panic("Config not found: " + name)
	}
	return ret
}

//SetConfig sets the config 'name' to 'value'.
func SetConfig(name, value string) {
	settings[name] = value
}

const (
	JUNIT_EXEC        = "junit_exec"
	LINT4J            = "lint4j"
	FINDBUGS          = "findbugs"
	JUNIT_JAR         = "junit_jar"
	ANT               = "ant"
	ANT_JUNIT         = "ant_junit"
	JAVAC             = "javac"
	JAVA              = "java"
	JPF_JAR           = "jpf_jar"
	RUNJPF_JAR        = "runjpf_jar"
	GSON_JAR          = "gson_jar"
	JPF_RUNNER_DIR        = "jpf_runner_dir"
	TESTING_DIR       = "testing_dir"
	JPF_HOME          = "jpf_home"
	JPF_FINDER_DIR      = "jpf_finder_dir"
	PMD               = "pmd"
	CHECKSTYLE        = "checkstyle"
	CHECKSTYLE_CONFIG = "checkstyle_config"
	DIFF2HTML         = "diff2html"
	DIFF              = "diff"
)
