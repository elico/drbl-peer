package drblpeer

import (
	//	"../watcher"
	"fmt"

	"github.com/asaskevich/govalidator"
	//"github.com/bogdanovich/dns_resolver"
	"github.com/elico/dns_resolver"

	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	//"time"
)

var spacesAndTabs = regexp.MustCompile(`[\s\t]+`)

type DrblPeers struct {
	Peers     []DrblClient
	HitWeight int64
	Timeout   int
	Debug     bool
	/*
		ReverseOrderDomLookup bool
		ReverseOrderIpv4Lookup bool
		Ipv6Lookup bool
	*/
}

func NewPeerListFromFile(filename string, hitWeight int64, timeout int, debug bool) (*DrblPeers, error) {
	//peerName, protocol, path string, port int, weight int64, bladdr []string 	) *DrblClient
	// mainDrbPeer := drblpeer.New("199.85.126.20", "dns", "", 53, int64(128), []string{"156.154.175.216", "156.154.176.216"})
	// drblPeers := *drblpeer.DrblPeers{[]DrblClient{}, int64(128), 30, true}

	//var drblPeers drblpeer.DrblPeers{[]drblpeer.DrblClient{*mainDrbPeer}, int64(128), 30, true}

	// Read whole file
	// Split lines
	// Walk on the lines
	// If the line syntax is fine add it else move on
	// Swap the drbl peers list
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		//fmt.Println(err)
		return &DrblPeers{}, err

	}
	peersClient := make([]DrblClient, 0)
	newlist := &DrblPeers{peersClient, hitWeight, timeout, debug}

	lines := strings.Split(string(content), "\n")
	for _, peerline := range lines {
		if len(peerline) > 10 {
			peer, err := NewPeerFromLine(peerline)
			if err != nil {
				if debug {
					fmt.Println(err)
				}
				continue
			} else {
				newlist.Peers = append(newlist.Peers, *peer)
			}
		}
	}
	if newlist.Debug {
		fmt.Println("Peers number", int64(len(newlist.Peers)))
	}
	return newlist, nil
}

func NewPeerFromLine(peerline string) (*DrblClient, error) {
	// Cleanup
	peerline = spacesAndTabs.ReplaceAllString(peerline, " ")
	/*
		if strings.Contains(peerline,"\t") {
			return &DrblClient{}, fmt.Errorf("tabs are not allowed in peerline +> %s", peerline)
		}
	*/
	if len(peerline) > 10 {
		lparts := strings.Split(peerline, " ")

		if len(lparts) > 5 {
			port, err := strconv.ParseUint(lparts[3], 10, 64) //port
			if err != nil {
				return &DrblClient{}, err
			}
			weight, err := strconv.ParseUint(lparts[4], 10, 64) //weight
			if err != nil {
				return &DrblClient{}, err
			}
			/*
				switch {
				case govalidator.IsIP(lparts[1]):
					;;
				case govalidator.IsDNSName(lparts[1]):
					if lparts[0] != "http" {
						addr, err := net.LookupHost(lparts[1])
						if err != nil {
								return &DrblClient{}, fmt.Errorf("%s hostname cannot be resolved", lparts[1])
						}
						// Choosing a static IP address
						lparts[1] = addr[0]
					}
				default:
					return &DrblClient{}, fmt.Errorf("%s is not a valid hostname or ip address", lparts[1])
				}
			*/
			if !govalidator.IsHost(lparts[1]) {
				return &DrblClient{}, fmt.Errorf("%s is not a valid hostname or ip address", lparts[1])
			}
			// http\dns\dnsbl ip\domain /path/vote/ port weigth bl ip's
			// 0								1					2						3			4			5
			switch {
			case lparts[0] == "http" || lparts[0] == "https":
				return &DrblClient{lparts[1],
					lparts[2],
					int(port),
					int64(weight),
					lparts[0],
					[]string{},
					dns_resolver.New([]string{lparts[1]}),
					&http.Client{},
				}, nil

			case lparts[0] == "dns":
				blIpAddr := make([]string, 1)
				for _, addr := range lparts[5:] {
					if govalidator.IsIPv4(addr) {
						blIpAddr = append(blIpAddr, addr)
					}
				}
				return &DrblClient{lparts[1],
					lparts[2],
					int(port),
					int64(weight),
					lparts[0],
					blIpAddr,
					dns_resolver.NewWithPort([]string{lparts[1]}, strconv.Itoa(int(port))),
					&http.Client{},
				}, nil
			case lparts[0] == "dnsbl":
				blIpAddr := make([]string, 1)
				for _, addr := range lparts[5:] {
					if govalidator.IsIPv4(addr) {
						blIpAddr = append(blIpAddr, addr)
					}
				}
				return &DrblClient{lparts[1],
					lparts[2],
					int(port),
					int64(weight),
					lparts[0],
					blIpAddr,
					dns_resolver.NewWithPort([]string{lparts[1]}, strconv.Itoa(int(port))),
					&http.Client{},
				}, nil
			}
		}
		return &DrblClient{}, fmt.Errorf("drblpeer: malformed peerline %s", peerline)
	} else {
		return &DrblClient{}, fmt.Errorf("drblpeer: malformed peerline %s", peerline)
	}
}

/*
func PeerListWatcher(filename string, peersList *DrblPeers, debug bool) {
	doneChan := make(chan bool)
	newlist, err := NewPeerListFromFile(filename, peersList.HitWeight, peersList.Timeout, peersList.Debug)
	if err != nil {
		fmt.Println(err)
		//DENY all unfortunate case
	} else {
		fmt.Println("no error on first parsing config file =>", filename)
		peersList = newlist
	}
	for {
		go func(doneChan chan bool) {
			defer func() {
				doneChan <- true
			}()

			err := watcher.WatchFile(filename)
			if err != nil && debug {
				fmt.Println("Error Watching config file =>", filename, "Error =>", err)
			}
			if debug {
				fmt.Println("File", filename, "has been changed")
			}
		}(doneChan)

		select {
		case <-doneChan:
			newlist, err := NewPeerListFromFile(filename, peersList.HitWeight, peersList.Timeout, peersList.Debug)
			if err != nil {
				//DENY all unfortunate case
				fmt.Println("Error parsing config file =>", filename, "Error =>", err)
			} else {
				peersList = newlist
			}
		}
	}
}
*/

//Block and weight
func (peersList *DrblPeers) Check(hostname string) (bool, int64) {
	block := false

	localWeight := peersList.HitWeight

	for _, peer := range peersList.Peers {
		if localWeight <= int64(0) {
			block = true
			return block, localWeight
		}

		found, allowaccess, admin, key, err := peer.Check(hostname)
		if err != nil {
			if peersList.Debug {
				fmt.Println("peer", peer.Peername, "had an error", err, "while checking:", hostname, "Allow acces:", allowaccess)
			}
			continue
		}
		if peersList.Debug {
			fmt.Println("peer", peer.Peername, ", results =>", found, allowaccess, admin, key, hostname)
		}

		if found && !allowaccess {
			if peersList.Debug {
				fmt.Println("Peer", peer.Peername, "weigth =>", peer.Weight, "hostname =>,", hostname)
			}
			atomic.AddInt64(&localWeight, -peer.Weight)
		} else {
			if peersList.Debug {
				fmt.Println("Peer", peer.Peername, "weigth =>", peer.Weight, "hostname =>,", hostname)
			}
		}
	}
	if localWeight <= int64(0) {
		block = true
	}
	return block, localWeight
}
