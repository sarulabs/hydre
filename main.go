package main

import "os"

func main() {
	p := Phoenix{}
	p.Serve(os.Args)
}
