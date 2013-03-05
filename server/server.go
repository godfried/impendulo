package server

import( 
"strconv"
"fmt"
"net"
"container/list"
"strings"
"io"
"io/ioutil"
"intlola/client"
"bytes"
)


func Write(fname string, data *bytes.Buffer) error{
	Log("Writing to: ", fname)
	err := ioutil.WriteFile(fname, data.Bytes(), 0666)
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
func FileReader(c *client.Client) {
	buffer := new(bytes.Buffer)
	p := make([]byte, 2048)
	bytesRead, err := c.Conn.Read(p)
	for err == nil {
		buffer.Write(p[:bytesRead])
		bytesRead, err = c.Conn.Read(p)
	}
	if err != io.EOF{
		Log("Client ", c.Token, " resulted in unexpected error: ", err) 
	}
	Log("Reader stopped for ", c.Token)
	err = Write(strconv.Itoa(int(c.Token))+c.File, buffer)
	if err != nil{
		Log(c.Token, " had write error: ", err)
	} else{
		Log("Successfully wrote for: " , c.Token)
	}
	Remove(c)
}


func ConnHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, error := conn.Read(buffer)
	if error != nil {
		Log("Client connection error: ", error)
		return
	}
	req := strings.TrimSpace(string(buffer[:bytesRead]))
	if !strings.HasPrefix(req, "CONNECT"){
		Log("Invalid connection request: "+req)
		return
	}
	token  := getToken()
	fname := strings.TrimSpace(req[strings.Index(req, ":")+1:])
	newClient := client.NewClient(token, conn, fname)
	Log("Connected to new client ", token, " sending ", fname, " on ", conn.RemoteAddr())
	conn.Write([]byte("ACCEPT"))
	clientList.PushBack(*newClient)
	FileReader(newClient)
}

var tokens []byte = make([] byte, 1000) 
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
					go ConnHandler(connection)
				}
			}
		}
	}
}

