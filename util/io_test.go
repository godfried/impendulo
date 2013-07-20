package util

import (
	"bytes"
	"errors"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestMapStorage(t *testing.T) {
	id1 := bson.NewObjectId()
	id2 := bson.NewObjectId()
	id3 := bson.NewObjectId()
	id4 := bson.NewObjectId()
	m1 := map[bson.ObjectId]bool{id1: true, id2: false, id3: false, id4: true}
	err := SaveMap(m1, "test.gob")
	if err != nil {
		t.Error(err, "Error saving map")
	}
	m2, err := LoadMap("test.gob")
	if err != nil {
		t.Error(err, "Error loading map")
	}
	if len(m1) != len(m2) {
		t.Error(errors.New("Error loading map; invalid size"))
	}
	for k, v := range m1 {
		if v != m2[k] {
			t.Error(errors.New("Error loading map, values not equal."))
		}
	}
}

func TestReadBytes(t *testing.T) {
	orig := []byte("bytes")
	buff := bytes.NewBuffer(orig)
	ret := ReadBytes(buff)
	if !bytes.Equal(orig, ret) {
		t.Error(errors.New("Bytes not equal"))
	}
}

func TestZip(t *testing.T) {
	files := map[string][]byte{"readme.txt": []byte("This archive contains some text files."), "gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"), "todo.txt": []byte("Get animal handling licence.\nWrite more examples.")}
	zipped, err := Zip(files)
	if err != nil {
		t.Error(err)
	}
	unzipped, err := UnzipToMap(zipped)
	if err != nil {
		t.Error(err)
	}
	if len(files) != len(unzipped) {
		t.Error(errors.New("Zip error; invalid size"))
	}
	for k, v := range files {
		if !bytes.Equal(v, unzipped[k]) {
			t.Error(errors.New("Zip error, values not equal."))
		}
	}

}
