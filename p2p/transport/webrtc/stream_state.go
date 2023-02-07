package libp2pwebrtc

import (
	"sync"

	pb "github.com/libp2p/go-libp2p/p2p/transport/webrtc/pb"
)

type channelState uint8

const (
	stateOpen channelState = iota
	stateReadClosed
	stateWriteClosed
	stateClosed
)

type (
	webRTCStreamState struct {
		stream *webRTCStream

		mu    sync.RWMutex
		state channelState
	}
)

func (ss *webRTCStreamState) HandleInboundFlag(flag pb.Message_Flag) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.state == stateClosed {
		return
	}

	switch flag {
	case pb.Message_FIN:
		ss.closeReadInner()

	case pb.Message_STOP_SENDING:
		ss.closeWriteInner()

	case pb.Message_RESET:
		ss.closeInner(true)

	default:
		// ignore values that are invalid for flags
	}
}

func (ss *webRTCStreamState) State() channelState {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.state
}

func (ss *webRTCStreamState) AllowRead() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.state == stateOpen || ss.state == stateWriteClosed
}

func (ss *webRTCStreamState) CloseRead() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.state == stateClosed {
		return
	}

	ss.closeReadInner()
}

func (ss *webRTCStreamState) closeReadInner() {
	if ss.state == stateOpen {
		ss.state = stateReadClosed
	} else if ss.state == stateWriteClosed {
		ss.closeInner(false)
	}
}

func (ss *webRTCStreamState) AllowWrite() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.state == stateOpen || ss.state == stateReadClosed
}

func (ss *webRTCStreamState) CloseWrite() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.state == stateClosed {
		return
	}

	ss.closeWriteInner()
}

func (ss *webRTCStreamState) closeWriteInner() {
	if ss.state == stateOpen {
		ss.state = stateWriteClosed
	} else if ss.state == stateReadClosed {
		ss.closeInner(false)
	}
}

func (ss *webRTCStreamState) Closed() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.state == stateClosed
}

func (ss *webRTCStreamState) Close() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.state = stateClosed
}

func (ss *webRTCStreamState) closeInner(reset bool) {
	if ss.state != stateClosed {
		ss.state = stateClosed
		if err := ss.stream.close(reset, true); err != nil {
			log.Debugf("failed to close (reset: %v) stream: %v", reset, err)
		}
	}
}