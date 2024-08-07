package main

import (
	"fmt"
	"net/http"

	_ "github.com/reiver/rodmagus/api"
	"github.com/reiver/rodmagus/srv/http"
)

func main() {
	log("-<([ hello world ])>-")
	log()
	log("rodmagus")
	log("rôd maguš")
	log("𐎼𐎢𐎫 𐎶𐎦𐎢𐏁")
	log()

	var tcpport string = tcpPort()
	logf("tcp-port = %q", tcpport)

	var addr string = fmt.Sprintf(":%s", tcpport)
	logf("tcp-address = %q", addr)

	var handler http.Handler = &httpsrv.Mux

	{
		log()
		log("Here we go…")
		err := http.ListenAndServe(addr, handler)
		if nil != err {
			logf("ERROR: HTTP server had problem listening-and-serving: %s", err)
			return
		}
	}
}
