/*
An example of how to use go-icap.

Run this program and Squid on the same machine.
Put the following lines in squid.conf:

icap_enable on
icap_service service_req reqmod_precache icap://127.0.0.1:1344/hostname-logger/
adaptation_access service_req allow all

(The ICAP server needs to be started before Squid is.)

Set your browser to use the Squid proxy.

*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/asaskevich/govalidator"

	drblpeer "github.com/elico/drbl-peer"
	"github.com/elico/icap"
)

var (
	isTag             = "HOSTNAME-LOGGER"
	debug             int
	address           string
	maxConnections    string
	yamlconfig        bool
	peersFileName     string
	remoteLoggingMode bool
	blockWeight       int
	timeout           int

	// redisAddress   *string
	fullOverride = false
	// upnkeyTimeout  *int
	// useGoCache     *bool
	err error
	// letters        = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

var drblPeers *drblpeer.DrblPeers

// GlobalHTTPClient --dasdasddas
var GlobalHTTPClient *http.Client

func setCookie(req *icap.Request) bool {
	if debug > 0 {
		fmt.Println("Checking Request for \"Set-Cookie\"")
	}
	if _, setCookieExists := req.Request.Header["Set-Cookie"]; setCookieExists {
		if debug > 0 {
			fmt.Println("Set-Cookie Exists: ")
			fmt.Println(req.Request.Header["Set-Cookie"][0])
		}
		return true
	}
	if debug > 0 {
		fmt.Println("Set-Cookie Doesn't Exists!")
	}
	return false
}

func noCache(req *icap.Request) bool {
	if debug > 0 {
		fmt.Println("Checking Request or response for \"no-cache\" => ")
	}

	if _, cacheControlExists := req.Request.Header["Cache-Control"]; cacheControlExists {
		if debug > 0 {
			fmt.Println("Cache-Control Exists in the request: ")
			fmt.Println(req.Request.Header["Cache-Control"][0])
		}
		if strings.Contains(req.Request.Header["Cache-Control"][0], "no-cache") {
			if debug > 0 {
				fmt.Println("\"no-cache\" Exists in the request!")
			}
			return true
		}
	}
	if _, cacheControlExists := req.Response.Header["Cache-Control"]; cacheControlExists {
		if debug > 0 {
			fmt.Println("Cache-Control Exists in the response: ")
			fmt.Println("Cache-Control Header =>", reflect.TypeOf(req.Response.Header["Cache-Control"]))

			fmt.Println("Reflect tpyeof Cache-Control Header =>", reflect.TypeOf(req.Response.Header["Cache-Control"]))
			fmt.Println("len Cache-Control Header =>", len(req.Response.Header["Cache-Control"]))
			fmt.Println(req.Response.Header["Cache-Control"])
		}
		if len(req.Response.Header["Cache-Control"]) > 0 && strings.Contains(strings.Join(req.Response.Header["Cache-Control"], ", "), "no-cache") {
			fmt.Println("\"no-cache\" Exists in the response!")
			return true
		}
	}

	if debug > 0 {
		fmt.Println("Cache-Control headers Doesn't Exists in this requset and response")
	}
	return false
}

func wrongMethod(req *icap.Request) bool {
	if debug > 0 {
		fmt.Println("Checking Request method => ", req.Request.Method, req.Request.URL.String())
	}

	if req.Request.Method == "GET" {
		return false
	}
	return true

}

func hostnameLogger(w icap.ResponseWriter, req *icap.Request) {
	localDebug := false
	if strings.Contains(req.URL.RawQuery, "debug=1") {
		localDebug = true
	}

	h := w.Header()
	h.Set("isTag", isTag)
	h.Set("Service", "ICAP hostname logger serivce")

	if debug > 0 {
		fmt.Fprintln(os.Stderr, "Printing the full ICAP request")
		fmt.Fprintln(os.Stderr, req)
		fmt.Fprintln(os.Stderr, req.Request)
		fmt.Fprintln(os.Stderr, req.Response)
	}
	switch req.Method {
	case "OPTIONS":
		h.Set("Methods", "REQMOD, RESPMOD")
		h.Set("Options-TTL", "1800")
		h.Set("Allow", "204, 206")
		h.Set("Preview", "0")
		h.Set("Transfer-Preview", "*")
		h.Set("Max-Connections", maxConnections)
		h.Set("X-Include", "X-Client-IP, X-Authenticated-Groups, X-Authenticated-User, X-Subscriber-Id, X-Server-Ip, X-Store-Id")
		w.WriteHeader(200, nil, false)
	case "REQMOD":
		modified := false
		allow204 := false

		if _, allow204Exists := req.Header["Allow"]; allow204Exists {
			if strings.Contains(req.Header["Allow"][0], "204") {
				allow204 = true
			}
		}

		if debug > 0 || localDebug {
			for k, v := range req.Header {
				fmt.Fprintln(os.Stderr, "The ICAP headers:")
				fmt.Fprintln(os.Stderr, "key size:", len(req.Header[k]))
				fmt.Fprintln(os.Stderr, "key:", k, "value:", v)
			}
		}

		if debug > 0 || localDebug {
			for k, v := range req.Request.Header {
				fmt.Fprintln(os.Stderr, "key:", k, "value:", v)
			}
		}

		if debug > 0 {
			fmt.Println("Request hostname:", req.Request.URL.Hostname(), "is not a googlevideo.com match")
			fmt.Println("Request query UPN value:", req.Request.URL.Query().Get("upn"))
		}

		if debug > 0 {
			fmt.Println("204 allowed?", allow204)
			fmt.Println("modified?", modified)
			fmt.Println("end of the line 204 response!.. Shouldn't happen.")
		}
		var testHostname string
		switch req.Request.Method {
		case "CONNECT":
			doubleDotsCounter := strings.Count(req.Request.RequestURI, ":")
			bracketsCounter := strings.Count(req.Request.RequestURI, "[")
			if bracketsCounter < 1 && doubleDotsCounter == 1 {
				parts := strings.Split(req.Request.RequestURI, ":")
				if govalidator.IsIPv4(parts[0]) {
					// Nothing to log or check, it's an IPV4 hostname
					w.WriteHeader(204, nil, false)
					return
				}
				if govalidator.IsDNSName(parts[0]) {
					// set the testhostname
					testHostname = parts[0]
				}
			}
			if bracketsCounter > 0 {
				// Which means that the request host is an IPV6 hostname and we d not log it by default
				w.WriteHeader(204, nil, false)
				return
			}

		default:
			if govalidator.IsIPv4(req.Request.URL.Hostname()) {
				// Nothing to log or check, it's an IPV4 hostname
				w.WriteHeader(204, nil, false)
				return
			}
			if govalidator.IsIPv6(req.Request.URL.Hostname()) {
				// Nothing to log or check, it's an IPV6 hostname
				w.WriteHeader(204, nil, false)
				return
			}
			if govalidator.IsDNSName(req.Request.URL.Hostname()) {
				// set the testhostname
				testHostname = req.Request.URL.Hostname()
			}

		}

		go func() {
			_, _ = drblPeers.Check(testHostname)
		}()

		w.WriteHeader(204, nil, false)
		return
	case "RESPMOD":
		w.WriteHeader(204, nil, false)
		return
	default:
		w.WriteHeader(405, nil, false)
		if debug > 0 || localDebug {
			fmt.Fprintln(os.Stderr, "Invalid request method")
		}
	}
}

func defaultIcap(w icap.ResponseWriter, req *icap.Request) {
	localDebug := false
	if strings.Contains(req.URL.RawQuery, "debug=1") {
		localDebug = true
	}

	h := w.Header()
	h.Set("isTag", isTag)
	h.Set("Service", "Default ICAP serivce")

	if debug > 0 || localDebug {
		fmt.Fprintln(os.Stderr, "Printing the full ICAP request")
		fmt.Fprintln(os.Stderr, req)
		fmt.Fprintln(os.Stderr, req.Request)
	}
	switch req.Method {
	case "OPTIONS":
		h.Set("Methods", "REQMOD, RESPMOD")
		h.Set("Options-TTL", "1800")
		h.Set("Allow", "204")
		h.Set("Preview", "0")
		h.Set("Transfer-Preview", "*")
		h.Set("Max-Connections", maxConnections)
		h.Set("This-Server", "Default ICAP url which bypass all requests adaptation")
		h.Set("X-Include", "X-Client-IP, X-Authenticated-Groups, X-Authenticated-User, X-Subscriber-Id, X-Server-Ip, X-Store-Id")
		w.WriteHeader(200, nil, false)
	case "REQMOD":
		if debug > 0 || localDebug {
			fmt.Fprintln(os.Stderr, "Default REQMOD, you should use the apropriate ICAP service URL")
		}
		w.WriteHeader(204, nil, false)
	case "RESPMOD":
		if debug > 0 || localDebug {
			fmt.Fprintln(os.Stderr, "Default RESPMOD, you should use the apropriate ICAP service URL")
		}
		w.WriteHeader(204, nil, false)
	default:
		w.WriteHeader(405, nil, false)
		if debug > 0 || localDebug {
			fmt.Fprintln(os.Stderr, "Invalid request method")
		}
	}
}

func init() {
	fmt.Fprintln(os.Stderr, "Starting Hostname Logger ICAP serivce")

	// debug = flag.Bool("d", false, "Debug mode can be \"1\" or \"0\" for no")
	// maxConnections = flag.String("maxcon", "4000", "Maximum number of connections for the ICAP service")
	// redisAddress = flag.String("redis-address", "127.0.0.1:6379", "Redis DB address to store youtube tokens")
	// upnkeyTimeout = flag.Int("cache-key-timeout", 360, "Redis or GoCache DB key timeout in Minutes")
	// useGoCache = flag.Bool("Use GoCache", true, "GoCache DB is used by default and if disabled then Redis is used")

	// flag.Parse()

	flag.IntVar(&blockWeight, "block-weight", 4096, "Peers blacklist weight")
	flag.IntVar(&timeout, "query-timeout", 30, "Timeout for all peers response")
	flag.IntVar(&debug, "debug", 0, "Verbosity of the debug mode 0 is nonoe")

	flag.BoolVar(&yamlconfig, "yamlconfig", false, "Use a yaml formated blacklist file instead of a text based")
	flag.BoolVar(&remoteLoggingMode, "logger-mode", true, "Run in logging mode")

	flag.StringVar(&maxConnections, "maxcon", "4000", "Maximum number of connections for the ICAP service")

	flag.StringVar(&peersFileName, "peers-filename", "peersfile.txt", "Blacklists peers filename")
	flag.StringVar(&address, "listen", "127.0.0.1:1344", "Listening address for the ICAP service")

	flag.Parse()

	if yamlconfig {
		drblPeers, _ = drblpeer.NewPeerListFromYamlFile(peersFileName, int64(blockWeight), timeout, debug > 0)
	} else {
		drblPeers, _ = drblpeer.NewPeerListFromFile(peersFileName, int64(blockWeight), timeout, debug > 0)
	}

	if debug > 0 {
		fmt.Println("Peers", drblPeers)
	}
}

func main() {
	fmt.Fprintln(os.Stderr, "running requests Hostname Logger ICAP serivce :D")

	if debug > 0 {
		fmt.Fprintln(os.Stderr, "Config Variables:")
		fmt.Fprintln(os.Stderr, "Debug Level: => ", debug)
		fmt.Fprintln(os.Stderr, "Listen Address: => "+address)
		fmt.Fprintln(os.Stderr, "Maximum number of Connections: => "+maxConnections)
	}

	GlobalHTTPClient = &http.Client{}
	GlobalHTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("redirect")
	}

	icap.HandleFunc("/hostname-logger/", hostnameLogger)
	icap.HandleFunc("/", defaultIcap)
	icap.ListenAndServe(address, nil)
}
