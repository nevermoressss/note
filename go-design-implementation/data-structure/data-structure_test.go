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
