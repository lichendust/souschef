/*
	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import "hash/fnv"

type hash struct {
	uint32
	word string
}

func (h hash) String() string {
	return h.word
}

func (a *hash) UnmarshalText(text []byte) error {
	str := string(text)
	a.uint32 = uint32_from_string(str)
	a.word   = str
	return nil
}

func (a hash) MarshalText() ([]byte, error) {
	return []byte(a.word), nil
}

func uint32_from_string(input string) uint32 {
	if input == "" {
		return 0
	}
	hash := fnv.New32a()
	hash.Write([]byte(input))
	return hash.Sum32()
}

func new_hash(input string) hash {
	return hash {
		uint32_from_string(input),
		input,
	}
}