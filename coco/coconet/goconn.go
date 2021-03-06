package coconet

import (
	"errors"
	"fmt"
	"sync"

	"github.com/dedis/crypto/abstract"
)

// directory is a testing structure for the goConn. It allows us to simulate
// tcp network connections locally (and is easily adaptable for network
// connections). A single directory should be shared between all goConns's that
// are operating in the same network 'space'. If they are on the same tree they
// should share a directory.
type GoDirectory struct {
	sync.Mutex                        // protects accesses to channel and nameToPeer
	channel    map[string]chan []byte // one channel per peer-to-peer connection
	nameToPeer map[string]*GoConn     // keeps track of duplicate connections
}

/// newDirectory creates a new directory for registering goPeers.
func NewGoDirectory() *GoDirectory {
	return &GoDirectory{channel: make(map[string]chan []byte),
		nameToPeer: make(map[string]*GoConn)}
}

// goConn is a Conn type for representing connections in an in-memory tree. It
// uses channels for communication.
type GoConn struct {
	// the directory maps each (from,to) pair to a channel for sending
	// (from,to). When receiving one reads from the channel (from, to). Thus
	// the sender "owns" the channel.
	dir    *GoDirectory
	from   string
	to     string
	fromto string
	tofrom string

	mupk   sync.RWMutex
	pubkey abstract.Point
}

// PeerExists is an ignorable error that says that this peer has already been
// registered to this directory.
var PeerExists error = errors.New("peer already exists in given directory")

// NewGoPeer creates a goPeer registered in the given directory with the given
// hostname. It returns an ignorable PeerExists error if this peer already
// exists.
func NewGoConn(dir *GoDirectory, from, to string) (*GoConn, error) {
	gc := &GoConn{dir, from, to, from + "::::" + to, to + "::::" + from, sync.RWMutex{}, nil}
	dir.Lock()
	fromto := gc.FromTo()
	tofrom := gc.ToFrom()
	defer dir.Unlock()
	if c, ok := dir.nameToPeer[fromto]; ok {
		// return the already existant peer\
		return c, PeerExists
	}
	dir.nameToPeer[fromto] = gc
	if _, ok := dir.channel[fromto]; !ok {
		dir.channel[fromto] = make(chan []byte, 1)
	}
	if _, ok := dir.channel[tofrom]; !ok {
		dir.channel[tofrom] = make(chan []byte, 1)
	}
	return gc, nil
}

// Name returns the from+to identifier of the goConn.
func (c *GoConn) Name() string {
	return c.to
}

func (c *GoConn) FromTo() string {
	return c.fromto
}

func (c *GoConn) ToFrom() string {
	return c.tofrom
}

func (c *GoConn) Connect() error {
	return nil
}

func (c *GoConn) Close() {}

func (c *GoConn) SetPubKey(pk abstract.Point) {
	c.mupk.Lock()
	c.pubkey = pk
	c.mupk.Unlock()
}

func (c *GoConn) PubKey() abstract.Point {
	c.mupk.Lock()
	pl := c.pubkey
	c.mupk.Unlock()
	return pl
}

// Put sends data to the goConn through the channel.
func (c *GoConn) Put(data BinaryMarshaler) error {
	fromto := c.FromTo()
	c.dir.Lock()
	ch := c.dir.channel[fromto]
	// the directory must be unlocked before sending data. otherwise the
	// receiver would not be able to access this channel from the directory
	// either.
	c.dir.Unlock()
	b, err := data.MarshalBinary()
	if err != nil {
		return err
	}
	ch <- b
	return nil
}

// Get receives data from the sender.
func (c *GoConn) Get(bum BinaryUnmarshaler) error {
	tofrom := c.ToFrom()
	c.dir.Lock()
	ch := c.dir.channel[tofrom]
	// as in Put directory must be unlocked to allow other goroutines to reach
	// their send lines.
	c.dir.Unlock()

	data := <-ch
	err := bum.UnmarshalBinary(data)
	if err != nil {
		fmt.Println("failed to unmarshal binary from ", tofrom, ch, data)
		fmt.Printf("\tinto: %#v\n", bum)
		bum = nil
	} else {
		// fmt.Println("correct unmarshal binary from ", tofrom, ch, data)
	}
	return err
}
