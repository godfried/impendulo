package db

import (
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"strings"
	"fmt"
)

func GetCheckstyleResult(matcher, selector interface{}) (ret *checkstyle.CheckstyleResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetPMDResult(matcher, selector interface{}) (ret *pmd.PMDResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetFindbugsResult(matcher, selector interface{}) (ret *findbugs.FindbugsResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetJPFResult(matcher, selector interface{}) (ret *jpf.JPFResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetJUnitResult(matcher, selector interface{}) (ret *junit.JUnitResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetJavacResult(matcher, selector interface{}) (ret *javac.JavacResult, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

func GetResult(name string, matcher, selector interface{}) (ret tool.Result, err error) {
	if strings.HasPrefix(name, javac.NAME) {
		ret, err = GetJavacResult(matcher, selector)
	} else if strings.HasPrefix(name, junit.NAME) {
		ret, err = GetJUnitResult(matcher, selector)
	} else if strings.HasPrefix(name, jpf.NAME) {
		ret, err = GetJPFResult(matcher, selector)
	} else if strings.HasPrefix(name, findbugs.NAME) {
		ret, err = GetFindbugsResult(matcher, selector)
	} else if strings.HasPrefix(name, pmd.NAME) {
		ret, err = GetPMDResult(matcher, selector)
	} else if strings.HasPrefix(name, checkstyle.NAME) {
		ret, err = GetCheckstyleResult(matcher, selector)
	} else {
		err = fmt.Errorf("Unknown result %q.", name)
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(r tool.Result) (err error) {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(RESULTS)
	err = col.Insert(r)
	if err != nil {
		err = &DBAddError{r.String(), err}
	}
	return
}
