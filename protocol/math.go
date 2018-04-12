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


const mi32 = 0x7fffffff

/* base < it <= end */
func isrange(base, end, it uint32) bool {
	E,I := end-base, it-base
	return (I<=E) && (I<=mi32) && (it!=base)
}
/* base <= it <= end */
func isrange2(base, end, it uint32) bool {
	E,I := end-base, it-base
	return (I<=E) && (I<=mi32)
}

func relative(base, it uint32) uint {
	return uint(it-base)-1
}

func window(maxbuf, bufsiz, maxwin uint) uint {
	i := maxbuf-bufsiz
	if i>mi32 { i = 0 }
	if i>maxwin { i = maxwin }
	return i
}

func fchk(field uint16, flag uint16) bool {
	return (field&flag)!=0
}

func islower(a,b uint32) bool {
	return (a-b)>mi32
}
