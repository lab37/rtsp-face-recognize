package main

import (
	"net"

	"golang.org/x/net/ipv4"
)

type message struct {
	Data []byte
	CM   *ipv4.ControlMessage
	Src  net.Addr
}

type dataST struct {
	Cmd     string `json:"cmd"`
	Model   string `json:"model"`
	Payload string `json:"data"`
}

type payload struct {
	RGB          int `json:"rgb"`
	Illumination int `json:"illumination"`
}

func udpMulticastReceiver(address string, port int, interfaceName string, messageCh chan message) error {
	localInterface, err := GetInterface(interfaceName)
	if err != nil {
		return err
	}

	udpConn, err := getUDPConnection(address, port, localInterface)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	packetConn := ipv4.NewPacketConn(udpConn)
	defer packetConn.Close()
	packetConn.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true)
	buf := make([]byte, 2048)

	for {
		n, cm, src, err := packetConn.ReadFrom(buf)
		if err != nil {
			return err
		}
		data := make([]byte, n)
		copy(data, buf)
		messageCh <- message{Data: data, CM: cm, Src: src}
	}
}

func getUDPConnection(address string, port int, localInterface *net.Interface) (*net.UDPConn, error) {
	var udpConn *net.UDPConn
	var err error
	ip := net.ParseIP(address)
	udpAddr := &net.UDPAddr{IP: ip, Port: port}
	if ip.IsMulticast() {
		udpConn, err = net.ListenMulticastUDP("udp", localInterface, udpAddr)
	} else {
		udpConn, err = net.ListenUDP("udp", udpAddr)
	}
	return udpConn, err
}

// GetInterface returns the interface associated with the name provided.
// It wraps the net.InterfaceByName only adding functionality to allow
// the specified interface to be an empty string. This helps with command
// line processing where the default value is an empty string.
func GetInterface(interfaceName string) (*net.Interface, error) {
	var localInterface *net.Interface
	var err error
	if interfaceName != "" {
		localInterface, err = net.InterfaceByName(interfaceName)
	}
	return localInterface, err
}
