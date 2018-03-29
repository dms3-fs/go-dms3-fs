package p2p

import (
	"io"
	"sync"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	net "gx/ipfs/QmYj8wdn5sZEHX2XMDWGBvcXJNdzVbaVpHmXvhHBVZepen/go-libp2p-net"
	manet "gx/ipfs/QmcGXGdw9BWDysPJQHxJinjGHha3eEg4vzFETre4woNwcX/go-multiaddr-net"
)

// Stream holds information on active incoming and outgoing p2p streams.
type Stream struct {
	Id uint64

	Protocol string

	OriginAddr ma.Multiaddr
	TargetAddr ma.Multiaddr

	Local  manet.Conn
	Remote net.Stream

	Registry *StreamRegistry
}

// Close closes stream endpoints and deregisters it
func (s *Stream) Close() error {
	s.Local.Close()
	s.Remote.Close()
	s.Registry.Deregister(s.Id)
	return nil
}

// Rest closes stream endpoints and deregisters it
func (s *Stream) Reset() error {
	s.Local.Close()
	s.Remote.Reset()
	s.Registry.Deregister(s.Id)
	return nil
}

func (s *Stream) startStreaming() {
	go func() {
		io.Copy(s.Local, s.Remote)
		s.Reset()
	}()

	go func() {
		_, err := io.Copy(s.Remote, s.Local)
		if err != nil {
			s.Reset()
		} else {
			s.Close()
		}
	}()
}

// StreamRegistry is a collection of active incoming and outgoing proto app streams.
type StreamRegistry struct {
	Streams map[uint64]*Stream
	lk      *sync.Mutex

	nextId uint64
}

// Register registers a stream to the registry
func (r *StreamRegistry) Register(streamInfo *Stream) {
	r.lk.Lock()
	defer r.lk.Unlock()

	streamInfo.Id = r.nextId
	r.Streams[r.nextId] = streamInfo
	r.nextId++
}

// Deregister deregisters stream from the registry
func (r *StreamRegistry) Deregister(streamId uint64) {
	r.lk.Lock()
	defer r.lk.Unlock()

	delete(r.Streams, streamId)
}