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


type Ring struct {
	Value interface{}
	prev,next *Ring
}
func (r *Ring) Self() *Ring { return r }
func (r *Ring) init() {
	r.next = r
	r.prev = r
}
func (r *Ring) Prev() *Ring { return r.prev }
func (r *Ring) Next() *Ring { return r.next }
func (r *Ring) Unlink() *Ring {
	p,n := r.prev,r.next
	p.next = n
	n.prev = p
	r.prev = nil
	r.next = nil
	return r
}
/*
from a<->c
to a<->b<->c
*/
func (a *Ring) insert(c, b *Ring) {
	/* a->b->c */
	/* a->b */
	a.next = b
	/* b->c */
	b.next = c
	
	/* a<-b<-c */
	/* b<-c */
	c.prev = b
	/* a<-b */
	b.prev = a
}
func (a *Ring) InsertAfter(b *Ring) *Ring {
	if a.next==nil { a.init() }
	a.insert(a.next,b)
	return b
}
func (c *Ring) InsertBefore(b *Ring) *Ring {
	if c.prev==nil { c.init() }
	c.prev.insert(c,b)
	return b
}

