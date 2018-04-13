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



package protocol

import "encoding/binary"
import "io"
import "math/rand"
import "fmt"

type Header struct{
	Illen  uint16
	Ilflg  uint16
	Ilid   uint32
	Ilack  uint32
	Ilmax  uint32
	Ilrs   uint16
	Ilrd   uint16
}
// 20 bytes
func (h *Header) swapid() {
	h.Ilid,h.Ilack = h.Ilack,h.Ilid
}
func (h *Header) String() string {
	return fmt.Sprint("{",
		" Len:",h.Illen,
		" Flg:",h.Ilflg,
		" Id :",h.Ilid,
		" Ack:",h.Ilack,
		" Max:",h.Ilmax,
		" Rs :",h.Ilrs,
		" Rd :",h.Ilrd,
		" }")
}

const HeaderLength = 20

const (
	F_data uint16 = 1<<iota
	F_ack
	F_keep
	F_respin
	F_reset
)

const (
	S_closed uint = iota
	S_open
	S_taken /* Trying to open a connection. */
)

type ConCap struct {
	Reconnect bool
	Autoflush bool
}

type Connection struct {
	Out io.Writer
	State uint
	
	Idl,Idr uint32
	Ctl,Ctr uint32
	Icl uint32
	
	Rsid,Rsmx uint32
	Recvq InQueue
	Ack,Tmo,Rmit,Respin,Reset,Close bool
	Ractr uint
	Rectr uint
	
	Tsid,Tsmx,Tsln,Tsct uint32
	Txq TxQueue

	Ilrl,Ilrr uint16
	
	Ev struct {
		Reconnect bool
	}
	
	Cap ConCap
}

func (con *Connection) TxInfo() string {
	return fmt.Sprint("id:",con.Tsid," mx:",con.Tsmx," len:",con.Tsln," ct:",con.Tsct)
}
func (con *Connection) Init() {
	con.Recvq.Init()
	con.Txq.Init()
}

func (con *Connection) iSnd(hdr *Header,d []byte) {
	pkt := GetMblk()
	defer pkt.Dispose()
	
	hdr.Illen = uint16(len(d))
	
	binary.Write(pkt,binary.BigEndian,hdr)
	pkt.Write(d)
	
	pkt.Flip()
	
	con.Out.Write(pkt.Body())
}


func (con *Connection) check(hdr *Header) bool {
	a := (hdr.Ilrs==0)||(hdr.Ilrs==con.Ilrr)
	b := (hdr.Ilrd==0)||(hdr.Ilrd==con.Ilrl)
	if con.Ilrr==0 { con.Ilrr = hdr.Ilrs; a = true }
	return a&&b
}
func (con *Connection) prep(hdr *Header) {
	hdr.Ilrs = con.Ilrl
	hdr.Ilrd = con.Ilrr
}

func (con *Connection) Touch() {
	switch con.State {
	case S_taken:
		con.State = S_open
		con.Ilrl = uint16(rand.Int31()&0xffff)
		con.Ilrr = 0
		con.Rsid = 0
		con.Rsmx = 32
		con.Tsid = 0
		con.Tsmx = 32
		con.Tsln = 0
		con.Tsct = 1
		
		// Clear Receive Queue
		con.Recvq.Clear()
		
		// Clear Transmit Queue
		con.Txq.Clear()
	}
}
func (con *Connection) Rcv(d *Mblk) {
	hdr := getHeader()
	defer d.Dispose()
	defer headers.Put(hdr)
	
	if binary.Read(d.SetPos(0),binary.BigEndian,hdr)!=nil { return }
	
	retry:
	switch con.State {
	case S_closed,S_taken:
		if !isrange(0,32,hdr.Ilid) {
			con.Reset = true
			con.Tick()
			return
		}
		
		con.State = S_open
		if con.State==S_closed {
			con.Ilrl = uint16(rand.Int31()&0xffff)
			con.Ilrr = 0
		}
		con.Rsid = 0
		con.Rsmx = 32
		con.Tsid = 0
		con.Tsmx = 32
		con.Tsln = 0
		con.Tsct = 1
		
		// Clear Receive Queue
		con.Recvq.Clear()
		
		// Clear Transmit Queue
		con.Txq.Clear()
		if !con.check(hdr) {
			con.Reset = true
			con.Tick()
			return
		}
	case S_open:
		if !con.check(hdr) {
			//fmt.Println("!check",hdr)
			con.State = S_closed
			if con.Cap.Reconnect {
				con.Ev.Reconnect = true
				goto retry
			}
		}
	}
	
	ilflg := hdr.Ilflg
	
	if fchk(ilflg,F_reset) {
		con.State = S_closed
	}
	
	/* Handle the ack flag */
	if fchk(ilflg,F_ack) {
		nTsid := hdr.Ilack
		nTsmx := hdr.Ilmax
		if !isrange(con.Tsid,con.Tsln,nTsid) { goto ack_done }
		if !islower(nTsid,nTsmx) { goto ack_done }
		
		/* Pull queue. */
		con.Txq.Shift(uint(nTsid-con.Tsid))
		
		con.Tsid = nTsid
		con.Tsmx = nTsmx
		if fchk(ilflg,F_respin) || !isrange(con.Tsid,con.Tsln,con.Tsct) { con.Tsct = con.Tsid+1 }
	}
	ack_done:
	
	/* Handle the data flag */
	if fchk(ilflg,F_data) {
		if !isrange(con.Rsid,con.Rsmx,hdr.Ilid) { return }
		
		pos := con.Recvq.Lkupout(relative(con.Rsid,hdr.Ilid),true)
		if nil!=*pos { // Duplicate Packet: Handle Ack!
			con.Rmit = true
			con.Ack = true
			con.Tick()
			return
		}
		
		pl := d
		if pl.Len()<int(hdr.Illen) { return } /* Payload too short! */
		pl.GbLim(int(hdr.Illen)) /* Cut payload. */
		
		pl.Retain() /* Revert the deferred dispose. */
		*pos = pl
		
		n := con.Recvq.ShiftO2I()
		if n!=0 {
			con.Rsid += uint32(n)
			con.Ractr += n
			con.Respin = con.Recvq.Findinuse(uint( con.Rsmx - con.Rsid ))
			con.Rsmx = con.Rsid + uint32(window(64,con.Recvq.LenI(),32))
			con.Ack = true
		} else {
			con.Respin = con.Recvq.Findinuse(uint( con.Rsmx - con.Rsid ))
			con.Rectr ++
		}
		con.Rectr++
		if con.Rectr >= 8 {
			con.Rectr = 0
			con.Tick()
		}
	}
}

func (con *Connection) Tick() {
	var data []byte
	hdr := getHeader()
	defer headers.Put(hdr)
	con.prep(hdr)
	hdr.Ilid = con.Tsct
	
	switch con.State {
	case S_closed,S_taken:
		if hdr.Ilid==0 {
			hdr.Ilid = 1
		}
	}
	
	if con.Respin&&con.Rmit {
		hdr.Ilflg |= F_ack|F_respin
		hdr.Ilack = con.Rsid
		hdr.Ilmax = con.Rsmx
	}
	
	if isrange(con.Tsid,con.Tsln,con.Tsct) {
		v := con.Txq.Get(relative(con.Tsid,con.Tsct))
		if b,ok := v.([]byte); ok {
			hdr.Ilflg |= F_data
			con.Tsct++
			data = b
		} else if b,ok := v.(*Mblk); ok {
			hdr.Ilflg |= F_data
			con.Tsct++
			data = b.Body()
		}
	}
	
	// Piggy-Back an ack.
	if con.Ack || hdr.Ilflg!=0 {
		hdr.Ilflg |= F_ack
		hdr.Ilack = con.Rsid
		hdr.Ilmax = con.Rsmx
		con.Ractr = 0
	}
	
	if con.Tmo && hdr.Ilflg==0 {
		hdr.Ilflg |= F_keep
	}
	
	if con.Reset {
		hdr.Ilflg |= F_reset
	}
	
	con.Ack    = false
	con.Tmo    = false
	con.Rmit   = false
	con.Respin = false
	con.Reset  = false
	con.Close  = false
	
	if hdr.Ilflg!=0 { con.iSnd(hdr,data) }
}

func (con *Connection) QueueOut(queue <- chan interface{}) {
	for isrange(con.Tsid,con.Tsmx,con.Tsln+1) {
		select {
		case e := <- queue:
			con.Txq.PushBack(e)
			con.Tsln++
		default:
			return
		}
	}
}
func (con *Connection) LenOut() uint {
	return uint(con.Tsln-con.Tsid)
}
func (con *Connection) QueueIn(queue chan <- interface{}) {
	con.Recvq.ShiftIn(queue)
	ill := con.Recvq.LenI()
	con.Rsmx = con.Rsid + uint32(window(64,ill,32))
	if ill==0 {
		con.Rmit = true
		con.Respin = true
	}
}

