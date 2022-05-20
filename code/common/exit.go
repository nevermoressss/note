package common

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/**
 * @author nevermoress
 * @description // 用于优雅关闭项目，AddExitFun添加注册前执行的方法，默认监听 Interrupt，Kill，15 特殊监听请先调用WithListen修改监听方式
 * @date 16:08 2021/2/7
 **/

var e exit

type exit struct {
	once      sync.Once
	eFun      []func()
	listenFun func()
}

// 添加优雅关闭执行的方法
func AddExitFun(f func()) {
	e.addExitFun(f)
}

// 特殊listen
func WithListen(f func()) {
	e.listenFun = f
}

// 执行监听
func (e *exit) listen() {
	if e.listenFun != nil {
		e.listenFun()
		return
	}
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-term
		for _, fun := range e.eFun {
			fun()
		}
		os.Exit(0)
	}()
}

// 添加关闭前需要用到的方法
func (e *exit) addExitFun(f func()) {
	e.once.Do(
		func() {
			e.listen()
		},
	)
	e.eFun = append(e.eFun, f)
}
