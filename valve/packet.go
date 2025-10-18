// Licensed under the GNU General Public License, version 3 or higher.
package valve

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"net"
	"time"
)

const kMaxPacketSize = 1400

var ErrOutOfBounds = errors.New("read out of bounds")

type PacketBuilder struct {
	bytes.Buffer
}

func (pb *PacketBuilder) WriteCString(str string) {
	pb.WriteString(str)
	pb.WriteByte(0)
}

func (pb *PacketBuilder) WriteBytes(bytes []byte) {
	pb.Write(bytes)
}

type PacketReader struct {
	buffer []byte
	pos    int
}

func NewPacketReader(packet []byte) *PacketReader {
	return &PacketReader{
		buffer: packet,
		pos:    0,
	}
}

func (pr *PacketReader) canRead(size int) error {
	if size+pr.pos > len(pr.buffer) {
		return ErrOutOfBounds
	}
	return nil
}

func (pr *PacketReader) Slice(count int) []byte {
	if pr.canRead(count) != nil {
		return nil
	}
	bytes := pr.buffer[pr.pos : pr.pos+count]
	pr.pos += count
	return bytes
}

func (pr *PacketReader) Pos() int {
	return pr.pos
}

func (pr *PacketReader) ReadIPv4() (net.IP, error) {
	if err := pr.canRead(net.IPv4len); err != nil {
		return nil, err
	}

	ip := net.IP(pr.buffer[pr.pos : pr.pos+net.IPv4len])
	pr.pos += net.IPv4len
	return ip, nil
}

func (pr *PacketReader) ReadPort() (uint16, error) {
	if err := pr.canRead(2); err != nil {
		return 0, err
	}

	port := binary.BigEndian.Uint16(pr.buffer[pr.pos:])
	pr.pos += 2
	return port, nil
}

func (pr *PacketReader) ReadUint8() uint8 {
	b := pr.buffer[pr.pos]
	pr.pos++
	return b
}

func (pr *PacketReader) ReadUint16() uint16 {
	u16 := binary.LittleEndian.Uint16(pr.buffer[pr.pos:])
	pr.pos += 2
	return u16
}

func (pr *PacketReader) ReadUint32() uint32 {
	u32 := binary.LittleEndian.Uint32(pr.buffer[pr.pos:])
	pr.pos += 4
	return u32
}

func (pr *PacketReader) ReadInt32() int32 {
	return int32(pr.ReadUint32())
}

func (pr *PacketReader) ReadUint64() uint64 {
	u64 := binary.LittleEndian.Uint64(pr.buffer[pr.pos:])
	pr.pos += 8
	return u64
}

func (pr *PacketReader) ReadFloat32() float32 {
	bits := pr.ReadUint32()

	return math.Float32frombits(bits)
}

func (pr *PacketReader) TryReadString() (string, bool) {
	start := pr.pos
	for pr.pos < len(pr.buffer) {
		if pr.buffer[pr.pos] == 0 {
			pr.pos++
			return string(pr.buffer[start : pr.pos-1]), true
		}
		pr.pos++
	}
	return "", false
}

func (pr *PacketReader) ReadString() string {
	start := pr.pos
	for {
		// Note: it's intended that we panic for strings that are not null
		// terminated.
		if pr.buffer[pr.pos] == 0 {
			pr.pos++
			break
		}
		pr.pos++
	}
	return string(pr.buffer[start : pr.pos-1])
}

func (pr *PacketReader) More() bool {
	return pr.pos < len(pr.buffer)
}

type UdpSocket struct {
	timeout time.Duration
	cn      net.Conn
	buffer  [kMaxPacketSize]byte
	wait    time.Duration
	next    time.Time
}

func NewUdpSocket(address string, timeout time.Duration) (*UdpSocket, error) {
	cn, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &UdpSocket{
		timeout: timeout,
		cn:      cn,
	}, nil
}

func (us *UdpSocket) SetTimeout(timeout time.Duration) {
	us.timeout = timeout
}

func (us *UdpSocket) RemoteAddr() net.Addr {
	return us.cn.RemoteAddr()
}

func (us *UdpSocket) SetRateLimit(ratePerMinute int) {
	us.wait = (time.Minute / time.Duration(ratePerMinute)) + time.Second
}

func (us *UdpSocket) extendedDeadline() time.Time {
	return time.Now().Add(us.timeout)
}

func (us *UdpSocket) enforceRateLimit() {
	if us.wait == 0 {
		return
	}

	wait := time.Until(us.next)
	if wait > 0 {
		time.Sleep(wait)
	}
}

func (us *UdpSocket) setNextQueryTime() {
	if us.wait != 0 {
		us.next = time.Now().Add(us.wait)
	}
}

func (us *UdpSocket) Send(bytes []byte) error {
	us.enforceRateLimit()
	defer us.setNextQueryTime()

	// Set timeout.
	if us.timeout > 0 {
		us.cn.SetWriteDeadline(us.extendedDeadline())
	}

	// UDP is all or nothing.
	_, err := us.cn.Write(bytes)
	return err
}

func (us *UdpSocket) Recv() ([]byte, error) {
	defer us.setNextQueryTime()

	// Set timeout.
	if us.timeout > 0 {
		us.cn.SetReadDeadline(us.extendedDeadline())
	}

	n, err := us.cn.Read(us.buffer[0:kMaxPacketSize])
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, n)
	copy(buffer, us.buffer[:n])
	return buffer, nil
}

func (us *UdpSocket) Close() {
	us.cn.Close()
}
