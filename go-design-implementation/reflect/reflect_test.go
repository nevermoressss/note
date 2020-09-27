package reflect

import (
	"reflect"
	"testing"
)

func TestInterface(t *testing.T) {
	tp := reflect.TypeOf(&Student{})
	t.Log(tp.Elem())
	t.Log(tp)
	v := reflect.ValueOf(Student{})
	t.Log(v)
}

type Student struct {
	Name  string
	Age   int
	Class string
	Score int
}

func BenchmarkReflect_New(b *testing.B) {
	var s *Student
	sv := reflect.TypeOf(Student{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sn := reflect.New(sv)
		s, _ = sn.Interface().(*Student)
	}
	_ = s
}
func BenchmarkDirect_New(b *testing.B) {
	var s *Student
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = new(Student)
	}
	_ = s
}
func BenchmarkReflect_Set(b *testing.B) {
	var s *Student
	sv := reflect.TypeOf(Student{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sn := reflect.New(sv)
		s = sn.Interface().(*Student)
		s.Name = "Jerry"
		s.Age = 18
		s.Class = "20005"
		s.Score = 100
	}
}
func BenchmarkReflect_SetFieldByName(b *testing.B) {
	sv := reflect.TypeOf(Student{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sn := reflect.New(sv).Elem()
		sn.FieldByName("Name").SetString("Jerry")
		sn.FieldByName("Age").SetInt(18)
		sn.FieldByName("Class").SetString("20005")
		sn.FieldByName("Score").SetInt(100)
	}
}
func BenchmarkReflect_SetFieldByIndex(b *testing.B) {
	sv := reflect.TypeOf(Student{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sn := reflect.New(sv).Elem()
		sn.Field(0).SetString("Jerry")
		sn.Field(1).SetInt(18)
		sn.Field(2).SetString("20005")
		sn.Field(3).SetInt(100)
	}
}
func BenchmarkDirect_Set(b *testing.B) {
	var s *Student
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = new(Student)
		s.Name = "Jerry"
		s.Age = 18
		s.Class = "20005"
		s.Score = 100
	}
}
