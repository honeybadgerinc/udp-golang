package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"strings"
	"time"
	"unsafe"
)

// #include <sys/syscall.h>
import "C"

var messagesSent int
var messagesReceived int

const (
	PAYLOAD_SIZE = 100
	MSG_COUNT    = 1e8
)

// "The msghdr structure is used to minimize the number of directly supplied parameters to the recvmsg() and sendmsg() functions."
type MMsghdr struct {
	Msg unix.Msghdr
	cnt int
}

func main() {
	address := os.Args[1]
	localAddress := os.Args[2]

	var addressWithPort strings.Builder
	addressWithPort.WriteString(address)
	addressWithPort.WriteString(":40000")

	remoteAddress, _ := net.ResolveUDPAddr("udp4", addressWithPort.String())
	sendMMsg(remoteAddress, localAddress)
}

func sendMMsg(addr *net.UDPAddr, localAddress string) {
	start := time.Now()
	fmt.Println(start)

	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.ParseIP(localAddress),
		Port: 40000,
		Zone: "",
	})
	if err != nil {
		fmt.Println(err)
	}

	laddr := UDPAddrToSockaddr(&net.UDPAddr{Port: 1234, IP: net.IPv4zero})
	raddr := UDPAddrToSockaddr(addr)

	fd := connectToUDP(laddr, raddr)

	var inboundMsgArr [MSG_COUNT]MMsghdr
	var outboundMsgArr [MSG_COUNT]MMsghdr
	for messageCountIndex := 0; messageCountIndex < MSG_COUNT; messageCountIndex++ {
		payload := make([]byte, PAYLOAD_SIZE)
		/*
			struct iovec {
				void  *iov_base;    Starting localAddress of the buffer
				size_t iov_len 		Number of bytes to transfer
		*/
		var ioVector unix.Iovec
		// Base is the starting localAddress of a buffer.
		ioVector.Base = &payload[0]
		// Len(gth) i the length of the entire buffer.
		// So in this case, it's the length of an individual packet
		ioVector.SetLen(len(payload))

		var msg unix.Msghdr
		// Iov is the scatter/gather array
		msg.Iov = &ioVector
		// Number of messages in the Iovec
		msg.SetIovlen(1)

		outboundMsgArr[messageCountIndex] = MMsghdr{msg, 0}
		inboundMsgArr[messageCountIndex] = MMsghdr{msg, 0}
	}

	var sentPacketsCount int
	var recvPacketsCount int
	readInBuffer := make([]byte, PAYLOAD_SIZE)
	for start := time.Now(); time.Since(start) < time.Minute*10; {
		/*
			trap = C.SYS.sendmmsg: the C API to invoke
			a1 = socket file descriptor
			a2 = message vector pointer
			a3 = flags: the number of messages to send
		*/
		numPacketsSent, _, sendErr := unix.Syscall(C.SYS_sendmmsg, uintptr(fd), uintptr(unsafe.Pointer(&outboundMsgArr[0])), uintptr(MSG_COUNT))
		if sendErr != 0 {
			panic("error on sendmmsg")
		}
		fmt.Println("num. packets sent -> ", numPacketsSent)
		sentPacketsCount += int(numPacketsSent)
		bytesReadIn, _, _ := udpConn.ReadFromUDP(readInBuffer)
		fmt.Println("bytes read in -> ", bytesReadIn)
		recvPacketsCount += 1
		readInBuffer = make([]byte, PAYLOAD_SIZE)
	}
	fmt.Println("packets sent -> ", sentPacketsCount)
	fmt.Println("packets received -> ", recvPacketsCount)
	fmt.Println(time.Now().Sub(start))
}

func connectToUDP(localAddr, remoteAddr unix.Sockaddr) int {
	/*
		SOCK_DGRAM = Supports datagrams (connectionless, unreliable messages of a fixed maximum length).
		AF_INET = IPv4 Internet protocols
	*/
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	logError(err)

	/*
		SOL_SOCKET = permits sending datagram messages
		SO_REUSEADDR = allow local address reuse
	*/
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	logError(err)

	err = unix.Bind(fd, localAddr)
	logError(err)

	err = unix.Connect(fd, remoteAddr)
	logError(err)

	return fd
}

func UDPAddrToSockaddr(addr *net.UDPAddr) *unix.SockaddrInet4 {
	raddr := &unix.SockaddrInet4{Port: addr.Port, Addr: [4]byte{addr.IP[12], addr.IP[13], addr.IP[14], addr.IP[15]}}
	return raddr
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
