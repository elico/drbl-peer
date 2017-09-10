package drblpeer

import (
	//	"../watcher"
	"fmt"
	"sync/atomic"

	"github.com/asaskevich/govalidator"
	//"github.com/bogdanovich/dns_resolver"
	"github.com/elico/dns_resolver"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

type YamlPeerDrblName struct {
	Peer     string   `name`
	Type     string   `type`
	Host     string   `host`
	Port     int      `port`
	Weight   int      `weight`
	Path     string   `path`
	Expected []string `expected`
}

type YamlDrblPeers struct {
	Clients []YamlPeerDrblName `peers`
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

func NewPeerListFromYamlFile(filename string, hitWeight int64, timeout int, debug bool) (*DrblPeers, error) {
	//peerName, protocol, path string, port int, weight int64, bladdr []string 	) *DrblClient
	// mainDrbPeer := drblpeer.New("199.85.126.20", "dns", "", 53, int64(128), []string{"156.154.175.216", "156.154.176.216"})
	// drblPeers := *drblpeer.DrblPeers{[]DrblClient{}, int64(128), 30, true}

	//var drblPeers drblpeer.DrblPeers{[]drblpeer.DrblClient{*mainDrbPeer}, int64(128), 30, true}

	// Read whole file
	// Split lines
	// Walk on the lines
	// If the line syntax is fine add it else move on
	// Swap the drbl peers list

	var peers YamlDrblPeers
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return &DrblPeers{}, err
	}
	err = yaml.Unmarshal(content, &peers)
	if err != nil {
		return &DrblPeers{}, err
	}
	if debug {
		fmt.Println(string(content))
		fmt.Println(peers.Clients)
	}
	peersClient := make([]DrblClient, 0)
	newlist := &DrblPeers{peersClient, hitWeight, timeout, debug}

	for _, client := range peers.Clients {
		peer, err := NewPeerFromYaml(client)
		if err != nil {
			if debug {
				fmt.Println(err)
			}
			continue
		} else {
			newlist.Peers = append(newlist.Peers, *peer)
		}
		if newlist.Debug {
			fmt.Println("Peers number", int64(len(newlist.Peers)))
		}
	}
	return newlist, nil
}

func NewPeerFromYaml(peer YamlPeerDrblName) (*DrblClient, error) {

	port := uint(peer.Port)     //port
	weight := uint(peer.Weight) //weight

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
	if !govalidator.IsHost(peer.Host) {
		return &DrblClient{}, fmt.Errorf("%s is not a valid hostname or ip address", peer.Host)
	}
	// http\dns\dnsrbl ip\domain /path/vote/ port weigth bl ip's
	// 0								1					2						3			4			5
	switch {
	case peer.Type == "http" || peer.Type == "https":
		return &DrblClient{peer.Host,
			peer.Path,
			int(port),
			int64(weight),
			peer.Type,
			[]string{},
			dns_resolver.New([]string{peer.Host}),
			&http.Client{},
		}, nil

	case peer.Type == "dns":
		blIpAddr := make([]string, 1)
		for _, addr := range peer.Expected {
			if govalidator.IsIPv4(addr) {
				blIpAddr = append(blIpAddr, addr)
			}
		}
		return &DrblClient{peer.Host,
			peer.Path,
			int(port),
			int64(weight),
			peer.Type,
			blIpAddr,
			dns_resolver.NewWithPort([]string{peer.Host}, strconv.Itoa(int(port))),
			&http.Client{},
		}, nil
	case peer.Type == "dnsrbl":
		blIpAddr := make([]string, 1)
		for _, addr := range peer.Expected {
			if govalidator.IsIPv4(addr) {
				blIpAddr = append(blIpAddr, addr)
			}
		}
		return &DrblClient{peer.Host,
			peer.Path,
			int(port),
			int64(weight),
			peer.Type,
			blIpAddr,
			dns_resolver.NewWithPort([]string{peer.Host}, strconv.Itoa(int(port))),
			&http.Client{},
		}, nil
	}
	return &DrblClient{}, fmt.Errorf("drblpeer: malformed peerYaml %s", peer)
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
			// http\dns\dnsrbl ip\domain /path/vote/ port weigth bl ip's
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
			case lparts[0] == "dnsrbl":
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
	if len(peersList.Peers) < 1 {
		localWeight = 100 * 1024
	}
	if peersList.Debug {
		fmt.Println("Peerlist size", peersList.Peers, "while checking:", hostname, "Local weigth:", localWeight)
	}
	if len(peersList.Peers) < 1 {
		return block, localWeight
	}

	for _, peer := range peersList.Peers {
		if localWeight <= int64(0) {
			block = true
			return block, localWeight
		}

		found, allowaccess, admin, key, err := peer.Check(hostname, peersList.Debug)
		if err != nil {
			if peersList.Debug {
				fmt.Println("peer", peer.Peername, "had an error", err, "while checking:", hostname, "Allow acces:", allowaccess)
			}
			continue
		}
		if peersList.Debug {
			fmt.Println("peer", peer.Peername, ", results: found =>", found, "allow-access =>", allowaccess, "admin =>", admin, "key =>", key, "hostname =>", hostname)
		}

		if found {
			atomic.AddInt64(&localWeight, -peer.Weight)
		}
		if found && !allowaccess {
			if peersList.Debug {
				fmt.Println("Peer", peer.Peername, "weigth =>", peer.Weight, "hostname =>,", hostname, "!allowaccess", "found =>", found)
			}
		} else {
			if peersList.Debug {
				fmt.Println("Peer", peer.Peername, "weigth =>", peer.Weight, "hostname =>,", hostname, "allowaccess", "found =>", found)
			}
		}
	}
	if localWeight <= int64(0) {
		block = true
	}
	return block, localWeight
}

func (peersList *DrblPeers) CheckUrlWithSrc(requestUrl, src string) (bool, int64) {
	block := false

	localWeight := peersList.HitWeight
	if len(peersList.Peers) < 1 {
		localWeight = 100 * 1024
	}
	if peersList.Debug {
		fmt.Println("Peerlist size", peersList.Peers, "while checking:", requestUrl, "For Source", src, "Local weigth:", localWeight)
	}
	if len(peersList.Peers) < 1 {
		return block, localWeight
	}

	for _, peer := range peersList.Peers {

		if localWeight <= int64(0) {
			block = true
			return block, localWeight
		}

		found, allowaccess, admin, key, err := peer.HttpCheckUrlWithSrc(requestUrl, src, peersList.Debug)
		if peer.Protocol == "http" || peer.Protocol == "https" {
			// OK+		if peersList.Debug {
			if peersList.Debug {
				fmt.Println("peer", peer.Peername, "peer-protocol", peer.Protocol, ", results: found =>", found, "allow-access =>", allowaccess, "admin =>", admin, "url =>", requestUrl, "src =>", src)
			}
		} else {
			continue
		}
		if err != nil {
			if peersList.Debug {
				fmt.Println("peer", peer.Peername, "had an error", err, "while checking:", requestUrl, "Allow acces:", allowaccess)
			}
			continue
		}
		if peersList.Debug {
			fmt.Println("peer", peer.Peername, ", results: found =>", found, "allow-access =>", allowaccess, "admin =>", admin, "key =>", key, "url =>", requestUrl, "src =>", src)
		}

		if found {
			atomic.AddInt64(&localWeight, -peer.Weight)
		}
		if found && !allowaccess {
			if peersList.Debug {
				fmt.Println("Peer", peer.Peername, "weigth =>", peer.Weight, "url =>,", requestUrl, "!allowaccess", "src =>", src, "found =>", found)
			}
		} else {
			if peersList.Debug {
				fmt.Println("Peer", peer.Peername, "weigth =>", peer.Weight, "url =>,", requestUrl, "allowaccess", "src =>", src, "found =>", found)
			}
		}
	}
	if localWeight <= int64(0) {
		block = true
	}
	return block, localWeight
}
