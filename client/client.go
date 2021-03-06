/*
 * Copyright (c) 2018 Simon Schmidt
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */


package client

import "net"
import "github.com/byte-mug/udpil/protocol"
import "time"

const (
	SIG_UNKNOWN = 0
	SIG_DRAIN = 1
	SIG_FULL  = 2
)

type Socket struct{
	In, Out chan interface{}
	Pkts chan *protocol.Mblk
	Brk chan int
	Sig chan int
	Restart chan int
	Closer chan DispKey
	Closee DispKey
}
func (s *Socket) Init() *Socket {
	s.In = make(chan interface{},16)
	s.Out = make(chan interface{},16)
	s.Pkts = make(chan *protocol.Mblk,16)
	s.Brk = make(chan int)
	s.Sig = make(chan int,1)
	s.Restart = make(chan int,1)
	return s
}

func (s *Socket) Input(udp *net.UDPConn) {
	for {
		select {
		case <- s.Brk: return
		default:
		}
		pkt := protocol.GetMblk()
		n,e := udp.Read(pkt.GbBody())
		if e==nil {
			pkt.GbLim(n)
			s.Pkts <- pkt
		} else {
			pkt.Dispose()
		}
	}
}
func (s *Socket) Close() {
	defer func(){ recover() }()
	close(s.Brk)
}
func (s *Socket) teardown() {
	if s.Closer!=nil { s.Closer <- s.Closee }
}
func (s *Socket) restart() {
	select {
	case s.Restart <- 1:
	default:
	}
}
func (s *Socket) Dispatch(con *protocol.Connection) {
	defer s.teardown()
	tick := time.Tick(time.Second)
	for {
		con.QueueOut(s.Out)
		con.QueueIn(s.In)
		for i := con.LenOut(); i!=0 ; i-- {
			con.Tick()
		}
		select {
		case <- s.Brk: return
		case pkt := <- s.Pkts:
			con.Rcv(pkt)
			if con.State==protocol.S_closed { s.Close() }
			if con.Ev.Reconnect {
				s.restart()
				con.Ev.Reconnect = false
			}
		case <- tick:
			con.Tmo = true
			con.Rmit = true
			con.Tick()
		case i := <- s.Sig:
			if i==SIG_FULL { continue }
			con.Rmit = true
			con.Tick()
		}
	}
}

