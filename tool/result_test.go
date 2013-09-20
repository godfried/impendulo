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

package tool

import (
	"testing"
)

func TestAddCoords(t *testing.T) {
	bad := CreateChart("bad")
	bad["data"] = 5
	AddCoords(bad, 0, 0)
	bad["data"] = nil
	AddCoords(bad, 0, 0)
	AddCoords(nil, 0, 0)
	good := CreateChart("good")
	AddCoords(good, 1000, 1)
	res := good["data"].([]map[string]float64)
	x := res[len(res)-1]["x"]
	expect := 1.0
	if x != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "x", x)
	}
	expect = 1.0
	y := res[len(res)-1]["y"]
	if y != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "y", y)
	}
	AddCoords(good, -1, 5)
	AddCoords(good, -1, 50)
	res = good["data"].([]map[string]float64)
	x = res[len(res)-1]["x"]
	expect = 2.0
	if x != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "x", x)
	}
	expect = 50.0
	y = res[len(res)-1]["y"]
	if y != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "y", y)
	}
}
