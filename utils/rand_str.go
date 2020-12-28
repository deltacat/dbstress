package utils

/*
	utils function to gen a random string
	https://cloud.tencent.com/developer/article/1580647
	https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
*/

import (
	"math/rand"
	"time"
	"unsafe"
)

// StrDataLength string data length
const StrDataLength int = 64

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())
var testStrs []string
var testStrNum int32 = 100000
var testStrIdx int32 = 0

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandStringBytesMaskImprSrcUnsafe gen random string
// NOTE: not thread safe!!!
func randStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// RandStrSafe return rand string
// to save generating time while run test, init a random string before run test.
// here only pick one of them from predefined strings.
// idx is not thread safe but could be ignore
func RandStrSafe(n int) (result string) {
	result = testStrs[testStrIdx]
	idx := testStrIdx
	idx++
	if idx >= testStrNum {
		idx = 0
	}
	testStrIdx = idx
	return
}

func init() {
	testStrs = make([]string, testStrNum)
	for i := range testStrs {
		testStrs[i] = randStringBytesMaskImprSrcUnsafe(StrDataLength)
	}
}
