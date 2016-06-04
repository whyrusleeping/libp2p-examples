package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	host "gx/ipfs/QmXJBB9U6e6ennAJPzk8E2rSaVGuHVR2jCxE9H9gPDtRrq/go-libp2p/p2p/host"
	bhost "gx/ipfs/QmXJBB9U6e6ennAJPzk8E2rSaVGuHVR2jCxE9H9gPDtRrq/go-libp2p/p2p/host/basic"
	metrics "gx/ipfs/QmXJBB9U6e6ennAJPzk8E2rSaVGuHVR2jCxE9H9gPDtRrq/go-libp2p/p2p/metrics"
	net "gx/ipfs/QmXJBB9U6e6ennAJPzk8E2rSaVGuHVR2jCxE9H9gPDtRrq/go-libp2p/p2p/net"
	conn "gx/ipfs/QmXJBB9U6e6ennAJPzk8E2rSaVGuHVR2jCxE9H9gPDtRrq/go-libp2p/p2p/net/conn"
	swarm "gx/ipfs/QmXJBB9U6e6ennAJPzk8E2rSaVGuHVR2jCxE9H9gPDtRrq/go-libp2p/p2p/net/swarm"
	testutil "gx/ipfs/QmXJBB9U6e6ennAJPzk8E2rSaVGuHVR2jCxE9H9gPDtRrq/go-libp2p/testutil"
	pstore "gx/ipfs/QmZ62t46e9p7vMYqCmptwQC1RhRv5cpQ5cwoqYspedaXyq/go-libp2p-peerstore"

	ipfsaddr "github.com/ipfs/go-ipfs/thirdparty/ipfsaddr"
	ma "gx/ipfs/QmYzDkkgAEmrcNzFCiYo6L1dTX4EAG1gZkbtdbd9trL4vd/go-multiaddr"
	context "gx/ipfs/QmZy2y8t9zQH2a1b8q2ZSLKp17ATuJoCNxxyMFG5qFExpt/go-net/context"
)

var _ = io.Copy

// create a 'Host' with a random peer to listen on the given address
func makeDummyHost(listen string) (host.Host, error) {
	addr, err := ma.NewMultiaddr(listen)
	if err != nil {
		return nil, err
	}

	pid, err := testutil.RandPeerID()
	if err != nil {
		return nil, err
	}

	// bandwidth counter, should be optional in the future
	bwc := metrics.NewBandwidthCounter()

	// create a new swarm to be used by the service host
	netw, err := swarm.NewNetwork(context.Background(), []ma.Multiaddr{addr}, pid, pstore.NewPeerstore(), bwc)
	if err != nil {
		return nil, err
	}

	log.Printf("I am %s/ipfs/%s\n", addr, pid.Pretty())
	return bhost.New(netw), nil
}

func main() {
	conn.EncryptConnections = false
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	flag.Parse()

	listenaddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *listenF)

	ha, err := makeDummyHost(listenaddr)
	if err != nil {
		log.Fatal(err)
	}

	message := []byte("hello libp2p!")
	// Set a stream handler on host A
	ha.SetStreamHandler("/echo/1.0.0", func(s net.Stream) {
		defer s.Close()
		log.Println("writing message")
		s.Write(message)
	})

	if *target == "" {
		log.Println("listening on for connections...")
		for {
			time.Sleep(time.Hour)
		}
	}

	a, err := ipfsaddr.ParseString(*target)
	if err != nil {
		log.Fatalln(err)
	}

	pi := pstore.PeerInfo{
		ID:    a.ID(),
		Addrs: []ma.Multiaddr{a.Transport()},
	}

	log.Println("connecting to target")
	err = ha.Connect(context.Background(), pi)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("opening stream...")
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set
	s, err := ha.NewStream(context.Background(), "/echo/1.0.0", a.ID())
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("reading message")
	out, err := ioutil.ReadAll(s)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("GOT: ", string(out))
}
