package util

import (
	"math/rand"
	"time"
	"unsafe"
)

func RandomString(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	const str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var letterIdxBits int64 = 6
	var letterIdxMask int64 = 1<<letterIdxBits - 1
	letterIdxMax := 63 / letterIdxBits

	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(str) {
			b[i] = str[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}
