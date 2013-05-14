package main

import (
	"errors"
	"flag"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/server"
	"github.com/godfried/cabanga/utils"
	"log"
)

var port, address, users, mode string

func init() {
	flag.StringVar(&port, "p", "9000", "Specify the port to listen on.")
	flag.StringVar(&address, "a", "0.0.0.0", "Specify the address.")
	flag.StringVar(&users, "u", "", "Specify a file with new users.")
	flag.StringVar(&mode, "m", "s", "Specify a mode to run in.")
}

func main() {
	flag.Parse()
	if mode == "u" {
		err := AddUsers(users)
		if err != nil {
			utils.Log("Adding users encountered error ", err)
		}
	} else if mode == "s" {
		runServer(address, port)
	} else {
		log.Fatal(errors.New("Unknown running mode: "), mode)
	}
}

func AddUsers(fname string) error {
	users, err := utils.ReadUsers(fname)
	if err == nil {
		err = db.AddMany(db.USERS, users...)
	}
	return err
}

func runServer(addr, port string) {
	utils.Log("Starting server at: ", address, " on port ", port)
	server.Run(addr, port)
}