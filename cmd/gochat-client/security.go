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
	mac.Write(ciphermsg)
	macsum := mac.Sum(nil)

	return string(ciphermsg[:]), string(nonce[:]), string(macsum[:])
}

func Decrypt(ciphermsg []byte, nonce []byte, macsum []byte, key string) string {
	ckey := []byte(key)
	msg := ""

	mac := hmac.New(sha256.New, ckey)
	mac.Write(ciphermsg)
	rmacsum := mac.Sum(nil)

	if hmac.Equal(macsum, rmacsum) {
		aesgcm := getGCM(ckey)
		cmsg, err := aesgcm.Open(nil, nonce, ciphermsg, nil)
		if err != nil {
			fmt.Println("Could not decrypt message with nonce")
		}
		msg = string(cmsg[:])
	} else {
		fmt.Println("Message received was not authentic")
		fmt.Println(macsum)
		fmt.Println(rmacsum)
	}

	return msg
}
