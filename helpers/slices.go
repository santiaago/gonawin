/*
* Copyright (c) 2014 Santiago Arias | Remy Jourde
*
* Permission to use, copy, modify, and distribute this software for any
* purpose with or without fee is hereby granted, provided that the above
* copyright notice and this permission notice appear in all copies.
*
* THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
* WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
* MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
* ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
* WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
* ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
* OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package helpers

// Insert inserts new value into a slice at a given position
// source: https://code.google.com/p/go-wiki/wiki/SliceTricks
func Insert(s []int64, value int64, i int64) []int64 {
	s = append(s, 0)
	copy(s[i+1:], s[i:])
	s[i] = value
	return s
}

// Contains indicates if a value exists in a given slice.
// If the value exists, its position in the slice is returned otherwise -1.
func Contains(s []int64, value int64) (bool, int) {
	for i, v := range s {
		if v == value {
			return true, i
		}
	}
	return false, -1
}
