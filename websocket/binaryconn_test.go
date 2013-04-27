// Copyright 2013 Gary Burd & Zhang Peihao
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package websocket_test

import (
	"github.com/zhangpeihao/go-websocket/websocket"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type wsBinaryHandler struct {
	*testing.T
}

func (t wsBinaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		t.Logf("bad method: %s", r.Method)
		return
	}
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		t.Logf("bad origin: %s", r.Header.Get("Origin"))
		return
	}
	conn, err := websocket.NewBianryConn(w, r, http.Header{"Set-Cookie": {"sessionId=1234"}}, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		t.Logf("bad handshake: %v", err)
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		t.Logf("upgrade error: %v", err)
		return
	}
	defer conn.Close()
	for {
		b, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Logf("ioutil.ReadAll(conn)) error: %v", err)
			return
		}

		if _, err = conn.Write(b); err != nil {
			t.Logf("conn.Write error: %v", err)
			return
		}
	}
}

func TestBinaryConn(t *testing.T) {
	s := httptest.NewServer(wsBinaryHandler{t})
	defer s.Close()
	conn, resp, err := websocket.Connect(s.URL, 1024, 1024)
	if err != nil {
		t.Fatal("Connect err:", err)
	}

	defer conn.Close()

	var sessionId string
	for _, c := range resp.Cookies() {
		if c.Name == "sessionId" {
			sessionId = c.Value
		}
	}
	if sessionId != "1234" {
		t.Error("Set-Cookie not received from the server.")
	}

	if _, err = conn.Write([]byte("HELLO")); err != nil {
		t.Error("Write err:", err)
	}

	time.Sleep(time.Second)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	b, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(b) != "HELLO" {
		t.Fatalf("message=%s, want %s", b, "HELLO")
	}
}
