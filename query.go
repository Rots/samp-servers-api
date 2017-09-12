// Some of the code in this module was from urShadow, it was adapted and modified 2017-07-01

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// QueryType represents a query method from the SA:MP set: i, r, c, d, x, p
type QueryType uint8

const (
	// Info is the 'i' packet type
	Info QueryType = 'i'
	// Rules is the 'r' packet type
	Rules QueryType = 'r'
	// Players is the 'c' packet type
	Players QueryType = 'c'
)

// LegacyQuery stores state for old-style masterlist queries
type LegacyQuery struct {
	addr    *net.UDPAddr
	conn    net.Conn
	Timeout time.Duration
}

// GetServerLegacyInfo wraps a set of legacy queries and returns a new Server object with the
// available fields populated.
func GetServerLegacyInfo(host string) (server Server, err error) {
	lq, err := NewLegacyQuery(host, time.Second*5)
	defer func() {
		err := lq.Close()
		if err != nil {
			logger.Fatal("failed to close legacy query handler",
				zap.String("host", host),
				zap.Error(err))
		}
	}()
	if err != nil {
		return server, err
	}

	server.Core, err = lq.GetInfo()
	if err != nil {
		return server, err
	}
	server.Core.Address = host

	server.Rules, err = lq.GetRules()
	if err != nil {
		return server, err
	}

	if server.Core.Players < 100 {
		server.PlayerList, err = lq.GetPlayers()
		if err != nil {
			return server, err
		}
	}

	return server, err
}

// NewLegacyQuery creates a new legacy query handler for a server
func NewLegacyQuery(host string, timeout time.Duration) (lq *LegacyQuery, err error) {
	lq = new(LegacyQuery)
	lq.addr, err = net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve: %v", err)
	}

	lq.conn, err = net.DialUDP("udp", nil, lq.addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	lq.Timeout = timeout

	return lq, nil
}

// Close closes a legacy query manager's connection
func (lq *LegacyQuery) Close() error {
	if lq != nil {
		return lq.conn.Close()
	}
	return nil
}

// SendQuery writes a SA:MP format query with the specified opcode, returns the raw response bytes
func (lq *LegacyQuery) SendQuery(opcode QueryType) ([]byte, error) {
	var (
		err      error
		response = make([]byte, 2048)
		request  = new(bytes.Buffer)
		n        int
	)

	port := [2]byte{
		byte(lq.addr.Port & 0xFF),
		byte((lq.addr.Port >> 8) & 0xFF),
	}

	binary.Write(request, binary.LittleEndian, []byte("SAMP"))
	binary.Write(request, binary.LittleEndian, lq.addr.IP.To4())
	binary.Write(request, binary.LittleEndian, port[0])
	binary.Write(request, binary.LittleEndian, port[1])
	binary.Write(request, binary.LittleEndian, opcode)

	_, err = lq.conn.Write(request.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write: %v", err)
	}

	waitRead := time.After(lq.Timeout)
	waitResponse := make(chan []byte)
	waitError := make(chan error)
	go func() {
		n, err = lq.conn.Read(response)
		if err != nil {
			waitError <- fmt.Errorf("failed to read response: %v", err)
		}
		if n > cap(response) {
			waitError <- fmt.Errorf("read response over buffer capacity")
		}
		waitResponse <- response
	}()

waiter:
	for {
		select {
		case <-waitRead:
			return nil, fmt.Errorf("socket read timed out")
		case err := <-waitError:
			return nil, err

		case response = <-waitResponse:
			break waiter
		}
	}

	return response[:n], nil
}

// GetInfo returns the core server info for displaying on the browser list.
func (lq *LegacyQuery) GetInfo() (server ServerCore, err error) {
	response, err := lq.SendQuery(Info)
	if err != nil {
		return server, err
	}

	ptr := 11

	server.Password = (response[ptr] == 1)
	ptr++

	server.Players = int(binary.LittleEndian.Uint16(response[ptr : ptr+2]))
	ptr += 2

	server.MaxPlayers = int(binary.LittleEndian.Uint16(response[ptr : ptr+2]))
	ptr += 2

	hostnameLen := int(binary.LittleEndian.Uint16(response[ptr : ptr+4]))
	ptr += 4

	server.Hostname = string(response[ptr : ptr+hostnameLen])
	ptr += hostnameLen

	gamemodeLen := int(binary.LittleEndian.Uint16(response[ptr : ptr+4]))
	ptr += 4

	server.Gamemode = string(response[ptr : ptr+gamemodeLen])
	ptr += gamemodeLen

	languageLen := int(binary.LittleEndian.Uint16(response[ptr : ptr+4]))
	ptr += 4

	if languageLen > 0 {
		server.Language = string(response[ptr : ptr+languageLen])
		// ptr += languageLen
	} else {
		server.Language = "-"
	}

	return
}

// GetRules returns a map of rule properties from a server. The legacy query uses established keys
// such as "Map" and "Version"
func (lq *LegacyQuery) GetRules() (rules map[string]string, err error) {
	response, err := lq.SendQuery(Rules)
	if err != nil {
		return rules, err
	}

	rules = make(map[string]string)

	var (
		key string
		val string
		len int
	)

	ptr := 11
	amount := binary.LittleEndian.Uint16(response[ptr : ptr+2])
	ptr += 2

	for i := uint16(0); i < amount; i++ {
		len = int(response[ptr])
		ptr++

		key = string(response[ptr : ptr+len])
		ptr += len

		len = int(response[ptr])
		ptr++

		val = string(response[ptr : ptr+len])
		ptr += len

		rules[key] = val
	}

	return
}

// GetPlayers simply returns a slice of strings, score is rather arbitrary so it's omitted.
func (lq *LegacyQuery) GetPlayers() (players []string, err error) {
	response, err := lq.SendQuery(Players)
	if err != nil {
		return nil, err
	}

	var (
		count  uint16
		length int
	)

	ptr := 11
	count = binary.LittleEndian.Uint16(response[ptr : ptr+2])
	ptr += 2

	players = make([]string, count)

	for i := uint16(0); i < count; i++ {
		length = int(response[ptr])
		ptr++

		players[i] = string(response[ptr : ptr+length])
		ptr += length
		ptr += 4 // score, unused
	}

	return players, nil
}
