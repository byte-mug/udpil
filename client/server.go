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

type PacketWrapper struct{
	Conn *net.UDPConn
	Addr *net.UDPAddr
}
func (pw *PacketWrapper) Write(b []byte) (int, error) {
	return pw.Conn.WriteToUDP(b,pw.Addr)
}

type ServerSocket struct{
	In chan *Socket
	Brk chan int
	Socks map[DispKey]*Socket
	Closer chan DispKey
	Cap protocol.ConCap
}

func (s *ServerSocket) Init() *ServerSocket {
	s.In = make(chan *Socket,16)
	s.Brk = make(chan int)
	s.Socks = make(map[DispKey]*Socket)
	s.Closer = make(chan DispKey,128)
	return s
}
func (s *ServerSocket) Input(udp *net.UDPConn) {
	var dk DispKey
	var con *protocol.Connection
	var pw  *PacketWrapper
	for {
		select {
		case <- s.Brk: return
		case dk = <- s.Closer:
			if sck,ok := s.Socks[dk]; ok {
				select {
				case <- sck.Brk:
					delete(s.Socks,dk)
				default:
				}
			}
			continue
		default:
		}
		pkt := protocol.GetMblk()
		n,addr,e := udp.ReadFromUDP(pkt.GbBody())
		if e!=nil { pkt.Dispose(); continue }
		pkt.GbLim(n)
		
		/* ===== Begin Packet Processing here ===== */
		
		dk.From(addr)
		
		if sck,ok := s.Socks[dk]; ok {
			select {
			case <- sck.Brk:
				delete(s.Socks,dk)
			default:
				sck.Pkts <- pkt
				continue
			}
		}
		
		if pw==nil { pw = &PacketWrapper{Conn:udp} }
		
		pw.Addr = addr
		
		if con==nil {
			con = new(protocol.Connection)
			con.Init()
			con.Cap = s.Cap
		}
		con.Out = pw
		con.State = protocol.S_closed
		
		con.Rcv(pkt)
		
		if con.State != protocol.S_closed {
			sck := new(Socket).Init()
			sck.Closer = s.Closer
			sck.Closee = dk
			go sck.Dispatch(con)
			s.Socks[dk] = sck
			con = nil
			pw = nil
			s.In <- sck
		} else {
			pw.Addr = nil
		}
	}
}

