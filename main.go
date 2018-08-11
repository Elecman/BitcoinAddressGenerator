package main

import "fmt"

func main() {
	start := G
	for i := 0; i < 10; i++ {
		start = start.MultiplyScalar()
		fmt.Println(start.EncodeUncompressedSec(), start.GetAddress(MAINNET))
	}

}