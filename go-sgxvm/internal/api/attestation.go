//go:build !nosgx && attestationServer
// +build !nosgx,attestationServer

package api

// #include <stdlib.h>
// #include "bindings.h"
import "C"

import (
	"fmt"
	"github.com/SigmaGmbH/librustgo/types"
	"google.golang.org/protobuf/proto"
	"net"
	"runtime"
	"sync"
)

// StartAttestationServer starts attestation server on provided address
func StartAttestationServer(dcapAddress string) error {
	fmt.Println("[Attestation Server] Trying to start attestation server")

	listener, err := net.Listen("tcp", dcapAddress)
	if err != nil {
		fmt.Println("[Attestation Server] Cannot start listener for DCAP attestation")
		return err
	}

	var mutex sync.Mutex

	go handleConnections(listener, &mutex)

	fmt.Printf("[Attestation Server] Started Attestation Server\nDCAP attestation: %s", dcapAddress)
	return nil
}

func handleConnections(listener net.Listener, mutex *sync.Mutex) {
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("[Attestation Server] DCAP listener: Got error ", err.Error(), ", connection: ", connection.RemoteAddr().String())
			connection.Close()
			continue
		}

		if err := handleIncomingRARequest(connection, mutex); err != nil {
			fmt.Println("[Attestation Server] DCAP listener: Attestation failed. Reason: ", err)
			connection.Close()
			continue
		}
	}
}

// Handles incoming request for Remote Attestation
func handleIncomingRARequest(connection net.Conn, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()
	defer connection.Close()

	println("[Attestation Server] Attesting peer: ", connection.RemoteAddr().String())

	tcpConn, ok := connection.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("[Attestation Server] connection is not a TCP connection")
	}

	// Extract file descriptor for socket
	file, err := tcpConn.File()
	if err != nil {
		fmt.Println("[Attestation Server] Cannot get access to the connection. Reason: ", err.Error())
		return err
	}

	// Create protobuf encoded request and send it to Rust side
	req := types.SetupRequest{
		Req: &types.SetupRequest_PeerAttestationRequest{
			PeerAttestationRequest: &types.PeerAttestationRequest{
				Fd: int32(file.Fd()),
			},
		},
	}

	reqBytes, err := proto.Marshal(&req)
	if err != nil {
		fmt.Println("[Attestation Server] Failed to encode req:", err)
		return err
	}

	_, err = SendProtobufRequest(reqBytes)
	return err
}

// SendProtobufRequest sends protobuf-encoded request to Rust side
func SendProtobufRequest(data []byte) (C.UnmanagedVector, error) {
	// Pass request to Rust
	d := MakeView(data)
	defer runtime.KeepAlive(data)

	errmsg := NewUnmanagedVector(nil)
	ptr, err := C.handle_initialization_request(d, &errmsg)
	if err != nil {
		return NewUnmanagedVector(nil), ErrorWithMessage(err, errmsg)
	}

	return ptr, nil
}
