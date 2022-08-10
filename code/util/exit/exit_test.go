package exit

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestDefaultExit(t *testing.T) {
	AddExitFun(func() {
		t.Log(123)
		t.Log(456)
	})
	t.Log("awaiting signal")
	select {}
	t.Log("走不到这里的")
}

func TestWithListen(t *testing.T) {
	WithListen(
		func() {
			term := make(chan os.Signal, 1)
			signal.Notify(term, os.Interrupt, os.Kill, syscall.SIGTERM)
			go func() {
				for {
					<-term
					for _, fun := range e.eFun {
						fun()
					}
					t.Log("哈哈哈就是不退出")
				}
			}()
		},
	)
	AddExitFun(func() {
		t.Log(123)
		t.Log(456)
	})
	t.Log("awaiting signal")
	select {}
	t.Log("走不到这里的")
}