// TODO http api

package main

import (
	"fmt"
	"os"
)

func startApiServer(host string, port int) {
	fmt.Printf("Listening on %v:%v", host, port)
	os.Exit(0)
}
