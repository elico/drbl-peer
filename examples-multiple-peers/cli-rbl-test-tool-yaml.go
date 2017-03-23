package main

import (
	"flag"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/elico/drbl-peer"
	"io"
)

var blockWeight int
var timeout int
var peersFileName string
var debug bool

func main() {
	flag.IntVar(&blockWeight, "block-weight", 128, "Peers blacklist weight")
	flag.IntVar(&timeout, "query-timeout", 30, "Timeout for all peers response")
	flag.BoolVar(&debug, "debug", false, "Run in debug mode")
	flag.StringVar(&peersFileName, "peers-filename", "peersfile.yaml", "Blacklists peers YAML filename")

	flag.Parse()

	drblPeers, _ := drblpeer.NewPeerListFromYamlFile(peersFileName, int64(blockWeight), timeout, debug)
	if debug {
		fmt.Println("Peers", drblPeers)
	}

	l, err := readline.NewEx(&readline.Config{
		//		HistoryFile:     "/tmp/readline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		l.SetPrompt("Enter domain name: ")
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		res, weight := drblPeers.Check(line)
		fmt.Println("Result =>", res, weight)
	}

}
