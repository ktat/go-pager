# go-pager

pager.

# Usage
```
  import "github.com/ktat/go-pager"
 
	var p pager.Pager
	p.Init()
	p.SetContent(`aaaaaaaaaa`)
	if p.PollEvent() == false {
		p.Close()
	}
```
# License

MIT

# Author

Atushi Kato (ktat)
