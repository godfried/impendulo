//Package server provides a TCP server which receives user submissions and
//snapshots and sends them to be processed.
package server

import (
	"fmt"
	"github.com/godfried/impendulo/util"
	"io"
	"net"
)

const (
	OK                  = "ok"
	SEND                = "send"
	LOGIN               = "begin"
	LOGOUT              = "end"
	REQ                 = "req"
	PROJECTS            = "projects"
	SUBMISSION_NEW      = "submission_new"
	SUBMISSION_CONTINUE = "submission_continue"
	LOG_SERVER          = "server/server.go"
)

type (

	//HandlerSpawner is an interface used to spawn RWCHandlers.
	HandlerSpawner interface {
		Spawn() RWCHandler
	}

	//RWCHandler is an interface with basic methods for handling connections.
	RWCHandler interface {
		Start(rwc io.ReadWriteCloser)
		End(err error)
	}
)

//Run is used to listen for new tcp connections and
//spawn a new goroutine for each connection.
//Each goroutine launched will handle its connection and
//its type is determined by HandlerSpawner.
func Run(port string, spawner HandlerSpawner) {
	netListen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		util.Log(fmt.Errorf(
			"Encountered error %q when listening on port %q",
			err, port), LOG_SERVER)
		return
	}
	defer netListen.Close()
	for {
		conn, err := netListen.Accept()
		if err != nil {
			util.Log(fmt.Errorf(
				"Encountered error %q when accepting connection",
				err), LOG_SERVER)
		} else {
			go func(c net.Conn) {
				handler := spawner.Spawn()
				handler.Start(c)
			}(conn)
		}
	}
}
