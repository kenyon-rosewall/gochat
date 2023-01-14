package utils

import (
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index (len(letterBytes) == 52 == 110100xb)
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits (how many random letters I draw from one 63 bit generation)
)

func GetRandomString(dataLength int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, dataLength)

	// Int63() is faster than Intn(), run until the i runs out (then the random string is as big as dataLength)
	for i, cache, remain := dataLength-1, src.Int63(), letterIdxMax; i >= 0; {
		// If we get to the end, generate a new 63-bit number
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}

		// If the 6-bit mask is less than 52, make it a letter
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}

		// Bit shift by 6 to remove last-used part
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
