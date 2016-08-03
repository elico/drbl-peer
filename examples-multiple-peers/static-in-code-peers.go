package main

import (
	"github.com/elico/drbl-peer"
	"bufio"
	"fmt"
	"os"
)

func main() {
	peerOne := drblpeer.New("199.85.126.20", "dns", "/", 53, int64(128), []string{"156.154.175.216", "156.154.176.216"})
	peerTwo := drblpeer.New("199.85.127.20", "dns", "/", 53, int64(128), []string{"156.154.175.216", "156.154.176.216"})

	drblPeers := drblpeer.DrblPeers{[]drblpeer.DrblClient{*peerOne, *peerTwo}, int64(128), 30, false}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter domain name: ")
		text, _ := reader.ReadString('\n')
		// The len - 1 is since there is a "new line" character on every new line
		res, weight := drblPeers.Check(text[:len(text)-1])

		fmt.Println("Result =>", res, weight)
	}

}
