package pager

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"regexp"
)

type Pager struct {
	str          string   // contents to display
	lines        int      // num of lines in str
	Files        []string // files
	ignoreY      int      // ignore lines
	Index        int      // file index
	File         string   // current file
	isSlashOn    bool     // input search string mode
	isSearchMode bool     // search mode
	searchIndex  int      // current search index
	searchStr    string   // string to search
}

func (p *Pager) SetContent(s string) {
	p.str = s
	lines := regexp.MustCompile("(?m)$").FindAllString(p.str, -1)
	p.lines = len(lines)
}

func (p *Pager) drawLine(x, y int, str string, canSkip bool) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault
	foundIndex := 0
	runes := []rune(str)

	minusX := 0

	colorMap := map[string]termbox.Attribute{
		"0m": termbox.ColorBlack,
		"1m": termbox.ColorRed,
		"2m": termbox.ColorGreen,
		"3m": termbox.ColorYellow,
		"4m": termbox.ColorBlue,
		"5m": termbox.ColorMagenta,
		"6m": termbox.ColorCyan,
		"7m": termbox.ColorWhite,
	}
	attrMap := map[string]termbox.Attribute{
		"1m": termbox.AttrBold,
		"3m": termbox.AttrReverse,
		"4m": termbox.AttrUnderline,
	}

	if false {
		// for debug
		for i := 0; i < len(runes); i++ {
			if runes[i] == '\033' {
				print("ESC")
			} else {
				print(string(runes[i]))
			}
		}
		panic(1)
	}

	searchStringLen := len(p.searchStr)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\n' {
			y++
			minusX = i + 1
		}
		if searchStringLen > 0 && i+searchStringLen < len(runes) { // highlight search string
			if string(runes[i:i+searchStringLen]) == p.searchStr[foundIndex:searchStringLen] {
				backgroundColor = termbox.ColorCyan
				foundIndex = searchStringLen - 1
			} else if foundIndex == 0 {
				backgroundColor = termbox.ColorDefault
			} else {
				foundIndex--
			}
		}
		if runes[i] == '\033' { // not good
			minusX++
			if i+2 < len(runes) && string(runes[i:i+3]) == "\033[m" { // reset?
				color = termbox.ColorDefault
				backgroundColor = termbox.ColorDefault
				i += 2
				minusX += 2
			} else if i+3 < len(runes) && runes[i+1] == '[' && runes[i+3] == 'm' {
				t := string(runes[i+2 : i+4])
				if t == "0m" { // reset
					color = termbox.ColorDefault
					backgroundColor = termbox.ColorDefault
				} else { // attribute
					if a, ok := attrMap[t]; ok {
						color |= a
					} else {
						// not supported attribute
					}
				}
				i += 3
				minusX += 3
			} else if i+4 < len(runes) && string(runes[i:i+2]) == "\033[" && (runes[i+2] == '0' || runes[i+2] == '3' || runes[i+2] == '4') && runes[i+4] == 'm' { // \033[30m or  \033[40m
				// color
				c := string(runes[i+3 : i+5])
				if runes[i+2] == '3' {
					color = colorMap[c]
				} else if runes[i+2] == '4' {
					backgroundColor = colorMap[c]
				} else {
					panic(1)
					color |= termbox.AttrBold
				}
				i += 4
				minusX += 4
			} else if i+7 < len(runes) && string(runes[i:i+2]) == "\033[" && runes[i+4] == ';' && (runes[i+5] == '3' || runes[i+2] == '4') && runes[i+7] == 'm' { // \033[01;30m or \033[01;40m
				// attr + color
				a := string(runes[i+2 : i+4])
				c := string(runes[i+6 : i+8])
				if runes[i+2] == '3' {
					color = colorMap[c]
				} else {
					backgroundColor = colorMap[c]
				}
				if a == "01" {
					color |= termbox.AttrBold
				}
				i += 7
				minusX += 7
			} else if i+1 < len(runes) && string(runes[i:i+2]) == "\033[" { // \033[K
				i += 2
				minusX += 2
			}
			continue
		}
		if canSkip {
			termbox.SetCell(x+i-minusX, y-(p.ignoreY)+1, runes[i], color, backgroundColor)
		} else {
			termbox.SetCell(x+i, y, runes[i], termbox.ColorBlue, termbox.ColorWhite)
		}

	}
}

func (p *Pager) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	p.drawLine(0, 0, p.str, true)
	maxX, _ := termbox.Size()
	empty := make([]byte, maxX)
	mode := ""
	file := ""
	nextFileUsage := ""
	if p.File != "" {
		file = " :: [file: " + p.File + " ]"
	}
	if p.isSearchMode {
		mode = fmt.Sprintf(file+" :: [searching: %s (lines: %d)] :: [forward search: n] [backward search: N] [exit search: ESC/Ctrl-C]", p.searchStr, p.ignoreY)
	} else if p.isSlashOn {
		mode = fmt.Sprintf(file+" :: [input search string: %s ]", p.searchStr)
	} else if file != "" {
		mode = file
	}
	if len(p.Files) > 1 {
		nextFileUsage = "[next file: Ctrl-h,Ctrl-l]"
	}
	searchIndex := fmt.Sprintf("(Index: %d)", p.searchIndex)
	p.drawLine(0, 0, "USAGE [exit: ESC/q] [scroll: j,k/C-n,C-p] "+searchIndex+nextFileUsage+mode+string(empty), false)
	termbox.Flush()
}

func (p *Pager) viewModeKey(ev termbox.Event) int {
	switch ev.Key {
	case termbox.KeyEsc, termbox.KeyCtrlC:
		termbox.Flush()
		return 0
	case termbox.KeyArrowRight, termbox.KeyCtrlL:
		if p.isMaxIndex() {
			p.Index--
		} else {
			termbox.Sync()
		}
		termbox.Flush()
		return 1
	case termbox.KeyArrowLeft, termbox.KeyCtrlH:
		if p.Index >= 1 {
			p.Index -= 2
			termbox.Sync()
			return 1
		}
		p.Draw()
	case termbox.KeyCtrlN, termbox.KeyArrowDown, termbox.KeyEnter:
		p.scrollDown()
	case termbox.KeyCtrlP, termbox.KeyArrowUp:
		p.scrollUp()
	case termbox.KeyCtrlD, termbox.KeySpace:
		matched := regexp.MustCompile("(?s)\\n").FindAllString(p.str, -1)
		_, y := termbox.Size()
		if p.ignoreY+29 < (len(matched) - y) {
			p.ignoreY += 29
		} else {
			p.ignoreY = len(matched) - y
		}
		p.Draw()
	case termbox.KeyCtrlU:
		p.ignoreY -= 29
		if p.ignoreY < 0 {
			p.ignoreY = 0
		}
		p.Draw()
	default:
		if ev.Ch == 'j' {
			p.scrollDown()
		} else if ev.Ch == 'k' {
			p.scrollUp()
		} else if ev.Ch == 'q' {
			termbox.Sync()
			return 0
		} else if ev.Ch == '<' {
			p.ignoreY = 0
			return 1
		} else if ev.Ch == '>' {
			matched := regexp.MustCompile("(?s)\\n").FindAllString(p.str, -1)
			_, y := termbox.Size()
			p.ignoreY = len(matched) - y
			println(p.ignoreY)
			termbox.Sync()
			p.Draw()
		} else if ev.Ch == '/' {
			p.isSlashOn = true
			p.isSearchMode = false
			p.Draw()
		} else {
			p.Draw()
		}
	}
	return 2
}

func (p *Pager) PollEvent() bool {
	p.Draw()
	for {
		if p.isSlashOn == false {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				ret := p.viewModeKey(ev)
				if ret == 1 {
					return true
				} else if ret == 0 {
					return false
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
				case termbox.KeyDelete, termbox.KeyCtrlD, termbox.KeyBackspace, termbox.KeyBackspace2:
					p.deleteSearchString()
				case termbox.KeyEsc, termbox.KeyCtrlC:
					p.isSearchMode = false
					p.isSlashOn = false
					p.searchStr = ""
				default:
					if ev.Ch == 'q' {
						p.isSearchMode = false
						p.isSlashOn = false
						p.searchStr = ""
					} else if ev.Ch == 'n' {
						p.searchForward()
					} else if ev.Ch == 'N' {
						p.searchBackward()
					} else {
						p.viewModeKey(ev)
					}
				}
			}
			p.Draw()
		} else { // isSlashOn
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyDelete, termbox.KeyCtrlD, termbox.KeyBackspace, termbox.KeyBackspace2:
					p.deleteSearchString()
				case termbox.KeyEnter:
					p.isSearchMode = true
					p.searchIndex = 1
					p.searchForward()
				case termbox.KeyEsc:
					p.isSlashOn = false
					p.searchStr = ""
				default:
					p.searchStr += string(ev.Ch)
				}
			}
			p.Draw()
		}
	}
	return false
}

func (p *Pager) deleteSearchString() {
	if len(p.searchStr) > 0 {
		p.searchStr = p.searchStr[0 : len(p.searchStr)-1]
	}
}

func (p *Pager) searchString() [][]int {
	return regexp.MustCompile("(?mi)^.*"+p.searchStr+".*$").FindAllStringIndex(regexp.MustCompile("\\033\\[\\d+\\[m(.+?)0m").ReplaceAllString(p.str, "$1"), -1)
}

func (p *Pager) searchForward() {
	matched := p.searchString()
	if len(matched) >= p.searchIndex {
		p.ignoreY = p.getLines(p.str[0:matched[p.searchIndex-1][1]]) - 1
		if len(matched) > p.searchIndex {
			p.searchIndex++
		}
	}
}

func (p *Pager) searchBackward() {
	matched := p.searchString()
	if len(matched) > 0 && p.searchIndex > 1 {
		p.searchIndex--
		p.ignoreY = p.getLines(p.str[0:matched[p.searchIndex-1][1]]) - 1
	}
}

func (p *Pager) getLines(s string) (l int) {
	lines := regexp.MustCompile("(?m)^.*$").FindAllString(s, -1)
	l = len(lines)
	return
}

func (p *Pager) scrollDown() {
	lines := p.getLines(p.str)
	_, y := termbox.Size()
	if p.ignoreY < lines-y {
		p.ignoreY++
	}
	p.Draw()
}

func (p *Pager) scrollUp() {
	if p.ignoreY > 0 {
		p.ignoreY--
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
