package main

import (
	"bufio"
	"github.com/ktat/go-pager"
	"os"
)

func main() {
	var p pager.Pager
	p.Init()
	in := make(chan string)
	end := make(chan int)
	go readStdin(in)
	go func(end chan int) {
		if p.PollEvent() == false {
			end <- 1
		}
	}(end)
	go func() {
		for {
			l, ok := <-in
			if ok == false {
				end <- 1
				break
			} else {
				p.AddContent(l + "\n")
				p.Draw()
			}
		}
	}()
	<-end
	p.Close()
	close(in)
	close(end)
}

func readStdin(ch chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var l = scanner.Text()
		ch <- l
	}
}
