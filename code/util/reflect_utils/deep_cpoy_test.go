package reflect_utils

import (
	"fmt"
	json "github.com/json-iterator/go"
	"testing"
)

type TestStructA struct {
	A int
	B int32
	C int64
	D string
	E float32
	F float64
	H *InnerStructA
}

type InnerStructA struct {
	A int
	B int32
	C int64
	D string
	E float32
	F float64
}
type TestStructB struct {
	A int
	B int32
	C int64
	D string
	E float32
	F float64
	H *InnerStructB
}

type InnerStructB struct {
	A int
	B int32
	C int64
	D string
	E float32
	F float64
}

func TestDeepCopy(t *testing.T) {
	t1 := &TestStructA{
		A: 1,
		B: 2,
		C: 3,
		D: "4",
		E: 5,
		F: 6,
		H: &InnerStructA{
			A: 1,
			B: 2,
			C: 3,
			D: "4",
			E: 5,
			F: 6,
		},
	}

	t2 := &TestStructB{}

	DeepCopy(t2, t1)

	msg1, err := json.Marshal(t1)
	fmt.Println(string(msg1), err)

	msg2, err := json.Marshal(t2)
	fmt.Println(string(msg2), err)
}

func TestA1(t *testing.T)  {
	var a []int
	for _,i:=range a{
		println(i)
	}
}
func BenchmarkDeepCopy(b *testing.B)  {
	t1 := &TestStructA{
		A: 1,
		B: 2,
		C: 3,
		D: "4",
		E: 5,
		F: 6,
		H: &InnerStructA{
			A: 1,
			B: 2,
			C: 3,
			D: "4",
			E: 5,
			F: 6,
		},
	}

	t2 := &TestStructB{}
	DeepCopy(t2, t1)
}
