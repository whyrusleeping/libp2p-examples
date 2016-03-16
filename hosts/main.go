package main

import (
	"io/ioutil"
	"log"

	host "github.com/ipfs/go-libp2p/p2p/host"
	bhost "github.com/ipfs/go-libp2p/p2p/host/basic"
	metrics "github.com/ipfs/go-libp2p/p2p/metrics"
	net "github.com/ipfs/go-libp2p/p2p/net"
	swarm "github.com/ipfs/go-libp2p/p2p/net/swarm"
	peer "github.com/ipfs/go-libp2p/p2p/peer"
	testutil "github.com/ipfs/go-libp2p/testutil"

	context "golang.org/x/net/context"
	ma "gx/ipfs/QmcobAGsCjYt5DXoq9et9L8yR8er7o7Cu3DTvpaq12jYSz/go-multiaddr"
)

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
	netw, err := swarm.NewNetwork(context.Background(), []ma.Multiaddr{addr}, pid, peer.NewPeerstore(), bwc)
	if err != nil {
		return nil, err
	}

	return bhost.New(netw), nil
}

func main() {
	addrA := "/ip4/127.0.0.1/tcp/5550"
	addrB := "/ip4/127.0.0.1/tcp/5551"

	ha, err := makeDummyHost(addrA)
	if err != nil {
		log.Fatal(err)
	}

	hb, err := makeDummyHost(addrB)
	if err != nil {
		log.Fatal(err)
	}

	// Set a stream handler on host A
	ha.SetStreamHandler("/example", func(s net.Stream) {
		log.Println("GOT A CONNECTION!")
		s.Write([]byte("Hello World!"))
		s.Close()
	})

	pi := peer.PeerInfo{
		ID:    ha.ID(),
		Addrs: ha.Addrs(),
	}

	// connect host B to host A
	err = hb.Connect(context.Background(), pi)
	if err != nil {
		log.Fatalln(err)
	}

	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set
	s, err := hb.NewStream(context.Background(), "/example", ha.ID())
	if err != nil {
		log.Fatalln(err)
	}

	out, err := ioutil.ReadAll(s)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("GOT: ", string(out))
}
