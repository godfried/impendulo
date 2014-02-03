//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

//Package processing provides functionality for running a submission and its snapshots
//through the Impendulo tool suite.
package processing

import (
	"github.com/godfried/impendulo/util"
)

type (

	//Status is used to indicate a change in the files or
	//submissions being processed. It is also used to retrieve the current
	//number of files and submissions being processed.
	Status struct {
		Files       int
		Submissions int
	}
	Monitor struct {
		statusChan              chan Status
		uri                     string
		loader, waiter, changer *MessageHandler
	}
)

var (
	monitor *Monitor
)

//add adds the value of toAdd to this Status.
func (this *Status) add(toAdd Status) {
	this.Files += toAdd.Files
	this.Submissions += toAdd.Submissions
}

//MonitorStatus keeps track of Impendulo's current processing status.
func MonitorStatus(amqpURI string) (err error) {
	if monitor != nil {
		err = monitor.Stop()
		if err != nil {
			return
		}
	}
	monitor, err = NewMonitor(amqpURI)
	if err != nil {
		return
	}
	go monitor.Monitor()
	return
}

func NewMonitor(amqpURI string) (ret *Monitor, err error) {
	ret = &Monitor{
		statusChan: make(chan Status),
		uri:        amqpURI,
	}
	ret.changer, err = NewChanger(ret.uri, ret.statusChan)
	if err != nil {
		return
	}
	ret.waiter, err = NewWaiter(ret.uri, ret.statusChan)
	if err != nil {
		return
	}
	ret.loader, err = NewLoader(ret.uri, ret.statusChan)
	return
}

func (this *Monitor) Monitor() {
	handle := func(mh *MessageHandler) {
		merr := mh.Handle()
		if merr != nil {
			util.Log(merr, mh.Shutdown())
		}
	}
	go handle(this.changer)
	go handle(this.loader)
	go handle(this.waiter)
	var status *Status = new(Status)
	for val := range this.statusChan {
		switch val {
		case Status{}:
			//A zeroed Status indicates a request for the current Status.
			this.statusChan <- *status
		default:
			status.add(val)
			util.Log(*status, val)
		}
	}
}

func StopMonitor() (err error) {
	if monitor == nil {
		return
	}
	err = monitor.Stop()
	if err == nil {
		monitor = nil
	}
	return
}

func (this *Monitor) Stop() (err error) {
	close(this.statusChan)
	err = this.shutdownHandlers()
	return
}

func (this *Monitor) shutdownHandlers() (err error) {
	err = this.waiter.Shutdown()
	if err != nil {
		return
	}
	err = this.loader.Shutdown()
	if err != nil {
		return
	}
	err = this.changer.Shutdown()
	return
}
