package main

import (
	"github.com/elico/drbl-peer"
	"fmt"
	"github.com/chzyer/readline"
	"io"
)

func main() {
	peerOne := drblpeer.New("199.85.126.20", "dns", "/", 53, int64(128), []string{"156.154.175.216", "156.154.176.216"})
	peerTwo := drblpeer.New("199.85.127.20", "dns", "/", 53, int64(128), []string{"156.154.175.216", "156.154.176.216"})

	drblPeers := drblpeer.DrblPeers{[]drblpeer.DrblClient{*peerOne, *peerTwo}, int64(128), 30, false}

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
