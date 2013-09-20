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

package webserver

import (
	"bytes"
	"net/http"
	"sync"
)

type (
	//Buffer is a http.ResponseWriter which buffers all the data and headers.
	HttpBuffer struct {
		bytes.Buffer
		resp    int
		headers http.Header
		once    sync.Once
	}
)

//Header implements the Header method of http.ResponseWriter
func (this *HttpBuffer) Header() http.Header {
	this.once.Do(func() {
		this.headers = make(http.Header)
	})
	return this.headers
}

//WriteHeader implements the WriteHeader method of http.ResponseWriter
func (this *HttpBuffer) WriteHeader(resp int) {
	this.resp = resp
}

//Apply takes an http.ResponseWriter and calls the required methods on it to
//output the buffered headers, response code, and data. It returns the number
//of bytes written and any errors flushing.
func (this *HttpBuffer) Apply(w http.ResponseWriter) (n int, err error) {
	if len(this.headers) > 0 {
		h := w.Header()
		for key, val := range this.headers {
			h[key] = val
		}
	}
	if this.resp > 0 {
		w.WriteHeader(this.resp)
	}
	n, err = w.Write(this.Bytes())
	return
}
