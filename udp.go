package udp

import (
	"context"
	"net"
	"sync"

	tpt "github.com/libp2p/go-libp2p-transport"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	mafmt "github.com/whyrusleeping/mafmt"
)

type UDPTransport struct {
	connslock sync.Mutex
	conns     map[string]*UDPConn
}

func NewUDPTransport() *UDPTransport {
	return &UDPTransport{
		conns: make(map[string]*UDPConn),
	}
}

func (t *UDPTransport) Dialer(laddr ma.Multiaddr, _ ...tpt.DialOpt) (tpt.PacketDialer, error) {
	dialer, err := t.Listen(laddr)
	if err != nil {
		return nil, err
	}
	return dialer.(tpt.PacketDialer), nil
}

func (t *UDPTransport) Listen(laddr ma.Multiaddr) (tpt.PacketConn, error) {
	t.connslock.Lock()
	defer t.connslock.Unlock()
	s, ok := t.conns[laddr.String()]
	if ok {
		return s, nil
	}

	rawconn, err := manet.ListenUDP(laddr)
	if err != nil {
		return nil, err
	}

	conn := &UDPConn{
		UDPConn:   rawconn,
		transport: t,
	}

	t.conns[laddr.String()] = conn
	return conn, nil
}

func (t *UDPTransport) Matches(addr ma.Multiaddr) bool {
	return mafmt.UDP.Matches(addr)
}

type UDPConn struct {
	*manet.UDPConn
	transport *UDPTransport
}

func (c *UDPConn) ReadMsg(b []byte) (int, ma.Multiaddr, error) {
	n, _, _, netaddr, err := c.UDPConn.ReadMsgUDP(b, []byte{})
	if err != nil {
		return n, nil, err
	}

	raddr, err := manet.FromNetAddr(netaddr)
	if err != nil {
		return n, nil, err
	}

	return n, raddr, nil
}

func (c *UDPConn) WriteMsg(b []byte, raddr ma.Multiaddr) (int, error) {
	network, netaddr, err := manet.DialArgs(raddr)
	if err != nil {
		return 0, err
	}

	udpaddr, err := net.ResolveUDPAddr(network, netaddr)
	if err != nil {
		return 0, err
	}

	n, _, err := c.UDPConn.WriteMsgUDP(b, nil, udpaddr)
	return n, err
}

func (c *UDPConn) Dial(raddr ma.Multiaddr) (tpt.PacketConn, error) {
	return c, nil
}

func (c *UDPConn) DialContext(ctx context.Context, raddr ma.Multiaddr) (tpt.PacketConn, error) {
	return c, nil
}

func (C *UDPConn) Matches(addr ma.Multiaddr) bool {
	return mafmt.UDP.Matches(addr)
}

func (c *UDPConn) Transport() tpt.PacketTransport { return c.transport }

var _ tpt.PacketTransport = (*UDPTransport)(nil)
var _ tpt.PacketDialer = (*UDPConn)(nil)
var _ tpt.PacketConn = (*UDPConn)(nil)
