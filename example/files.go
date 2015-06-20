package main

import (
	"github.com/ktat/go-pager"
	"io/ioutil"
	"log"
)

func main() {
	var p pager.Pager
	p.Init()
	var files = make([]string, 0)
	files = append(files, "./files.go")
	files = append(files, "./simple.go")
	p.Files = files
	for i := 0; i < len(files); i++ {
		c, ioerr := ioutil.ReadFile(files[i])
		if ioerr != nil {
			log.Fatal(ioerr)
		}
		p.Index = i
		p.File = files[i]
		p.SetContent(string(c))
		if p.PollEvent() {
			i = p.Index
		}
	}
	p.Close()
}
