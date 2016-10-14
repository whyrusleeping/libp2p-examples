package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/net/context"

	util "github.com/ipfs/go-ipfs-util"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"

	metrics "github.com/libp2p/go-libp2p-metrics"
	inet "github.com/libp2p/go-libp2p-net"
	swarm "github.com/libp2p/go-libp2p-swarm"
	ma "github.com/multiformats/go-multiaddr"
)

func Fatal(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

func randPeerID() peer.ID {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(util.NewTimeSeededRand(), buf); err != nil {
		return ""
	}
	h := util.Hash(buf)
	return peer.ID(h)
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
		randId := randPeerID()
		fmt.Println("example peer id: ", randId.Pretty())
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
	pstore := peerstore.NewPeerstore()
	ctx := context.Background()

	// construct ourselves a swarmy thingy
	s, err := swarm.NewSwarm(ctx, []ma.Multiaddr{addr}, id, pstore, metrics.NewBandwidthCounter())
	if err != nil {
		Fatal(err)
	}

	// if we are the dialer, do a dial!
	if dialAddr != nil {

		// add the targets address to the peerstore
		pstore.AddAddr(dialPeer, dialAddr, peerstore.PermanentAddrTTL)

		dialAndSend(s, dialPeer)
		return
	}

	// set a function to handle streams
	s.SetStreamHandler(func(st inet.Stream) {
		out, err := ioutil.ReadAll(st)
		if err != nil {
			Fatal(err)
		}

		fmt.Println(string(out))
	})

	// just wait around
	time.Sleep(time.Hour)
}
