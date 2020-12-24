package utils

/*
	utils function to gen a random string
	https://cloud.tencent.com/developer/article/1580647
	https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
*/

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandStringBytesMaskImprSrcUnsafe gen random string
// NOTE: not thread safe!!!
func RandStringBytesMaskImprSrcUnsafe(n int) string {
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

var rr = rand.New(rand.NewSource(time.Now().UnixNano()))
var randLock sync.Mutex

// RandStrSafe return rand string, thread safe
func RandStrSafe(n int) string {
	randLock.Lock()
	defer randLock.Unlock()
	return RandStringBytesMaskImprSrcUnsafe(n)
}

// RandInt31Safe return rand integer, thread safe
func RandInt31Safe() int {
	randLock.Lock()
	defer randLock.Unlock()
	return int(rr.Int31())
}

// RandInt31nSafe return rand integer, thread safe
func RandInt31nSafe(n int32) int {
	randLock.Lock()
	defer randLock.Unlock()
	return int(rr.Int31n(n))
}
