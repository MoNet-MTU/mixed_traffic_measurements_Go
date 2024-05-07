
package main

import (
	"flag"
	"fmt"
	"os/exec"
	"time"
)

func main() {

	// Capture packets for a specified length of time
	// on a specified interface, after a specified delay,
	// with tcpdump (by default) or tshark
	//
	// Example usage:
	//   go build start-capture.go
	//   ./start-capture -d=10 -i ens33 -t=200 -tshark

	// Collect arguments from the command line
	ptrDuration := flag.String("t", "10", "capture duration in seconds")
	ptrInterface := flag.String("i", "ens33", "capture interface")
	ptrDelay := flag.String("d", "0", "delay in seconds before starting capture")
	ptrTshark := flag.Bool("tshark", false, "use tshark instead of tcpdump")
	ptrBPFilter := flag.String("bpf", "port 6653 or 6633", "capture filter in BPF format")
	flag.Parse()

	// Prepare the command string
	currTime := time.Now().Format(time.RFC3339)
	cmdString := ""
	if *ptrTshark {
		// Note tshark output defaults to pcapng format
		// If capture files aren't created, maybe the current user is not a
		// member of the wireshark group (sudo usermod -a -G wireshark $USER)
		// or doesn't have rights to execute dumpcap (sudo chmod +x /usr/bin/dumpcap)
		cmdString = fmt.Sprintf("sleep %s; tshark -a duration:%s -i %s -w ./capture-%s.pcapng %s",
			*ptrDelay, *ptrDuration, *ptrInterface, currTime, *ptrBPFilter)
	} else {
		cmdString = fmt.Sprintf("sleep %s; sudo timeout --preserve-status %s tcpdump -i %s -w ./capture-%s.pcap %s",
			*ptrDelay, *ptrDuration, *ptrInterface, currTime, *ptrBPFilter)
	}
	//fmt.Println(cmdString)

	// Run the command
	cmd := exec.Command("sh", "-c", cmdString)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

