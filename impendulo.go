//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

//Package impendulo is provides storage and analysis for code snapshots.
//It receives code snapshots via TCP or a web upload, runs analysis tools and tests on
//them and provides a web interface to view the results in.
package main

import (
	"flag"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/server"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/webserver"
	"runtime"
)

//Flag variables for setting ports to listen on, users file to process, mode to run in, etc.
var (
	Port, UsersFile, ConfigFile, ErrorLogging, InfoLogging, Backup string
	Web, Receiver, Processor, Debug                                bool
	MaxProcs, Timeout                                              int
	conn                                                           string
)

const (
	LOG_IMPENDULO = "impendulo.go"
)

func init() {
	//Setup flags
	flag.IntVar(&Timeout, "t", 5, "Specify the time limit for a tool to run in, in minutes (default 10).")
	flag.IntVar(&MaxProcs, "mp", 5, "Specify the maximum number of goroutines to run when processing submissions (default 10).")
	flag.BoolVar(&Web, "w", true, "Specify whether to run the webserver (default true).")
	flag.BoolVar(&Receiver, "r", true, "Specify whether to run the Intlola file receiver (default true).")
	flag.BoolVar(&Processor, "s", true, "Specify whether to run the Intlola file processor (default true).")
	flag.BoolVar(&Debug, "d", true, "Specify whether to run in debug mode (default true).")
	flag.StringVar(&Backup, "b", "", "Specify a backup db (default none).")
	flag.StringVar(&ErrorLogging, "e", "a", "Specify where to log errors to (default console & file).")
	flag.StringVar(&InfoLogging, "i", "f", "Specify where to log info to (default file).")
	flag.StringVar(&Port, "p", "8010", "Specify the port to listen on for files.")
	flag.StringVar(&UsersFile, "u", "", "Specify a file with new users.")
	flag.StringVar(&ConfigFile, "c", config.DefaultConfig(), "Specify a configuration file.")
}

func main() {
	//Set the number of processors to use to the number of available CPUs.
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	//Handle setup flags
	if Backup != "" {
		err := backupDB()
		if err != nil {
			util.Log(err, LOG_IMPENDULO)
			return
		}
	}
	err := setupConn(Debug)
	if err != nil {
		util.Log(err, LOG_IMPENDULO)
		return
	}
	defer db.Close()
	util.SetErrorLogging(ErrorLogging)
	util.SetInfoLogging(InfoLogging)
	tool.SetTimeout(Timeout)
	err = config.LoadConfigs(ConfigFile)
	if err != nil {
		util.Log(err, LOG_IMPENDULO)
		return
	}
	if UsersFile != "" {
		AddUsers()
	}
	//Run various processes if specified.
	if Web {
		RunWebServer(Receiver || Processor)
	}
	if Processor {
		RunFileProcessor(Receiver)
	}
	if Receiver {
		RunFileReceiver(false)
	}
}

//backupDB backs up the default database to a specified backup.
func backupDB() (err error) {
	err = db.Setup(db.DEFAULT_CONN)
	if err != nil {
		return
	}
	defer db.Close()
	err = db.CopyDB(db.DEFAULT_DB, Backup)
	return
}

//setupConn sets up the database connection
func setupConn(debug bool) (err error) {
	if debug {
		conn = db.DEBUG_CONN
	} else {
		conn = db.DEFAULT_CONN
	}
	err = db.Setup(conn)
	return
}

//AddUsers adds users from a text file to the database.
func AddUsers() {
	users, err := user.Read(UsersFile)
	if err != nil {
		util.Log(err, LOG_IMPENDULO)
		return
	}
	err = db.Setup(conn)
	if err != nil {
		util.Log(err, LOG_IMPENDULO)
		return
	}
	err = db.AddUsers(users...)
	if err != nil {
		util.Log(err, LOG_IMPENDULO)
	}

}

//RunWebServer runs the webserver
func RunWebServer(inRoutine bool) {
	if inRoutine {
		go webserver.Run()
	} else {
		webserver.Run()
	}
}

//RunFileReceiver runs the TCP file receiving server.
func RunFileReceiver(inRoutine bool) {
	if inRoutine {
		go server.Run(Port, new(server.SubmissionSpawner))
	} else {
		server.Run(Port, new(server.SubmissionSpawner))
	}
}

//RunFileProcessor runs the file processing server.
func RunFileProcessor(inRoutine bool) {
	if inRoutine {
		go processing.Serve(MaxProcs)
	} else {
		processing.Serve(MaxProcs)
	}
}
