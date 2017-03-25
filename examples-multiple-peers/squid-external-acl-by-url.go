package main

import (
	//"github.com/elico/drbl-peer"
	"../"
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var blockWeight int
var timeout int
var peersFileName string
var debug bool
var yamlconfig bool

var drblPeers *drblpeer.DrblPeers

func process_request(line string) {
	answer := "ERR"
	lparts := strings.Split(strings.TrimRight(line, "\n"), " ")
	if len(lparts[0]) > 1 {
		if debug {
			fmt.Fprintln(os.Stderr, "ERRlog: Proccessing request => \""+strings.TrimRight(line, "\n")+"\"")
		}
	}
	block, weight := drblPeers.CheckUrlWithSrc(lparts[1], lparts[2])
	if block {
		answer = "OK"
	}
	fmt.Println(lparts[0] + " " + answer + " weight=" + strconv.FormatInt(weight, 10))
}

func main() {
	flag.IntVar(&blockWeight, "block-weight", 128, "Peers blacklist weight")
	flag.IntVar(&timeout, "query-timeout", 30, "Timeout for all peers response")
	flag.BoolVar(&debug, "debug", false, "Run in debug mode")
	flag.BoolVar(&yamlconfig, "yamlconfig", false, "Use a yaml formated blacklist file")
	flag.StringVar(&peersFileName, "peers-filename", "peersfile.txt", "Blacklists peers filename")

	flag.Parse()

	if yamlconfig {
		drblPeers, _ = drblpeer.NewPeerListFromYamlFile(peersFileName, int64(blockWeight), timeout, debug)
	} else {
		drblPeers, _ = drblpeer.NewPeerListFromFile(peersFileName, int64(blockWeight), timeout, debug)
	}

	if debug {
		fmt.Println("Peers", drblPeers)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			// You may check here if err == io.EOF
			break
		}
		if strings.HasPrefix(line, "q") || strings.HasPrefix(line, "Q") {
			fmt.Fprintln(os.Stderr, "ERRlog: Exiting cleanly")
			break
		}

		go process_request(line)
	}
}
