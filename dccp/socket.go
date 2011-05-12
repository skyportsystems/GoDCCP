// Copyright 2010 GoDCCP Authors. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package dccp

import "rand"

// socket is a data structure, maintaining the DCCP socket variables.
// socket's methods are not re-entrant
type socket struct {
	ISS uint64 // Initial Sequence number Sent
	ISR uint64 // Initial Sequence number Received

	OSR uint64 // First OPEN Sequence number Received

	// Here and elsewhere, "greatest" is measured in circular sequence space (modulo 2^48)
	GSS uint64 // Greatest Sequence number Sent

	GSR uint64 // Greatest valid Sequence number Received (consequently, sent as AckNo back)
	GAR uint64 // Greatest valid Acknowledgement number Received on a non-Sync; initialized to S.ISS

	// XXX: Not set by Conn
	CCIDA byte // CCID in use for the A-to-B half-connection, Section 10
	CCIDB byte // CCID in use for the B-to-A half-connection, Section 10

	// XXX: Not set by Conn
	SWBF uint64 // Sequence Window/B Feature, see Section 7.5.1
	SWAF uint64 // Sequence Window/A Feature, see Section 7.5.1

	State       int
	Server      bool   // True if the endpoint is a server, false if it is a client
	ServiceCode uint32 // The service code of this connection

	// XXX: Not set by Conn
	PMTU  int // Path Maximum Transmission Unit
	CCMPS int // Congestion Control Maximum Packet Size
	MPS   int // Maximum Packet Size = min(PMTU, CCMPS)
	RTT   int // Round Trip Time
}

const (
	MSL                    = 2 * 60e9 // 2 mins in nanoseconds
	PARTOPEN_BACKOFF_FIRST = 200e6    // 200 miliseconds in nanoseconds, Section 8.1.5
	PARTOPEN_BACKOFF_MAX   = 4 * MSL  // 8 mins in nanoseconds, Section 8.1.5
)

// The nine possible states of a DCCP socket.  Listed in increasing order:
const (
	CLOSED = iota
	LISTEN
	REQUEST
	RESPOND
	PARTOPEN
	OPEN
	CLOSEREQ
	CLOSING
	TIMEWAIT
)

func (s *socket) SetServer(v bool) { s.Server = v }
func (s *socket) IsServer() bool   { return s.Server }

func (s *socket) GetState() int { return s.State }

func (s *socket) SetState(v int) { s.State = v }

// ChooseISS chooses a safe Initial Sequence Number
func (s *socket) ChooseISS() uint64 {
	iss := uint64(rand.Int63()) & 0xffffff
	s.ISS = iss
	return iss
}

func (s *socket) SetISR(v uint64) { s.ISR = v }

func (s *socket) GetOSR() uint64  { return s.OSR }
func (s *socket) SetOSR(v uint64) { s.OSR = v }

func (s *socket) GetGSS() uint64  { return s.GSS }
func (s *socket) SetGSS(v uint64) { s.GSS = v }

func (s *socket) GetGSR() uint64     { return s.GSR }
func (s *socket) SetGSR(v uint64)    { s.GSR = v }
func (s *socket) UpdateGSR(v uint64) { s.GSR = maxu64(s.GSR, v) }

func (s *socket) GetGAR() uint64     { return s.GAR }
func (s *socket) SetGAR(v uint64)    { s.GAR = v }
func (s *socket) UpdateGAR(v uint64) { s.GAR = maxu64(s.GAR, v) }

func maxu64(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

// TODO: Address the last paragraph of Section 7.5.1 regarding SWL,AWL calculation

// GetSWLH() computes SWL and SWH, see Section 7.5.1
func (s *socket) GetSWLH() (SWL uint64, SWH uint64) {
	return maxu64(s.GSR+1-s.SWBF/4, s.ISR), s.GSR + (3*s.SWBF)/4
}

// GetAWLH() computes AWL and AWH, see Section 7.5.1
func (s *socket) GetAWLH() (AWL uint64, AWH uint64) {
	return maxu64(s.GSS+1-s.SWAF, s.ISS), s.GSS
}

func (s *socket) InAckWindow(x uint64) bool {
	awl, awh := s.GetAWLH()
	return awl <= x && x <= awh
}
