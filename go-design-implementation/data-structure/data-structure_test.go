package data_structure

import (
	"testing"
)

func TestDataStructure(t *testing.T) {
	//append()
	bys := []byte{65}
	s := string(bys)
	t.Log(s)
	bys[0] = 66
	t.Log(s)
}

func TestBool(t *testing.T) {
	var a bool
	!a=returnBoool()
	t.Log(a)
}

func returnBoool() bool {
	return true
}
