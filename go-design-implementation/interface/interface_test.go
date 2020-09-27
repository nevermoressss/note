package main

import (
	"fmt"
	"testing"
)

type coder interface {
	code()
	debug()
	p() int
}

type Gopher struct {
	i int
	language string
}

func (p Gopher) code() {
	p.i++
	fmt.Printf("I am coding %s language\n", p.language)
}

func (p *Gopher) debug() {
	fmt.Printf("I am debuging %s language\n", p.language)
}

func (p Gopher) p() int{
	return p.i
}

func TestInterface(t *testing.T) {
	var c coder = &Gopher{1,"Go"}
	c.code()
	c.debug()
	t.Log(c.p())
}
