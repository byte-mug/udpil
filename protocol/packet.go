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

import "io"
import "sync"
import "fmt"

type Mblk struct{
	base []byte
	pos, lim int
	refc uint
}
func (m *Mblk) Clear() *Mblk {
	m.pos = 0
	m.lim = len(m.base)
	return m
}
func (m *Mblk) Read(b []byte) (int,error) {
	n := m.lim-m.pos
	o := len(b)
	if n==0 { return 0,io.EOF }
	if n>o { n = o }
	copy(b,m.base[m.pos:m.lim])
	m.pos += n
	return n,nil
}
func (m *Mblk) Write(b []byte) (int,error) {
	n := m.lim-m.pos
	o := len(b)
	if n==0 { return 0,io.EOF }
	if n>o { n = o }
	copy(m.base[m.pos:m.lim],b)
	m.pos += n
	if n<o { return n,io.EOF }
	return n,nil
}
func (m *Mblk) Flip() *Mblk {
	m.pos = m.lim
	m.pos = 0
	return m
}
func (m *Mblk) Pos() int { return m.pos }
func (m *Mblk) Lim() int { return m.lim }

func (m *Mblk) SetPos(i int) *Mblk {
	m.pos = i
	return m
}
func (m *Mblk) SetLim(i int) *Mblk {
	m.lim = i
	return m
}
func (m *Mblk) Len() int {
	return m.lim-m.pos
}
func (m *Mblk) Packet() []byte {
	return m.base[:m.lim]
}
func (m *Mblk) Body() []byte {
	return m.base[m.pos:m.lim]
}
func (m *Mblk) Consumed() []byte {
	return m.base[:m.pos]
}
func (m *Mblk) GbBody() []byte {
	return m.base[m.pos:]
}
func (m *Mblk) GbLim(i int) *Mblk {
	m.lim = m.pos + i
	return m
}
func (m *Mblk) Dispose() {
	m.refc--
	if m.refc!=0 { return }
	msgBlocks.Put(m)
}
func (m *Mblk) Retain() {
	m.refc++
}
func (m *Mblk) String() string {
	return fmt.Sprintf("{ %p %d %d }",m.base,m.pos,m.lim)
}

var msgBlocks = sync.Pool{New: func() interface{} {
	return &Mblk{base:make([]byte,1220)}
}}

func GetMblk() *Mblk {
	m := msgBlocks.Get().(*Mblk)
	m.Clear()
	m.refc = 1
	return m
}

func GetMsgBytes(i interface{}) []byte {
	switch v := i.(type) {
	case []byte: return v
	case *Mblk: return v.Body()
	}
	return nil
}
func FreeMsg(i interface{}) {
	v,ok := i.(*Mblk)
	if !ok { return }
	if v==nil { return }
	v.Dispose()
}

var headers = sync.Pool{New: func() interface{} {
	return new(Header)
}}
func getHeader() *Header {
	h := headers.Get().(*Header)
	*h = Header{} // Clear header.
	return h
}

