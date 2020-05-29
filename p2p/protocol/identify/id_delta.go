package identify

import (
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/helpers"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	pb "github.com/libp2p/go-libp2p/p2p/protocol/identify/pb"

	ggio "github.com/gogo/protobuf/io"
)

const deltaMessageSize = 2 * 1024 // 2K

// IDDelta_1_1_0 is the current Delta protocol that exchanges standalone Delta messages.
const IDDelta_1_1_0 = "/p2p/id/delta/1.1.0"

// IDDelta_1_0_0 is the legacy Delta protocol that exchanges Delta messages wrapped in an
//  ID_1_0_0 message.
const IDDelta_1_0_0 = "/p2p/id/delta/1.0.0"

// deltaHandler handles incoming delta updates from peers.
func (ids *IDService) deltaHandler(s network.Stream) {
	var delta *pb.Delta
	r := ggio.NewDelimitedReader(s, deltaMessageSize)
	switch s.Protocol() {
	case IDDelta_1_1_0:
		mes := &pb.Delta{}
		if err := r.ReadMsg(mes); err != nil {
			log.Warning("error reading delta message: ", err)
			s.Reset()
			return
		}
		delta = mes
	case IDDelta_1_0_0:
		mes := pb.Identify_1_0_0{}
		if err := r.ReadMsg(&mes); err != nil {
			log.Warning("error reading identify delta message: ", err)
			s.Reset()
			return
		}
		delta = mes.GetDelta()
	default:
		log.Warnw("peer does not support delta protocol", "protocol", s.Protocol())
		s.Reset()
		return
	}

	c := s.Conn()
	defer helpers.FullClose(s)

	log.Debugf("%s received message from %s %s", s.Protocol(), c.RemotePeer(), c.RemoteMultiaddr())
	if delta == nil {
		return
	}

	p := s.Conn().RemotePeer()
	if err := ids.consumeDelta(p, delta); err != nil {
		log.Warningf("delta update from peer %s failed: %s", p, err)
	}
}

// consumeDelta processes an incoming delta from a peer, updating the peerstore
// and emitting the appropriate events.
func (ids *IDService) consumeDelta(id peer.ID, delta *pb.Delta) error {
	err := ids.Host.Peerstore().AddProtocols(id, delta.GetAddedProtocols()...)
	if err != nil {
		return err
	}

	err = ids.Host.Peerstore().RemoveProtocols(id, delta.GetRmProtocols()...)
	if err != nil {
		return err
	}

	evt := event.EvtPeerProtocolsUpdated{
		Peer:    id,
		Added:   protocol.ConvertFromStrings(delta.GetAddedProtocols()),
		Removed: protocol.ConvertFromStrings(delta.GetRmProtocols()),
	}
	ids.emitters.evtPeerProtocolsUpdated.Emit(evt)
	return nil
}
