package verboten

import (
	"net/http"
	"net/url"
	"time"

	"github.com/reiver/go-erorr"
	"github.com/reiver/go-errhttp"
	"github.com/reiver/go-httpsse"
	"github.com/reiver/go-mstdn/api/v1/streaming/public/local"

	"github.com/reiver/rodmagus/srv/http"
)

const path string = "/v1/streaming/query"

func init() {
	var handler http.Handler = http.HandlerFunc(serveHTTP)

	err := httpsrv.Mux.HandlePath(handler, path)
	if nil != err {
		e := erorr.Errorf("problem registering http-handler with path-mux for path %q: %w", path, err)
		panic(e)
	}
}

func serveHTTP(responsewriter http.ResponseWriter, request *http.Request) {

	if nil == responsewriter {
		return
	}

	if nil == request {
		errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
		return
	}

	var method string = request.Method

	if http.MethodGet != method {
		errhttp.ErrHTTPMethodNotAllowed.ServeHTTP(responsewriter, request)
		return
	}

	var query url.Values
	{
		var urloc *url.URL = request.URL
		if nil == urloc {
			errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
			return
		}

		query = urloc.Query()

		if len(query) < 0 {
			errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
			return
		}
	}

	var network string = query.Get("network")
	if "" == network {
		errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
		return
	}

	switch network {
	case "mastodon":
		serveMastodon(responsewriter, request)
	default:
		errhttp.ErrHTTPNotFound.ServeHTTP(responsewriter, request)
		return
	}
}

func serveMastodon(responsewriter http.ResponseWriter, request *http.Request) {

	if nil == responsewriter {
		return
	}

	if nil == request {
		errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
		return
	}

	var query url.Values
	{
		var urloc *url.URL = request.URL
		if nil == urloc {
			errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
			return
		}

		query = urloc.Query()

		if len(query) < 0 {
			errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
			return
		}
	}

	var from string = query.Get("from")
	if "" == from {
		errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
		return
	}

	var route httpsse.Route = httpsse.NewRoute()

	// Send a heartbeart SSE comment every 4.231 seconds.
	httpsse.HeartBeat(4231 * time.Millisecond, route)

	go func(route httpsse.Route){

		var req http.Request = http.Request{
			Method: http.MethodGet,
			URL: &url.URL{
				Scheme:"https",
				Host:from,
				Path:path,
			},
			Header: http.Header{
				"Authorization":[]string{"Bearer k6pjqnnbs9Honli29mF1yRAN5E92Yv6RGGYNomvHi_o"},
			},
		}

		client, err := local.Dial(&req)
		if nil != err {
//			t.Logf("ERROR: %s", err)
			return
		}
		if nil == client {
//			t.Log("ERROR: nil client")
			return
		}

		for client.Next() {

			var event local.Event
			err := client.Decode(&event)
			if nil != err {
//				t.Logf("DECODE-ERROR: (%T) %s", err, err)
				continue
			}

			if err := client.Err(); nil != err {
//				t.Errorf("CLIENT-ERROR: (%T) %s", err, err)
				return
			}
		}
	}(route)

	route.ServeHTTP(responsewriter, request)
}
