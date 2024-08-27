package verboten

import (
	"net/http"
	"net/url"
	"time"

	"github.com/reiver/go-erorr"
	"github.com/reiver/go-errhttp"
	"github.com/reiver/go-httpsse"
	"github.com/reiver/go-iter"
	"github.com/reiver/go-json"
	"github.com/reiver/go-mstdn/api/v1/streaming/public/local"

	"github.com/reiver/roodmagi/srv/http"
	. "github.com/reiver/roodmagi/srv/log"
)

const path string = "/v1/streaming/query"

func init() {
	var handler http.Handler = http.HandlerFunc(serveHTTP)

	err := httpsrv.Mux.HandlePath(handler, path)
	if nil != err {
		e := erorr.Errorf("problem registering http-handler with path-mux for path %q: %w", path, err)
		Log(e)
		panic(e)
	}
}

func serveHTTP(responsewriter http.ResponseWriter, request *http.Request) {

	if nil == responsewriter {
		Log("nil http.ResponseWriter")
		return
	}

	if nil == request {
		errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
		Log("nil *http.Request")
		return
	}

	var method string = request.Method

	if http.MethodGet != method {
		errhttp.ErrHTTPMethodNotAllowed.ServeHTTP(responsewriter, request)
		Logf("bad HTTP method: %q", method)
		return
	}

	var query url.Values
	{
		var urloc *url.URL = request.URL
		if nil == urloc {
			errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
			Log("nil http.Request.URL")
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
		Log("bad request: empty 'network'")
		return
	}

	switch network {
	case "mastodon":
		serveMastodon(responsewriter, request)
	default:
		errhttp.ErrHTTPNotFound.ServeHTTP(responsewriter, request)
		Logf("unsupported 'network': %q", network)
		return
	}
}

func serveMastodon(responsewriter http.ResponseWriter, request *http.Request) {

	if nil == responsewriter {
		Log("nil http.ResponseWriter")
		return
	}

	if nil == request {
		errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
		Log("nil *http.Request")
		return
	}

	var query url.Values
	{
		var urloc *url.URL = request.URL
		if nil == urloc {
			errhttp.ErrHTTPInternalServerError.ServeHTTP(responsewriter, request)
			Log("nil http.Request.URL")
			return
		}

		query = urloc.Query()

		if len(query) < 0 {
			errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
			Log("bad request: no query parameters")
			return
		}
	}

	var from string = query.Get("from")
	if "" == from {
		errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
		Log("bad request: empty 'from'")
		return
	}

	var remoteAccessToken string = query.Get("remote_access_token")
	if "" == remoteAccessToken {
		errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
		Log("bad request: 'remote_acces_token'")
		return
	}

	var route httpsse.Route = httpsse.NewRoute()

	// Send a heartbeart SSE comment every 4.231 seconds.
	httpsse.HeartBeat(4231 * time.Millisecond, route)

	go func(route httpsse.Route){
		var authorization string = "Bearer "+remoteAccessToken

		var u url.URL = url.URL{
			Scheme:"https",
			Host:from,
			Path:path,
		}

		var req http.Request = http.Request{
			Method: http.MethodGet,
			URL: &u,
			Header: http.Header{
				"Authorization":[]string{authorization},
			},
		}

		client, err := local.Dial(&req)
		if nil != err {
			Logf("error dialing %q: %s", &u, err)
			return
		}
		if nil == client {
			Logf("error dialing %q: nil client", &u)
			return
		}

		var iterator iter.Iterator = client


		for iterator.Next() {

			var event local.Event
			err := iterator.Decode(&event)
			if nil != err {
				Logf("error decoding: %s", &u, err)
				continue
			}

			if err := iterator.Err(); nil != err {
				Logf("client error: %s", &u, err)
				return
			}

			err = route.PublishEvent(func(eventwriter httpsse.EventWriter)error{
				if nil == eventwriter {
					var e error = errNilEventWriter
					Logf("error: %s", e)
					return e
				}

				eventwriter.WriteEvent("mastodon."+event.Name)

				{
					bytes, err := json.Marshal(event.Status)
					if nil != err {
						eventwriter.WriteComment("ERROR: "+err.Error())
						Logf("error trying to marshal event-status: %s", err)
						return err
					}

					eventwriter.WriteData(string(bytes))
				}

				return err
			})
		}
	}(route)

	route.ServeHTTP(responsewriter, request)
}
