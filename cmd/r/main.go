package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/appscode/go/net"
)

func main() {
	iface, ip, err := net.NodeIP()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("ip:", ip.String())
	fmt.Println("iface:", iface)

	b, err := ioutil.ReadFile("/home/tamal/Desktop/sp.txt")
	if err != nil {
		log.Fatalln(err)
	}
	sp := string(b)
	lines := strings.Split(sp, "\n")
	for _, line := range lines {
		// fmt.Println(line)
		if strings.HasPrefix(line, "Brick ") {
			l := line[len("Brick "):]
			d := strings.Split(l, ":")
			fmt.Println("Brick:", d[0])
			fmt.Println("Location:", d[1])
		}
		if strings.HasPrefix(line, "<gfid:") {
			l := line[len("<gfid:") : len(line)-1]
			fmt.Println(l)
		}
	}
}
