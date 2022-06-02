package main

import (
	"time"
	"hash/fnv"
	"math/rand"
)

const names = `abedableablyabutacctacedacesacheachyacidacmeacneacreactsadamaeonafarafroagaragesaginahemahoyaideaimsairyajarakinalasalfaallyalmaaloealpsalsoamidamokampsanewankhannaanonantiantsapexaquaarcharesaridarksarmyartsarty`

func init() {
	rand.Seed(time.Now().Unix())
}

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

func new_hash(input string) hash {
	return hash {
		uint32_from_string(input),
		input,
	}
}

func uint32_from_string(input string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(input))
	return hash.Sum32()
}

func new_name() hash {
	o := rand.Intn(20) * 4
	return new_hash(names[o:o + 4])
}