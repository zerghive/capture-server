package main

import (
	"appscope.net/conntrac"
	"fmt"
)

func main() {

	ep_chan := conntrac.NatConntrac()

	for {
		cr := <-ep_chan
		fmt.Println(cr.String())
	}
}
