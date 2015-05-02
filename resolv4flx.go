package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
)

var (
	workers       int
	fileName      string
	waitResolvers sync.WaitGroup
)

func init() {
	flag.IntVar(&workers, "workers", 5, "Number of worker threads to resolve DNS entries")

}

func main() {

	//check args
	argsCheck()

	// Launch DNS resolvers
	ch := make(chan string)
	for i := 0; i < workers; i++ {
		go resolveEntry(ch)
	}

	// Launch file reader
	waitResolvers.Add(1)
	readEntriesFile(fileName, ch)

	// Exit when when reader and resolvers are done
	waitResolvers.Wait()

	fmt.Println("Finished processing")
}

func usage() {
	// TODO print usage
	fmt.Println("Utility to resolve DNS entries in a file")
	os.Exit(2)

}

// Check command-line arguments
func argsCheck() {

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) != 1 {
		usage()
	}

	fileName = args[0]

}

func readEntriesFile(fileName string, ch chan string) {

	// decrement wait count on exit
	defer waitResolvers.Done()

	fhandler, err := os.Open(fileName)

	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(2)
	}

	defer fhandler.Close()

	scanner := bufio.NewScanner(fhandler)

	for scanner.Scan() {

		if scanner.Text() != "" {
			// add wait count for each file entry
			waitResolvers.Add(1)
			ch <- scanner.Text()
		}
	}

	close(ch)

}
