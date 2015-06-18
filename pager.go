package pager

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"regexp"
)

type Pager struct {
	Str          string   // contents to display
	Files        []string // files
	ignore       int      // ignore lines
	Index        int      // file index
	File         string   // current file
	isSlashOn    bool
	isSearchMode bool
	searchIndex  int
	searchString string
}

func (p *Pager) drawLine(x, y int, str string, canSkip bool) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault
	runes := []rune(str)

	minusX := 0
	ignoreLine := p.ignore

	colorMap := map[string]termbox.Attribute{
		"30m": termbox.ColorBlack,
		"31m": termbox.ColorRed,
		"32m": termbox.ColorGreen,
		"33m": termbox.ColorYellow,
		"34m": termbox.ColorBlue,
		"35m": termbox.ColorMagenta,
		"36m": termbox.ColorCyan,
		"37m": termbox.ColorWhite,
	}

	for i := 0; i < len(runes); i++ {
		if canSkip && ignoreLine > 0 {
			ignoreLine--
		} else {
			if runes[i] == '\n' {
				y++
				minusX = i
			}

			if i+2 < len(runes) {
				colorLiteral := string(runes[i : i+2])
				if colorLiteral == "\033[" {
					if runes[i+2] == '3' && runes[i+4] == 'm' {
						// color
						c := string(runes[i+2 : i+5])
						color = colorMap[c]

						i += 4
						minusX += 5
						continue
					} else if i+4 <= len(runes) && string(runes[i+2:i+4]) == "0m" {
						// reset
						color = termbox.ColorDefault

						i += 3
						minusX += 4
						continue
					}
					continue
				}
			}
			if canSkip {
				termbox.SetCell(x+i-minusX, y-(p.ignore)+1, runes[i], color, backgroundColor)
			} else {
				termbox.SetCell(x+i, y, runes[i], color, termbox.ColorWhite)
			}
		}
	}
}

func (p *Pager) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	p.drawLine(2, 1, p.Str, true)
	p.drawLine(0, 0, "Press ESC to exit. :"+p.File+":"+p.searchString+":"+fmt.Sprintf("%V", p.isSlashOn)+":"+fmt.Sprintf("%V", p.isSearchMode), false)
	termbox.Flush()
}

//var K_J = []key{{25, 8, 'J'}}
//var K_k = []key{{28, 8, 'k'}}

func (p *Pager) PollEvent() bool {
	p.Draw()
	for {
		if p.isSlashOn == false {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEsc:
					termbox.Flush()
					return false
				case termbox.KeyArrowRight, termbox.KeyCtrlL:
					if p.isMaxIndex() {
						p.Index--
					} else {
						termbox.Sync()
					}
					termbox.Flush()
					return true
				case termbox.KeyArrowLeft, termbox.KeyCtrlH:
					if p.Index >= 1 {
						p.Index -= 2
						termbox.Sync()
						return true
					}
					p.Draw()
				case termbox.KeyCtrlN, termbox.KeyArrowDown:
					p.scrollDown()
				case termbox.KeyCtrlP, termbox.KeyArrowUp:
					p.scrollUp()
				case termbox.KeyCtrlD, termbox.KeySpace:
					matched := regexp.MustCompile("(?s)\\n").FindAllString(p.Str, -1)
					_, y := termbox.Size()
					if p.ignore+29 < (len(matched) - y) {
						p.ignore += 29
					} else {
						p.ignore = len(matched) - y
					}
					p.Draw()
				case termbox.KeyCtrlU:
					p.ignore -= 29
					if p.ignore < 0 {
						p.ignore = 0
					}
					p.Draw()
				default:
					if ev.Ch == 106 { // j
						p.scrollDown()
					} else if ev.Ch == 107 { // k
						p.scrollUp()
					} else if ev.Ch == 113 { // q
						termbox.Sync()
						return false
					} else if ev.Ch == 47 { // /
						p.isSlashOn = true
						p.Draw()
					} else {
						p.Draw()
					}
				}
			default:
				p.Draw()
			}
		} else if p.isSearchMode {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEnter:
					// nothing to do
				case termbox.KeyEsc:
					p.isSearchMode = false
					p.isSlashOn = false
					p.searchString = ""
				default:
					if ev.Ch == 110 { // n
						p.searchForward()
					} else if ev.Ch == 78 { // N
						p.searchBackward()
					} else {
						p.searchString += string(ev.Ch)
						p.isSearchMode = false
					}
				}
			}
			p.Draw()
		} else {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEnter:
					p.isSearchMode = true
					p.searchForward()
				case termbox.KeyEsc:
					p.isSlashOn = false
					p.searchString = ""
				default:
					p.searchString += string(ev.Ch)
				}
			}
			p.Draw()
		}
	}
	return false
}

func (p *Pager) searchForward() {
	var lines []string
	matched := regexp.MustCompile(p.searchString).FindAllStringIndex(p.Str, p.searchIndex)
	if len(matched) > 0 {
		if len(matched) >= p.searchIndex {
			lines = regexp.MustCompile("(?m)\\n").FindAllString(p.Str[0:matched[p.searchIndex-1][1]], -1)
			p.ignore = len(lines)
			p.searchIndex++
		}
	}

}

func (p *Pager) searchBackward() {
	var lines []string
	matched := regexp.MustCompile(p.searchString).FindAllStringIndex(p.Str, p.searchIndex)
	if len(matched) > 0 {
		if p.searchIndex > 2 {
			p.searchIndex--
			lines = regexp.MustCompile("(?m)\\n").FindAllString(p.Str[0:matched[p.searchIndex-1][1]], -1)
			p.ignore = len(lines)
		}
	}

}

func (p *Pager) scrollDown() {
	matched := regexp.MustCompile("(?s)\\n").FindAllString(p.Str, -1)
	_, y := termbox.Size()
	if p.ignore < (len(matched) - y) {
		p.ignore++
	}
	p.Draw()
}

func (p *Pager) scrollUp() {
	if p.ignore > 0 {
		p.ignore--
	}
	p.Draw()
}

func (p *Pager) Init() {
	err := termbox.Init()
	p.isSlashOn = false
	p.isSearchMode = false
	p.searchIndex = 1
	if err != nil {
		panic(err)
	}
}

func (p *Pager) Close() {
	termbox.Flush()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Sync()
	termbox.Close()
}

func (p *Pager) isMaxIndex() bool {
	return len(p.Files) == (p.Index + 1)
}
