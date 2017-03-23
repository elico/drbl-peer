package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/elico/drbl-peer"
	"os"
	"strconv"
	"strings"
)

var blockWeight int
var timeout int
var peersFileName string
var debug bool

var drblPeers *drblpeer.DrblPeers

func process_request(line string) {
	answer := "ERR"
	lparts := strings.Split(strings.TrimRight(line, "\n"), " ")
	if len(lparts[0]) > 0 {
		if debug {
			fmt.Fprintln(os.Stderr, "ERRlog: Proccessing request => \""+strings.TrimRight(line, "\n")+"\"")
		}
	}
	block, weight := drblPeers.Check(lparts[1])
	if block {
		answer = "OK"
	}
	fmt.Println(lparts[0] + " " + answer + " weight=" + strconv.FormatInt(weight, 10))
}

func main() {
	flag.IntVar(&blockWeight, "block-weight", 128, "Peers blacklist weight")
	flag.IntVar(&timeout, "query-timeout", 30, "Timeout for all peers response")
	flag.BoolVar(&debug, "debug", false, "Run in debug mode")
	flag.StringVar(&peersFileName, "peers-filename", "peersfile.yaml", "Blacklists peers filename")

	flag.Parse()

	drblPeers, _ = drblpeer.NewPeerListFromYamlFile(peersFileName, int64(blockWeight), timeout, debug)

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
