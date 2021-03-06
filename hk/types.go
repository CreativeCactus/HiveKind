package hk

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/nsf/termbox-go"
)

/*
	Structs
*/
type ExtensionInterface struct {
	Stdin *bufio.Reader
}
type Action struct {
	Name    string
	Fn      func(*ExtensionInterface)
	History []string
}

func (a *Action) Toggle(ei *ExtensionInterface) {
	a.Fn(ei)
}
func (a *Action) Title() (string, termbox.Attribute, termbox.Attribute) {
	return "F:" + a.Name, termbox.AttrBold + termbox.ColorCyan, termbox.ColorDefault
}
func (a *Action) Children() (wrapped []Entry) {
	for _, v := range a.History {
		wrapped = append(wrapped, &Label{Text: v})
	}
	return
}

type STDIO struct {
	ID       string
	Stdin    *io.WriteCloser
	Stdout   []*Label
	ViewOpen bool
}

func (s *STDIO) Toggle(ei *ExtensionInterface) {
	s.ViewOpen = !s.ViewOpen
}
func (s *STDIO) Title() (string, termbox.Attribute, termbox.Attribute) {
	return "STDIO", termbox.AttrBold, termbox.ColorDefault
}
func (s *STDIO) Children() (wrapped []Entry) {
	if !s.ViewOpen {
		return []Entry{}
	}
	for _, v := range s.Stdout {
		wrapped = append(wrapped, v)
	}

	wrapped = append(wrapped, &Action{
		Name: "Send...",
		Fn: func(ei *ExtensionInterface) {
			buff := []byte{}
			print("\t\t\t")
			for input, _ := ei.Stdin.ReadByte(); input != '\r' && input != '\n'; input, _ = ei.Stdin.ReadByte() {
				buff = append(buff, input)
				print(string(input))
			}
			buff = append(buff, []byte("\r\n")...)
			buff = append(buff, []byte(`
`)...)
			//		(*s.Stdin).Write(buff)
			io.WriteString(*s.Stdin, "ayyyyyyyyyyyyyy\n")
		},
	})
	return
}

//method of turning a string of code into a running Node
type Meth struct {
	F func(string) (*Node, error)
}

//Label represents a string in a wrapper of an entry
type Label struct {
	Text string
	Tag  string
	Fg   termbox.Attribute
}

func (l *Label) Title() (string, termbox.Attribute, termbox.Attribute) {
	return "☞" + l.Text, l.Fg, termbox.ColorDefault
}
func (l *Label) Toggle(ei *ExtensionInterface) {}
func (l *Label) Children() []Entry {
	return []Entry{}
}

//MsgQue represents a line-by-line set of strings
type MsgQue struct {
	ID       string
	Unread   int
	Messages []string
	ViewOpen bool
}

func (q *MsgQue) Title() (string, termbox.Attribute, termbox.Attribute) {
	unread := Min(q.Unread, 10)
	icon := `O➀➁➂➃➄➅➆➇➈➉⊕`[unread : unread+1]
	return "[" + icon + "]" + q.ID + "[msgs]", termbox.ColorDefault, termbox.ColorDefault
}
func (q *MsgQue) Toggle(ei *ExtensionInterface) {
	q.ViewOpen = !q.ViewOpen
	q.Unread = 0
}
func (q *MsgQue) Add(s string) {
	q.Messages = append(q.Messages, s)
	if q.ViewOpen {
		return
	}
	q.Unread++
}
func (q *MsgQue) Children() (wrapped []Entry) {
	if !q.ViewOpen {
		return []Entry{}
	}
	for _, v := range q.Messages {
		wrapped = append(wrapped, &Label{
			Text: v,
		})
	}
	return wrapped
}

//FolderNode represents a list of entries
type FolderNode struct {
	ID    string
	Nodes []Entry
}

func (f *FolderNode) Title() (string, termbox.Attribute, termbox.Attribute) {
	return f.ID + "[fold]", termbox.ColorDefault, termbox.ColorDefault
}
func (f *FolderNode) Toggle(ei *ExtensionInterface) {}
func (f *FolderNode) Children() []Entry {
	return f.Nodes
}

//Node represents an active child process
type Node struct {
	ID       string
	Cmd      *exec.Cmd
	Stdio    *STDIO
	ViewOpen bool
}

func (n *Node) Title() (string, termbox.Attribute, termbox.Attribute) {
	return n.ID + "[node]", termbox.ColorDefault, termbox.ColorDefault
}
func (n *Node) Toggle(ei *ExtensionInterface) {
	n.ViewOpen = !n.ViewOpen
}
func (n *Node) Children() []Entry {
	if !n.ViewOpen {
		return []Entry{}
	}
	return []Entry{
		n.Stdio,
	}
}

//Entry represents a 'file' (node, function, data), or folder
type Entry interface {
	Title() (string, termbox.Attribute, termbox.Attribute)
	Toggle(ei *ExtensionInterface)
	Children() []Entry
}

//Template of a node
type Template struct {
	Method string
	Data   string
}

func Lim(l, v, h int) int {
	return Max(l, Min(v, h))
}
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
