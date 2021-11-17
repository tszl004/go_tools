package tools

import (
	cRand "crypto/rand"
	"encoding/binary"
	mRand "math/rand"
	"time"
)

type random struct {
	str  string
	seed int64
}

func (r *random) randStr(length int) string {
	_ = binary.Read(cRand.Reader, binary.BigEndian, &r.seed)
	mRand.Seed(r.seed)
	max := len(r.str)
	var str string
	for i := 0; i < length; i++ {
		rn := mRand.Intn(max)
		str += r.str[rn : rn+1]
	}
	return str
}

func Random(length int, randStr ...string) string {
	var r random
	if len(randStr) == 0 {
		r = random{"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", time.Now().Unix()}
	} else {
		r = random{randStr[0], time.Now().Unix()}
	}
	return r.randStr(length)
}

func Num(length int) string {
	return Random(length, "0123456789")
}

func Letter(length int) string {
	return Random(length, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}
