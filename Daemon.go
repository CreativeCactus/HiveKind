/// A master node daemon (capable of reporting as slave to a higher master)
// This will provide only management for underlying clients.
// At least one local client will always be running
// Client init involves setting up a filtered fwd.go channel to a socket file which will be accessed
// solely by the client. Client init will exist in a temporary directory.
// All client content, associated files and configuration will exist in structured JSON
// in a config file, using arbitrary compression on a per-field basis.
// Additionally there may be an associated zip or vfs with parts of a filesystem inside
//
// Client can be dumped and initialised via system call, or may be sent as a pure JS payload
// to browsers. In this way many temporary nodes can easily be set up simply by navigating to
// an active node.
// Initialised clients call back to the master via their socket and await instructions.
// At this point the master can send instructions to any or all clients with the corresponding instruction enabled.
// The daemon is typically not interfaced directly and will negotiate via a single local node
// The local node is authenticated via PKI, and allows listing of entities and their arbitrary properties
// whicha are then rendered in HTML or CLI depending on headers (nc works)

// one necessary function will be to lialise with the node pool to perform multi-part tasks such as generating proxy chains
// an important goal is to have dev environments (such as for compiling GO) generated in dockers
// The polyglot interface node should then determine the content type and either:
//  A) Find a node capable of executing the appropriate action and return the response
//  B) Generate a context with the request and return a list of options available
// For example, the following should be possible:
// # tar ./ADirFullOfGoCode - | nc 127.0.0.1:13337
// Returning:
// // CONTEXT:123456789 [Build_and_dump(path), Build_and_return([address]), Build_and_run(), Fmt(), Cancel()]
// The context is recorded with the requester's IP, and if no context ID is in the reply, the most recent from this IP is used
// small requests like ctx should reveal the visible contexts
// # echo "build and return" | nc 127.0.0.1:13337

// Each stored node template is in itself capable of carrying out commands depending on the deployment context (browser, for ex)
package main

import (
	"HiveKind/hk"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"

	"github.com/nsf/termbox-go"
)

//TODO: tag nodes and visualise templates and nodes inside a ui

var ROOT hk.FolderNode
var Logs = &hk.MsgQue{ID: "Logs", Messages: []string{"Initialised..."}, ViewOpen: false}

func main() {
	//å is a debug helper which will exemplify logging from within
	å()
	//Some of the node actions we expect to be able to perform
	n := RunNodeLocal("ui") //'node' here refers to a client, not njs
	//RunNodeDocker("ui")
	//RunNodeSsh()

	//Set up our root node in the tree
	ROOT = hk.FolderNode{
		Nodes: []hk.Entry{Logs, n},
	}

	//Set up a local sock to act as the comms bus
	masterSock := "/tmp/hivemaster.sock" //TODO: lock and defer delete
	err := os.Remove(masterSock)
	if err == nil {
		fmt.Println("Overwrote existing master sock")
	}
	session, err := net.Listen("unix", masterSock)
	if err != nil {
		log.Fatal("Write: ", err)
	}

	// Blocking Server, delegate incoming connections to handleComms
	go func() {
		for {
			conn, _ := session.Accept()
			go handleComms(conn)
		}
	}()

	TerminalInterface()
}

func handleComms(conn net.Conn) {
	rid := grid()
	ł(fmt.Sprintf("New connection: %s\n", rid))
	chanClient := chanFromConn(conn)
	for in := range chanClient {
		ł(fmt.Sprintf("%s::%s\n", rid, in))

		if len(in) > 0 {

			//chanClient<-[]byte("nah\n")
			conn.Write([]byte("nah\n"))
		}
	}
}

/*
	Display stuff
*/

//Cursor position
var cx = 0
var cy = 0

//Terminal size
var tx = 0
var ty = 0

//Virtual offset
var vx = 0
var vy = 0

//TerminalInterface inits termbox and handles input
func TerminalInterface() {
	//Sadly termbox cannot be injected.
	err := termbox.Init()
	termbox.SetCursor(0, 0)
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	//Terminal size
	tx, ty = termbox.Size()

	//First draw will have an event pointer nil, thus only draw to first level
	//we can provide an empty event to avoid this effect.
	var e *termbox.Event

	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		List(e, ROOT.Children(), 1, 0, false)
		termbox.Flush()
		ev := termbox.PollEvent()
		e = &ev
		CursorControl(*e)

		if e.Ch == '?' { //help
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			Disp("Example.", 1, 1, termbox.AttrBold+termbox.ColorRed, termbox.ColorDefault)
		}

		if e.Ch == 'c' {
			inBuffer := ""
			for ev.Key != termbox.KeyEnter {
				ev = termbox.PollEvent()
				if ev.Ch != ' ' {
					inBuffer += string(ev.Ch)
				}

			}
			Disp(inBuffer, cx, cy, termbox.ColorRed, termbox.ColorCyan)
		}

		if e.Key == termbox.KeyEsc || e.Key == termbox.KeyCtrlC {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			Disp("Goodbye.", 3, 7, termbox.ColorDefault, termbox.ColorDefault)
			return
		}
	}
}

//List will draw the entire entity tree, then return the visible section
//todo: redraw every n sec
func List(e *termbox.Event, entries []hk.Entry, xoff, yoff int, recursing bool) int {
	for _, v := range entries {
		if yoff >= vy {
			yindex := yoff - vy
			if cy == yindex && e != nil && e.Key == termbox.KeyEnter {
				(v).Toggle()
			}

			Disp("⛧ ", xoff, yindex, termbox.ColorDefault, termbox.ColorDefault)
			str, fg, bg := (v).Title()
			Disp(str, xoff+2, yindex, fg, bg)

			if childs := (v).Children(); len(childs) > 0 {
				yoff += List(e, childs, xoff+2, yoff+1, true) - 1
			}
		}
		yoff++
	}

	if recursing {
		return yoff
	}

	//Make sure our cursor is inside the list
	cy = hk.Lim(0, cy, yoff-1)
	termbox.SetCursor(cx, cy)
	return 0
}

//Disp will print a string to the screen
func Disp(msg string, x, y int, fg, bg termbox.Attribute) {
	xo := 0
	for _, v := range msg {
		if v == '\n' {
			xo = 0
		} else {
			termbox.SetCell(x+xo, y, v, fg, bg)
			if (x + xo) < tx {
				xo++
			} else {
				termbox.SetCell(tx, y, '…', fg, bg)
				continue
			}

		}
	}
}

//CursorControl handles navigation
func CursorControl(e termbox.Event) {
	switch e.Key {
	case termbox.KeyArrowUp:
		cy--
		break
	case termbox.KeyArrowDown:
		cy++
		break
		// case termbox.KeyArrowLeft:
		// 	cx--
		// 	break
		// case termbox.KeyArrowRight:
		// 	cx++
		// 	break
	}

	if e.Width != 0 {
		tx = e.Height - 1
	}
	if e.Height != 0 {
		ty = e.Height - 1
	}

	cx = hk.Lim(0, cx, tx)
	cy = hk.Lim(0, cy, ty)
	termbox.SetCursor(cx, cy)
}

/*
	Runner helpers
*/

func RunNodeLocal(typ string) *hk.Node {
	template, err := GetNodeTemplate(typ)
	if err != nil {
		ə(err, "GetTemplate")
		return nil
	}
	method, err := GetInitMethod(template.Method)
	if err != nil {
		ə(err, "GetMethod")
		return nil
	}
	node, err := method.F(template.Data)
	if err != nil {
		ə(err, "RunTemplate")
		return nil
	}
	return &node
}

func GetNodeTemplate(typ string) (t hk.Template, e error) {
	//some inbuilt files need to be loaded. move this to an init function later.
	//For now we will enjoy auto-reload
	clientjs, err := ioutil.ReadFile("./client.js")
	if err != nil {
		ə(err, "InitRead1Err")
		return
	}

	INBUILT := map[string]hk.Template{
		"ui": hk.Template{
			Method: "nodepipe",
			Data:   string(clientjs),
		},
	}
	t, ok := INBUILT[typ]
	if ok {
		return
	}

	//attempt to get the type from some db or a db node

	//finally give up
	return hk.Template{}, errors.New("Could not find node type: " + typ)
}

//An initialiser for a node
func GetInitMethod(method string) (m hk.Meth, e error) {
	INBUILT := map[string]hk.Meth{
		"nodepipe": hk.Meth{
			F: func(data string) (hk.Node, error) {
				rid := grid()
				cmd := exec.Command(`node`)
				stdin, scanout, scanerr, err := CmdToPipes(cmd)
				if err != nil {
					ə(err, "PiperErr")
					return hk.Node{}, err
				}

				if err := cmd.Start(); err != nil {
					ə(err, "StartErr")
					return hk.Node{}, err
				}
				if i, err := stdin.Write([]byte(data + "\n")); err != nil {
					ə(err, "InputErr")
					return hk.Node{}, err
				} else {
					print(i, " bytes sent to nodepipe "+rid+"\n")
					stdin.Close()
				}

				myStdio := hk.STDIO{Stdin: stdin}
				//Read values as they come in and send them to a buffer
				go func(myStdio *hk.STDIO) {
					for scanerr.Scan() {
						myStdio.Stdout = append(myStdio.Stdout, hk.Label{
							Text: scanerr.Text(),
							Tag:  "STDERR",
							Fg:   termbox.ColorRed})
					}
				}(&myStdio)
				go func(myStdio *hk.STDIO) {
					for scanout.Scan() {
						myStdio.Stdout = append(myStdio.Stdout, hk.Label{
							Text: scanout.Text(),
							Tag:  "STDOUT",
							Fg:   termbox.ColorGreen})
					}
				}(&myStdio)

				return hk.Node{
					ID:       rid,
					Cmd:      cmd,
					Stdio:    &myStdio,
					ViewOpen: false,
				}, nil
			},
		},
	}
	m, ok := INBUILT[method]
	if ok {
		return
	}

	//attempt to get the method from some db or a db node

	//finally give up
	return hk.Meth{}, errors.New("Could not find method: " + method)
}

func CmdToPipes(cmd *exec.Cmd) (sin io.WriteCloser, sout, serr *bufio.Scanner, e error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		ə(err, "PipeInErr")
		return sin, sout, serr, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ə(err, "PipeOutErr")
		return sin, sout, serr, err
	}
	scanout := ß(stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		ə(err, "PipeErrErr")
		return sin, sout, serr, err
	}
	scanerr := ß(stderr)
	return stdin, scanout, scanerr, nil
}

/*
	Connection helpers
*/

//PipedConn will pipe a connection to an endpoint
func PipedConn(conn net.Conn, rid, destProto, destIPPort string) {
	target, err := net.DialTimeout(destProto, destIPPort, 500000)
	if err != nil {
		fmt.Printf("%s:%s\n", rid, err.Error())
		conn.Write([]byte(fmt.Sprintf("5XX:%s\n", err.Error())))
		conn.Close()
		return
	}
	Pipe(conn, target, rid)
}

//Pipe creates a full-duplex pipe between the two sockets and transfers data from one to the other.
func Pipe(conn1 net.Conn, conn2 net.Conn, id string) {
	dbg := true
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)
	close := func() {
		conn1.Close()
		conn2.Close()
	}
	for {
		select {
		case b1 := <-chan1:
			if dbg {
				fmt.Printf(id+":Client: %s [eof?%v]\n", b1, b1 == nil)
			}
			if b1 == nil {
				close()
				return
			}
			conn2.Write(b1)
		case b2 := <-chan2:
			if dbg {
				fmt.Printf(id+":Server: %s [eof?%v]\n", b2, b2 == nil)
			}
			if b2 == nil {
				close()
				return
			}
			conn1.Write(b2)
		}
	}
}

// chanFromConn creates a channel from a Conn object, and sends everything it
//  Read()s from the socket to the channel.
func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)
	go func() {
		b := make([]byte, 1024)
		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()

	return c
}

/*
	Tiny helpers
*/

func grid() string {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out[:len(out)-1])
}

//compose a+a å
//debug to print calling line
func å() {
	ł(fmt.Sprintf("@:%s\n", ſ(1)))
}

//compose e+e ə
//throw a standard error
func ə(e error, lbl string) {
	ł(fmt.Sprintf("%s[%s]:%s\n", ſ(1), lbl, e.Error()))
}

//compose f+s ſ
//Get the name of the caller func
func ſ(back int) string {
	pc, _, ln, _ := runtime.Caller(back + 1)
	return fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), ln)
}

//compose s+s ß
//get a scanner for the given stream
func ß(stream io.ReadCloser) *bufio.Scanner {
	return bufio.NewScanner(stream)
}

//compose /+l ł
//print a string to the logs, rather than printf
func ł(s string) {
	Logs.Add(s)
}
