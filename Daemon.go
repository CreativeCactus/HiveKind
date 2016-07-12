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
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"log"
	"os"
	"os/exec"
	"runtime"
)


func main() {
	å()
	//Init, first we need to start a UI node to provide API access
	RunNodeLocal("gameui")//'node' here refers to a client, not njs
	//RunNodeDocker("ui")
		
	// Blocking Server
	masterSock := "/tmp/hivemaster.sock"
	err := os.Remove(masterSock)
	if err==nil {
		fmt.Println("Overwrote existing master sock")
	}	
	
	session, err := net.Listen("unix", masterSock)
	if err != nil {
		log.Fatal("Write: ", err)
	}
	for {
		conn, _ := session.Accept()
		go handleComms(conn)
	}
}

var NodePool []Node

func RunNodeLocal(typ string) {
	template, err := GetNodeTemplate(typ)
	if err != nil {
		ə(err, "GetTemplate")
		return
	}
	method, err := GetInitMethod(template.method)
	if err != nil {
		ə(err, "GetMethod")
		return
	}
	node, err := method.f(template.data)
	if err != nil {
		ə(err, "RunTemplate")
		return
	}
	go func() {
		for node.stdout.Scan() {
			fmt.Printf("%s.STDOUT: %s\n", node.id, node.stdout.Text())
		}
	}()
	go func() {
		for node.stderr.Scan() {
			fmt.Printf("%s.STDERR: %s\n", node.id, node.stderr.Text())
		}
	}()
}

func GetNodeTemplate(typ string) (t template, e error) {
	//some inbuilt files need to be loaded. move this to an init function later.
	//For now we will enjoy auto-reload
	clientjs, err := ioutil.ReadFile("./client.js")
	if err != nil {
		ə(err, "InitRead1Err")
		return
	}
	
	INBUILT := map[string]template{
		"gameui": template{
			method: "nodepipe",
			data:   string(clientjs),
		},
	}
	t, ok := INBUILT[typ]
	if ok {
		return
	}
	//attempt to get the type from some db or a db node

	//finally give up
	return template{}, errors.New("Could not find node type: " + typ)
}

//An initialiser for a node
type template struct {
	method string
	data   string
}

func GetInitMethod(method string) (m meth, e error) {
	INBUILT := map[string]meth{
		"nodepipe": meth{
			f: func(data string) (Node, error) {
				rid := grid()
				cmd := exec.Command(`node`)
				stdin, scanout, scanerr, err := CmdToPipes(cmd)
				if err != nil {
					ə(err, "PiperErr")
					return Node{}, err
				}

				if err := cmd.Start(); err != nil {
					ə(err, "StartErr")
					return Node{}, err
				}
				if i, err := stdin.Write([]byte(data + "\n")); err!=nil {
					ə(err, "InputErr")
					return Node{}, err
				} else {
					print(i," bytes sent to nodepipe "+rid+"\n")
					stdin.Close()
				}
				return Node{
					id:     rid,
					cmd:    cmd,
					stdin:  stdin,
					stdout: scanout,
					stderr:scanerr,
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
	return meth{}, errors.New("Could not find method: " + method)
}


//A method to handle the data of a template.
type meth struct {
	f func(string) (Node, error)
}

type Node struct {
	id     string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	stderr *bufio.Scanner
}

func CmdToPipes(cmd *exec.Cmd)(sin io.WriteCloser, sout, serr *bufio.Scanner, e error){
		stdin, err := cmd.StdinPipe()
		if err != nil {
			ə(err, "PipeInErr")
			return sin,sout,serr, err
		}
		
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			ə(err, "PipeOutErr")
			return sin,sout,serr, err
		}
		scanout := ß(stdout)
		
		stderr, err := cmd.StderrPipe()
		if err != nil {
			ə(err, "PipeErrErr")
			return sin,sout,serr, err
		}
		scanerr := ß(stderr)
		return stdin,scanout,scanerr,nil
}
func grid() string {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(out[:len(out)-1])
}




func handleComms(conn net.Conn) {
	rid := grid()
	fmt.Printf("New connection: %s \n",rid)	
	chanClient:= chanFromConn(conn)
	for in := range chanClient{
		fmt.Printf("%s::%s\n",rid,in)
		if len(in)>0 && in[0]==[]byte("1")[0]{
			//chanClient<-[]byte("nah\n")
			conn.Write([]byte("nah\n"))
		}
	}
}

func PipedConn(conn net.Conn,rid,dest_proto,dest_ipport string){
	target, err := net.DialTimeout(dest_proto, dest_ipport, 500000)
	if err != nil {
		fmt.Printf("%s:%s\n",rid,err.Error())
		conn.Write([]byte(fmt.Sprintf("5XX:%s\n", err.Error())))
		conn.Close()
		return
	}
	Pipe(conn, target, rid)
}

// Pipe creates a full-duplex pipe between the two sockets and transfers data from one to the other.
func Pipe(conn1 net.Conn, conn2 net.Conn, id string) {
	dbg:=true
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






















//compose a+a å
//debug to print calling line
func å() {
	fmt.Printf("@:%s\n", ſ(1))
}

//compose e+e ə
//throw a standard error
func ə(e error, lbl string) {
	fmt.Printf("%s[%s]:%s\n", ſ(1), lbl, e.Error())
}

//compose f+s ſ
//Get the name of the caller func
func ſ(back int) string {
	pc, _, ln, _ := runtime.Caller(back + 1)
	return fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), ln)
}

//compose s+s ß
//get a scanner for the given stream
func ß(stream io.ReadCloser) *bufio.Scanner{
	return bufio.NewScanner(stream)
}