package verboten

import (
	"io"
	"net/http"
	"net/url"

	"github.com/reiver/go-erorr"
	"github.com/reiver/go-errhttp"

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

	var from string = query.Get("from")
	if "" == network {
		errhttp.ErrHTTPBadRequest.ServeHTTP(responsewriter, request)
		return
	}

	switch network {
	case "mastodon":
		serveMastodon(responsewriter, from)
	default:
		errhttp.ErrHTTPNotFound.ServeHTTP(responsewriter, request)
		return
	}
}

func serveMastodon(responsewriter http.ResponseWriter, from string) {

	const network string = "mastodon"

	io.WriteString(responsewriter, "network = '"+network+"'\n")
	io.WriteString(responsewriter, "from = '"+from+"'\n")

}

