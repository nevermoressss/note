package test

import (
	"strings"
	"testing"
)

func TestSpace(t *testing.T) {
	t1:="         1             a               b                        "
	t.Log(strings.TrimSpace(t1))
}