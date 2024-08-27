package main

import (
	"fmt"
	"net/http"

	_ "github.com/reiver/roodmagi/api"
	"github.com/reiver/roodmagi/srv/http"
	. "github.com/reiver/roodmagi/srv/log"
)

func main() {
	Log("-<([ hello world ])>-")
	Log()
	Log("roodmagi")
	Log("rôd maguš")
	Log("𐎼𐎢𐎫 𐎶𐎦𐎢𐏁")
	Log()

	var tcpport string = tcpPort()
	Logf("tcp-port = %q", tcpport)

	var addr string = fmt.Sprintf(":%s", tcpport)
	Logf("tcp-address = %q", addr)

	var handler http.Handler = &httpsrv.Mux

	{
		Log()
		Log("Here we go…")
		err := http.ListenAndServe(addr, handler)
		if nil != err {
			Logf("ERROR: HTTP server had problem listening-and-serving: %s", err)
			return
		}
		Log("beware i live")
	}
}
