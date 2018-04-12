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

type InQueue struct {
	free,inner,outer Ring
}
func (q *InQueue) Init() {
	q.free.InsertAfter(q.inner.Self())
	q.inner.InsertAfter(q.outer.Self())
}
func (q *InQueue) Clear() {
	/*
	Before ...<->F<->Free...<->I<->Inner...<->O<->Outer...
	After ...<->I<->O<->F<->Free...<->Inner...<->Outer...
	*/
	
	/* Turn Inner<->...<->Outer to Inner<->Outer<->... */
	q.inner.InsertAfter(q.outer.Unlink())
	
	/* Turn Inner<->Outer<->...<->Free into Inner<->Outer<->Free<->...*/
	q.outer.InsertAfter(q.free.Unlink())
}
func (q *InQueue) alloc() *Ring {
	r := q.free.Next()
	if r==q.inner.Self() { return new(Ring) }
	return r.Unlink()
}
func (q *InQueue) Lkupout(i uint, create bool) *interface{} {
	for e := q.outer.Next(); e!=q.free.Self(); e = e.Next() {
		if i!=0 {
			i--
			continue
		}
		return &(e.Value)
	}
	if !create { return nil }
	for {
		e := q.free.InsertBefore(q.alloc())
		e.Value = nil
		if i!=0 {
			i--
			continue
		}
		return &(e.Value)
	}
	return nil
}
func (q *InQueue) Findinuse(n uint) bool {
	for e := q.outer.Next(); e!=q.free.Self(); e = e.Next() {
		if n==0 { break } /* EOL */
		if e.Value!=nil { return true }
	}
	return false
}

func (q *InQueue) ShiftO2I() (i uint) {
	F,O := q.free.Self(),q.outer.Self()
	for {
		e := O.Next()
		if e==F { break } /* Queue empty. */
		if e.Value == nil { break } /* Gab in queue */
		O.InsertBefore(e.Unlink())
		i++
	}
	return
}
func (q *InQueue) LenI() (i uint) {
	I,O := q.inner.Self(),q.outer.Self()
	for e := I.Next(); e!=O ; e = e.Next() { i++ }
	return
}
func (q *InQueue) ShiftIn(queue chan <- interface{}) {
	O,I := q.outer.Self(),q.inner.Self()
	for {
		e := I.Next()
		if e==O { break }
		select {
		case queue <- e.Value:
			e.Value = nil
			I.InsertBefore(e.Unlink())
			continue
		default:
		}
		break
	}
}


type TxQueue struct{
	use,free Ring
}
func (q *TxQueue) Init() {
	q.use.InsertAfter(q.free.Self())
}
func (q *TxQueue) Clear() {
	/*
	Before ...<->U<->Elements...<->F<->Free...
	After ...<->U<->F<->Elements...<->Free...
	*/
	q.use.InsertAfter(q.free.Unlink())
}
func (q *TxQueue) PushBack(v interface{}) {
	U,F := q.use.Self(),q.free.Self()
	e := F.Next()
	if e==U { e = new(Ring) } else { e.Unlink() }
	e.Value = v
	F.InsertBefore(e)
}
func (q *TxQueue) Len() (i uint) {
	U,F := q.use.Self(),q.free.Self()
	for e := U.Next(); e!=F ; e = e.Next() { i++ }
	return
}
func (q *TxQueue) Get(i uint) interface{} {
	U,F := q.use.Self(),q.free.Self()
	for e := U.Next(); e!=F ; e = e.Next() {
		if i!=0 {
			i--
			continue
		}
		return e.Value
	}
	return nil
}
func (q *TxQueue) Shift(i uint) {
	F := q.free.Self()
	for {
		if i==0 { break }
		i--
		e := q.use.Next()
		if e==F { break }
		e.Value = nil
		q.use.InsertBefore(e.Unlink())
	}
}


