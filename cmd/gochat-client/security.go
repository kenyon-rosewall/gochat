package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

func getGCM(key []byte) cipher.AEAD {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Could not create new AES cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Could not creat new AES GCM from cipher block")
	}

	return aesgcm
}

func Encrypt(msg string, key string) (string, string, string) {
	ckey := []byte(key)
	cmsg := []byte(msg)

	aesgcm := getGCM(ckey)
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Could not create random nonce")
	}
	ciphermsg := aesgcm.Seal(nil, nonce, cmsg, nil)

	mac := hmac.New(sha256.New, ckey)
	mac.Write(cmsg)
	macsum := mac.Sum(nil)

	return string(ciphermsg[:]), string(nonce[:]), string(macsum[:])
}

func Decrypt(ciphermsg string, nonce string, macsum string, key string) string {
	ckey := []byte(key)
	cciphermsg := []byte(ciphermsg)
	cnonce := []byte(nonce)
	cmacsum := []byte(macsum)
	msg := ""

	mac := hmac.New(sha256.New, ckey)
	mac.Write(cciphermsg)
	rmacsum := mac.Sum(nil)

	if hmac.Equal(cmacsum, rmacsum) {
		aesgcm := getGCM(ckey)
		cmsg, err := aesgcm.Open(nil, cnonce, cciphermsg, nil)
		if err != nil {
			fmt.Println("Could not decrypt message with nonce")
		}
		msg = string(cmsg[:])
	} else {
		fmt.Println("Message received was not authentic")
	}

	return msg
}
