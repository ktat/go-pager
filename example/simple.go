package main

import (
	"github.com/ktat/go-pager"
	"strconv"
)

func main() {
	var p pager.Pager
	p.Init()
	s := ""
	for i := 0; i < 1000; i++ {
		s += strconv.Itoa(i) + "\n"
	}
	p.SetContent(s)
	if p.PollEvent() == false {
		p.Close()
	}
}
