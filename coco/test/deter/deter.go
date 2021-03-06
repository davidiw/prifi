// deter is the deterlab process that should run on the boss node
//
// It spawns multiple timestampers and clients, while constructing
// the topology defined on cfg.json. It assumes that hosts.txt has
// the entire list of hosts to run timestampers on and that the final
// host is the designated logging server.
//
// The overall topology that is created is defined by cfg.json.
// The port layout for each node, however, is specified here.
// cfg.json will assign each node a port p. This is the port
// that each singing node is listening on. The timestamp server
// to which clients connect is listneing on port p+1. And the
// pprof server for each node is listening on port p+2. This
// means that in order to debug each client, you can forward
// the p+2 port of each node to your localhost.
//
// In the future the loggingserver will be connecting to the
// servers on the pprof port in order to gather extra data.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/dedis/prifi/coco/test/cliutils"
	"github.com/dedis/prifi/coco/test/config"
	"github.com/dedis/prifi/coco/test/graphs"
)

func main() {
	fmt.Println("running deter")
	// fs defines the list of files that are needed to run the timestampers.
	fs := []string{"exec", "timeclient", "cfg.json", "virt.txt", "phys.txt"}

	// read in the hosts file.
	virt, err := cliutils.ReadLines("virt.txt")
	if err != nil {
		log.Fatal(err)
	}
	phys, err := cliutils.ReadLines("phys.txt")
	if err != nil {
		log.Fatal(err)
	}
	vpmap := make(map[string]string)
	for i := range virt {
		vpmap[virt[i]] = phys[i]
	}
	// kill old processes
	var wg sync.WaitGroup
	for _, h := range phys {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			cliutils.SshRun("", h, "killall exec logserver timeclient scp ssh")
		}(h)
	}
	wg.Wait()
	logger := phys[len(phys)-1]
	phys = phys[:len(phys)-1]
	virt = virt[:len(virt)-1]

	// Read in and parse the configuration file
	file, e := ioutil.ReadFile("cfg.json")
	if e != nil {
		log.Fatal("Error Reading Configuration File: %v\n", e)
	}

	var tree graphs.Tree
	json.Unmarshal(file, &tree)

	hostnames := make([]string, 0, len(virt))
	tree.TraverseTree(func(t *graphs.Tree) {
		hostnames = append(hostnames, t.Name)
	})

	cf := config.ConfigFromTree(&tree, hostnames)
	cfb, err := json.Marshal(cf)
	if err != nil {
		log.Fatal(err)
	}

	// write out a true configuration file
	log.Println(string(cfb))
	err = ioutil.WriteFile("cfg.json", cfb, 0666)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("copying over files")
	// copy the files over to all the host machines.
	for _, f := range fs {
		for _, h := range phys {
			wg.Add(1)
			go func(h string, f string) {
				defer wg.Done()
				cliutils.Scp("", h, f, f)
			}(h, f)
		}
	}

	cliutils.Scp("", logger, "logserver", "")

	wg.Wait()

	// start up the logging server on the final host at port 10000
	fmt.Println("starting up logserver")
	loggerport := logger + ":10000"
	go cliutils.SshRunStdout("", logger, "cd logserver; ./logserver -addr="+loggerport)
	// wait a little bit for the logserver to start up
	time.Sleep(2 * time.Second)
	fmt.Println("starting time clients")
	for _, host := range hostnames {
		h, p, _ := net.SplitHostPort(host)
		pn, _ := strconv.Atoi(p)
		hp := net.JoinHostPort(h, strconv.Itoa(pn+1))
		go cliutils.SshRunStdout("", h, "./timeclient -rate=5000 -name=client@"+hp+" -server="+hp)
	}
	// now start up each timestamping server
	fmt.Println("starting up timestampers")
	tree.TraverseTree(func(t *graphs.Tree) {
		h, _, err := net.SplitHostPort(t.Name)
		if err != nil {
			log.Fatal("improperly formatted host. must be host:port")
		}
		// get the physical node this is associated with
		phys := vpmap[h]
		wg.Add(1)
		go func() {
			defer wg.Done()
			// run the timestampers
			log.Println("ssh timestamper at:", phys)
			log.Println("running timestamp at @", t.Name, "listening to logger:", loggerport)
			cliutils.SshRunStdout("", phys, "./exec -hostname="+t.Name+" -logger="+loggerport)
		}()
	})
	// wait for the servers to finish before stopping
	wg.Wait()
}
