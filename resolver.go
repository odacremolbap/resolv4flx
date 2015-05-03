package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
)

var outLock sync.Mutex

func resolveEntry(ch chan string) {

	var response string
	var err error

	for {
		entry := <-ch
		if entry == "" {
			// if nothing was sent through channel, loop again
			continue
		}

		domain, entryType := parseLine(entry)

		switch entryType {
		case "A", "AAAA":
			response, err = resolveA(domain)
		case "PTR":
			response, err = resolvePTR(domain)
		case "MX":
			response, err = resolveMX(domain)
		case "SRV":
			response, err = resolveSRV(domain)
		default:
			response = ""
			err = errors.New("Lookup for " + entryType + " not supported")
		}

		printResult(domain, entryType, response, err)

		waitResolvers.Done()
	}

}

func parseLine(entry string) (string, string) {

	var domain string
	entryFields := strings.Split(entry, "\t")

	if entryFields[0][len(entryFields[0])-1:] == "." {
		domain = entryFields[0][:len(entryFields[0])-1]
	} else {
		domain = entryFields[0]
	}

	return domain, entryFields[1]
}

func resolveA(domain string) (string, error) {

	ips, err := net.LookupIP(domain)

	if err != nil {
		return "", err
	}

	//TODO build string using buffers
	var header string
	response := ""

	for _, ip := range ips {

		if ip.To4() == nil {
			header = "IPv6:"
		} else {
			header = "IPv4:"
		}
		response += "[" + header + ip.String() + "]"
	}

	return response, nil
}

func resolveMX(domain string) (string, error) {

	mxs, err := net.LookupMX(domain)

	if err != nil {
		return "", err
	}

	//TODO build string using buffers
	response := ""

	for _, mx := range mxs {
		response += fmt.Sprintf("[host:%s,pref:%d]", mx.Host, mx.Pref)
	}

	return response, nil
}

func resolvePTR(domain string) (string, error) {

	//reverse ip octects and remove '.in-addr.arpa'
	ptrChunks := strings.Split(domain, ".")

	//allocate space
	bufIP := make([]byte, len(ptrChunks[0])+len(ptrChunks[1])+len(ptrChunks[2])+len(ptrChunks[3])+3)
	pointer := 0

	pointer = copy(bufIP, ptrChunks[3])
	pointer += copy(bufIP[pointer:], ".")
	pointer += copy(bufIP[pointer:], ptrChunks[2])
	pointer += copy(bufIP[pointer:], ".")
	pointer += copy(bufIP[pointer:], ptrChunks[1])
	pointer += copy(bufIP[pointer:], ".")
	copy(bufIP[pointer:], ptrChunks[0])

	domain = string(bufIP)

	names, err := net.LookupAddr(domain)

	if err != nil {
		return "", err
	}

	//TODO build string using buffers
	response := ""

	for _, name := range names {
		response += "[name:" + name + "]"
	}

	return response, nil

}

func resolveSRV(domain string) (string, error) {

	var serviceIndex int
	var protocolIndex int

	// extract service and protocol
	serviceIndex = strings.Index(domain, ".")

	if serviceIndex != -1 {
		protocolIndex = strings.Index(domain[serviceIndex+1:], ".") + serviceIndex
	}

	if serviceIndex == -1 || protocolIndex < serviceIndex {
		return "", errors.New("SRV query '" + domain + "' should be formated as 'service.protocol.domain'")
	}

	service := domain[:serviceIndex]
	protocol := domain[serviceIndex+1 : protocolIndex+1]

	// remove underscore

	if service[0:1] == "_" {
		service = service[1:]
	}

	if protocol[0:1] == "_" {
		protocol = protocol[1:]
	}

	cname, srvs, err := net.LookupSRV(service, protocol, domain[protocolIndex+2:])

	if err != nil {
		return "", err
	}

	//TODO build string using buffers
	response := "cname:" + cname

	for _, srv := range srvs {
		response += fmt.Sprintf("[target:%s,port:%d,priority:%d,weight:%d]", srv.Target, srv.Port, srv.Priority, srv.Weight)
	}

	return response, err
}

func printResult(query string, entryType string, response string, err error) {

	outLock.Lock()
	defer outLock.Unlock()

	fmt.Println("----------------------------------")
	fmt.Println("Query: " + query)
	fmt.Println("Type: " + entryType)

	if err != nil {
		fmt.Println("Error: " + err.Error())
	} else {
		fmt.Println("Response: " + response)
	}

}
