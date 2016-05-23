package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	inFile  = "generate/agent.txt"
	outFile = "agent.go"
)

func main() {
	in, err := ioutil.ReadFile(inFile)
	if err != nil {
		log.Fatalf("Error while reading %v: %v\n", inFile, err)
	}
	agent := strings.TrimSpace(string(in))

	out, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Could not create %v: %v\n", outFile, err)
	}
	defer out.Close()
	fmt.Fprintf(out, "package main\n\n")
	fmt.Fprintf(out, "const malAgent = \"%v\"\n", agent)
}
