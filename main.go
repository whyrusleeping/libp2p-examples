package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gx/ipfs/QmVL44QeoQDTYK8RVdpkyja7uYcK3WDNoBNHVLonf9YDtm/go-libp2p/p2p/metrics"
	"gx/ipfs/QmVL44QeoQDTYK8RVdpkyja7uYcK3WDNoBNHVLonf9YDtm/go-libp2p/p2p/net"
	"gx/ipfs/QmVL44QeoQDTYK8RVdpkyja7uYcK3WDNoBNHVLonf9YDtm/go-libp2p/p2p/net/swarm"
	ma "gx/ipfs/QmYzDkkgAEmrcNzFCiYo6L1dTX4EAG1gZkbtdbd9trL4vd/go-multiaddr"
	"gx/ipfs/QmZy2y8t9zQH2a1b8q2ZSLKp17ATuJoCNxxyMFG5qFExpt/go-net/context"
	"gx/ipfs/QmbyvM8zRFDkbFdYyt1MnevUMJ62SiSGbfDFZ3Z8nkrzr4/go-libp2p-peer"
)

func Fatal(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

func dialAndSend(s *swarm.Swarm, target peer.ID) {
	str, err := s.NewStreamWithPeer(context.Background(), target)
	if err != nil {
		Fatal(err)
	}

	fmt.Fprintln(str, "Hello World!")
	str.Close()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("to run a listener, specify peer id and listen port")
		fmt.Println("to run a dialer, specify our id and port, and the target id and port")
		Fatal("must specify at least three args")
	}

	// any valid multihash works if we have the secio disabled
	id, err := peer.IDB58Decode(os.Args[1])
	if err != nil {
		Fatal(err)
	}

	addr, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/" + os.Args[2])
	if err != nil {
		Fatal(err)
	}

	var dialPeer peer.ID
	var dialAddr ma.Multiaddr
	if len(os.Args) >= 5 {
		p, err := peer.IDB58Decode(os.Args[3])
		if err != nil {
			Fatal(err)
		}

		a, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/" + os.Args[4])
		if err != nil {
			Fatal(err)
		}

		dialPeer = p
		dialAddr = a
	}

	// new empty peerstore
	pstore := peer.NewPeerstore()
	ctx := context.Background()

	// construct ourselves a swarmy thingy
	s, err := swarm.NewSwarm(ctx, []ma.Multiaddr{addr}, id, pstore, metrics.NewBandwidthCounter())
	if err != nil {
		Fatal(err)
	}

	// if we are the dialer, do a dial!
	if dialAddr != nil {

		// add the targets address to the peerstore
		pstore.AddAddr(dialPeer, dialAddr, peer.PermanentAddrTTL)

		dialAndSend(s, dialPeer)
		return
	}

	// set a function to handle streams
	s.SetStreamHandler(func(st net.Stream) {
		out, err := ioutil.ReadAll(st)
		if err != nil {
			Fatal(err)
		}

		fmt.Println(string(out))
	})

	// just wait around
	time.Sleep(time.Hour)
}
