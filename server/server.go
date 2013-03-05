package server

import( "fmt"
"net"
"container/list"
"strconv"
"strings"
"io"
"io/ioutil"
"intlola/client")


func Write(fname string, data []byte) error{
	Log("data: "+string(data))
	err := ioutil.WriteFile(fname, data, 0666)
	return err
}
func Log(v ...interface{}) {
	fmt.Println(v...)
}


func Remove(c *client.Client) {
	c.Close()
	for entry := clientList.Front(); entry != nil; entry = entry.Next() {
		cur := entry.Value.(client.Client)
		if c.Equal(&cur) {
			Log("Removed: ", c.Token)
			clientList.Remove(entry)
			tokens = append(tokens, c.Token)
			break
		}
	}
}

func IOHandler(Incoming <-chan string) {
	for {
		Log("IOHandler: Waiting for input")
		input := <-Incoming
		for e := clientList.Front(); e != nil; e = e.Next() {
			c := e.Value.(client.Client)
			c.Incoming <-input
		}
	}
}

func FileReader(c *client.Client) {
	buffer := make([]byte, 2048)
	bytesRead, err := c.Conn.Read(buffer)
	for err == nil {
		data := buffer[:bytesRead]
		c.AddData(data)
		bytesRead, err = c.Conn.Read(buffer)
	}
	if err != io.EOF{
		Log("Client ", c.Token, " resulted in unexpected error: ", err) 
	}
	Log("Reader stopped for ", c.Token)
	err = Write(c.File, c.GetData())
	if err != nil{
		Log(c.Token, " had write error: ", err)
	}
	Remove(c)
}


func ConnHandler(conn net.Conn, ch chan string) {
	buffer := make([]byte, 1024)
	bytesRead, error := conn.Read(buffer)
	if error != nil {
		Log("Client connection error: ", error)
	}
	req := strings.TrimSpace(string(buffer[:bytesRead]))
	if req != "CONNECT"{
		Log("Invalid connection request: "+req)
		return
	}
	token  := getToken()
	name := strconv.Itoa(int(token))
	conn.Write([]byte("accepted connection with token: "+name))
	bytesRead, error = conn.Read(buffer)
	if error != nil {
		Log("File name retrieval error: ", error)
	}
	fname := string(buffer[:bytesRead])
	newClient := client.NewClient(token, ch, conn, fname)
	go FileReader(newClient)
	clientList.PushBack(*newClient)
}

var tokens []byte = make([] byte, 100) 
var clientList *list.List = list.New()
func init(){
	for i,_ := range tokens{
		tokens[i] = byte(i)
	}
}
func getToken()(byte){
	ret := tokens[0]
	tokens = tokens[1:]
	return ret
}

func Run(address string, port string){
	Log("Server Started")
	in := make(chan string)
	go IOHandler(in)
	service := address+":"+port
	tcpAddr, error := net.ResolveTCPAddr("tcp", service)
	if error != nil {
		Log("Error: Could not resolve address")
	} else {
		netListen, error := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if error != nil {
			Log(error)
		} else {
			defer netListen.Close()

			for {
				Log("Waiting for connections")
				connection, error := netListen.Accept()
				if error != nil {
					Log("Client error: ", error)
				} else {
					go ConnHandler(connection, in)
				}
			}
		}
	}
}

