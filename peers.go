package drblpeer

import (
	"github.com/asaskevich/govalidator"
	"github.com/bogdanovich/dns_resolver"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	//	"fmt"
)

type DrblClient struct {
	Peername    string
	Path        string
	Port        int
	Weight      int64
	Protocol    string
	BlResponses []string
	Resolver    *dns_resolver.DnsResolver
	Client      *http.Client
}

func New(peerName, protocol, path string, port int, weight int64, bladdr []string) *DrblClient {
	return &DrblClient{peerName,
		path,
		port,
		weight,
		protocol,
		bladdr,
		dns_resolver.New([]string{peerName}),
		&http.Client{},
	}
}

// In case of i/o timeout the resolver should be set to retry 3 times
/*
  instance.Resolver.RetryTimes = 0
*/

// return: found, allow\deny, admin\nonadmin, error)
func (instance *DrblClient) Check(hostname string) (bool, bool, bool, string, error) {
	found := false
	admin := false
	allow := true
	key := ""

	switch {
	case instance.Protocol == "http" || instance.Protocol == "https":
		testurl, _ := url.Parse(instance.Protocol + "://" + instance.Peername + ":" + strconv.Itoa(instance.Port) + instance.Path)
		testurlVals := url.Values{}
		testurlVals.Set("host", hostname)
		testurl.RawQuery = testurlVals.Encode()

		request, err := http.NewRequest("HEAD", testurl.String(), nil)
		//request.SetBasicAuth(*user, *pass)

		resp, err := instance.Client.Do(request)
		if err != nil {
			return found, allow, admin, key, err
		}

		if resp.Header.Get("X-Admin-Vote") == "YES" {
			admin = true
			found = true
		}
		if resp.Header.Get("X-Vote") == "BLOCK" {
			found = true
			allow = false
		}

		// For cases which debug is required you can use this key to see the BL VALUE
		//resp.Header.Get("X-Rate")

		key = resp.Header.Get("X-Rate-Key")

	case instance.Protocol == "dns":
		if len(hostname) > 1 {
			ip, err := instance.Resolver.LookupHost(hostname)
			if err != nil {
				/* to debug use this:
				//fmt.Println(instance, "Got error on lookup for", hostname)
				*/
				return found, false, admin, key, err
			}

			//	I could have added a loop over loop to verify that each host from the results
			//	is not from the blacklisting results but it's not that important

			if err == nil && len(ip) > 0 {
				for _, block := range instance.BlResponses {
					if ip[0].String() == block {
						found = true
						allow = false
						break
					}
				}
			}
		}

	case instance.Protocol == "dnsrbl":
		if govalidator.IsIPv4(hostname) {
			hostname = ReverseTheDomain(hostname)
		}
		if len(hostname) > 1 {
			ip, err := instance.Resolver.LookupHost(hostname + "." + instance.Peername)
			if err != nil {
				return found, false, admin, key, err
			}

			if err == nil && len(ip) > 0 {
				for _, block := range instance.BlResponses {
					if ip[0].String() == block {
						found = true
						allow = false
						break
					}
				}
			}
		}
	}
	return found, allow, admin, key, nil
}

// We need to implement a peer object which can be either DNS or HTTP,
// the url of the host can be:
// - host\ip:port
// - type: dns, dnsrbl, http

// Should have an interace\function: checkBl(host)(found, allowed, key)

func ReverseTheDomain(orig string) string {
	var c []string = strings.Split(orig, ".")

	for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
		c[i], c[j] = c[j], c[i]
	}

	return strings.Join(c, ".")
}
