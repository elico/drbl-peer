package main

import (
	"github.com/elico/drbl-peer"
	"bufio"
	"fmt"
	"os"
)

func main() {
	//        blockresults := []string{"156.154.175.216", "156.154.176.216"}
	//        resolver := dns_resolver.New([]string{"199.85.126.20", "199.85.127.20"})

	peer := drblpeer.New("199.85.126.20", "dns", "/", 53, int64(128), []string{"156.154.175.216", "156.154.176.216"})
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter domain name: ")
		text, _ := reader.ReadString('\n')
		res, _, _, _, _ := peer.Check(text[:len(text)-1])
		fmt.Println("Result =>", res, peer.Weight)
	}

}
