package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

func main() {
	server, err := net.Listen("tcp", ":1080")
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
		return
	}
	defer server.Close()

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Println("Accept failed: %v", err)
			continue
		}
		go process(client)
	}
}

/* Process the connection. */
func process(client net.Conn) {
	defer client.Close()

	// remoteAddr := client.RemoteAddr().String()
	// fmt.Printf("Connection from %s\n", remoteAddr)

	// client.Write([]byte("Hello world!"))
	err := Socks5_Auth(client)
	if err != nil {
		fmt.Println("Authorization failed: %v", err)
		return
	}

	target, err := Socks5_Connect(client)
	if err != nil {
		fmt.Println("Connection failed: %v", err)
		return
	}

	Socks5_Forward(client, target)
}

func Socks5_Auth(client net.Conn) error {
	var buf [512]byte

	// Read in first 2 byte of header information.
	n, err := client.Read(buf[:2])
	if n != 2 {
		return errors.New("Header Error " + err.Error())
	}

	// Check the version (Socks 5 only)
	ver, Nmethod := int(buf[0]), int(buf[1])
	if ver != 0x05 {
		return errors.New("Invalid Version!")
	}

	// Read in the method data.
	n, err = client.Read(buf[:Nmethod])
	if Nmethod != n {
		return errors.New("Method length mismatch!")
	}

	// We choose 0x00 method
	n, err = client.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("Write error" + err.Error())
	}

	return nil
}

func Socks5_Connect(client net.Conn) (net.Conn, error) {
	var buf [512]byte

	// Read in first 2 byte of header information.
	n, err := client.Read(buf[:4])
	if n != 4 {
		return nil, errors.New("Header Error " + err.Error())
	}

	// Check version,cmd and reservation
	ver, cmd, rsv, atyp := int(buf[0]), int(buf[1]), int(buf[2]), int(buf[3])
	if ver != 0x05 || cmd == 0 || rsv != 0x00 {
		return nil, errors.New("Invalid ver/cmd/rsv!")
	}

	// Get the full string and the buffer first
	n, err = client.Read(buf[:])
	if n < 2 {
		return nil, errors.New("Invalid address!" + err.Error())
	}

	// port number from 0 ~ 65535.
	port := binary.BigEndian.Uint16(buf[n-2 : n])
	addr := ""

	// Check the atyp
	switch atyp {
	case 0x01: // ipv4 case
		if n != 6 {
			return nil, errors.New("Invalid ipv4 address!")
		}
		addr = fmt.Sprintf("%d.%d.%d.%d:%d", buf[0], buf[1], buf[2], buf[3], port)
		break

	case 0x03: // domain case
		if n != int(buf[0])+3 {
			return nil, errors.New("Invalid domain!")
		}
		addr = string(buf[0:n-2]) + fmt.Sprintf(":%v", port)
		break

	case 0x04: // ipv6 case
		if n != 18 {
			return nil, errors.New("Invalid ipv6 address!")
		}
		addr = "["
		addr += fmt.Sprintf("%x:", binary.BigEndian.Uint16(buf[0:2]))
		addr += fmt.Sprintf("%x:", binary.BigEndian.Uint16(buf[2:4]))
		addr += fmt.Sprintf("%x:", binary.BigEndian.Uint16(buf[4:6]))
		addr += fmt.Sprintf("%x:", binary.BigEndian.Uint16(buf[6:8]))
		addr += fmt.Sprintf("%x:", binary.BigEndian.Uint16(buf[8:10]))
		addr += fmt.Sprintf("%x:", binary.BigEndian.Uint16(buf[10:12]))
		addr += fmt.Sprintf("%x:", binary.BigEndian.Uint16(buf[12:14]))
		addr += fmt.Sprintf("%x", binary.BigEndian.Uint16(buf[14:16]))
		addr += fmt.Sprintf("]:%d", port)
		break
	default: // Error case
		return nil, errors.New("Invalid aytp!")
	}

	// Tries to dial the address
	dest, err := net.Dial("TCP", addr)
	if err != nil {
		return nil, err
	}

	// Send a message to client
	n, err = client.Write([]byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		dest.Close()
		return nil, err
	}

	return dest, nil
}

func Socks5_Forward(client, target net.Conn) {
	// Forward 2 connection
	go forward(client, target)
	go forward(target, client)
}

func forward(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}
