package transfer

import (
	"encoding/binary"
	"fmt"
	"net"
)

const HeaderLength = 18

type TransferType int

const (
	Quit TransferType = iota
	Message
	Command
)

type Format uint16

const (
	Macsum Format = iota
	Nonce
	Body
)

type Transfer struct {
	SessionID    uint64
	MessageType  TransferType
	Body         string
	FullTransfer []byte
}

func createHeader(sessionID uint64, objectType uint16, data string) []byte {
	header := make([]byte, HeaderLength)
	binary.BigEndian.PutUint64(header[:8], sessionID)
	binary.BigEndian.PutUint16(header[8:10], objectType)
	binary.BigEndian.PutUint64(header[10:], uint64(len(data)))

	return header
}

func unpackPacket(t []byte, idx int) (uint64, uint16, []byte, int) {
	headerStart := idx
	left := idx + 8
	right := idx + 10
	headerEnd := idx + HeaderLength

	sessionID := binary.BigEndian.Uint64(t[headerStart:left])
	dataType := binary.BigEndian.Uint16(t[left:right])
	dataLength := binary.BigEndian.Uint64(t[right:headerEnd])
	dataEnd := headerEnd + int(dataLength)
	var data []byte

	if idx > 0 {
		data = t[headerEnd:dataEnd]
	}

	return sessionID, dataType, data, dataEnd
}

// TODO: Rewrite this with the length of the Transfer found in the TransferHeader
func getFullTransfer(conn net.Conn) ([]byte, error) {
	var fullTransfer []byte

	header := make([]byte, HeaderLength)
	_, err := conn.Read(header)
	if err != nil {
		return nil, err
	}
	fullTransfer = append(fullTransfer, header...)

	for i := 0; i < 3; i++ {
		h := make([]byte, HeaderLength)
		_, err := conn.Read(h)
		if err != nil {
			return nil, err
		}

		dataLength := binary.BigEndian.Uint64(h[10:])
		d := make([]byte, dataLength)
		_, err = conn.Read(d)
		if err != nil {
			return nil, err
		}

		h = append(h, d...)
		fullTransfer = append(fullTransfer, h...)
	}

	return fullTransfer, nil
}

func PackTransfer(sessionID uint64, transferType TransferType, msg string, key string) []byte {
	transfer := createHeader(sessionID, uint16(transferType), "")

	if transferType == Message {
		cipher, nonce, macsum := Encrypt(msg, key)

		cipherHeader := createHeader(sessionID, uint16(Body), cipher)
		nonceHeader := createHeader(sessionID, uint16(Nonce), nonce)
		macsumHeader := createHeader(sessionID, uint16(Macsum), macsum)

		cipherPacket := append(cipherHeader, cipher...)
		noncePacket := append(nonceHeader, nonce...)
		macsumPacket := append(macsumHeader, macsum...)

		transfer = append(transfer, cipherPacket...)
		transfer = append(transfer, noncePacket...)
		transfer = append(transfer, macsumPacket...)
	}

	// TODO: Update the TransferHeader to include the length of the entire Transfer

	return transfer
}

func UnpackTransfer(conn net.Conn, key string) *Transfer {
	t, err := getFullTransfer(conn)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	sessionID, transferType, _, packetEnd := unpackPacket(t, 0)
	body := ""

	if sessionID > 0 && key != "" && TransferType(transferType) != Quit {
		dataVerified := true
		cipherSid, cipherType, cipher, packetEnd := unpackPacket(t, packetEnd)
		if cipherSid != sessionID || Format(cipherType) != Body {
			dataVerified = false
		}
		nonceSid, nonceType, nonce, packetEnd := unpackPacket(t, packetEnd)
		if nonceSid != sessionID || Format(nonceType) != Nonce {
			dataVerified = false
		}
		macsumSid, macsumType, macsum, _ := unpackPacket(t, packetEnd)
		if macsumSid != sessionID || Format(macsumType) != Macsum {
			dataVerified = false
		}

		if !dataVerified {
			return nil
		}

		body = Decrypt(cipher, nonce, macsum, key)
	}

	result := &Transfer{
		SessionID:    sessionID,
		MessageType:  TransferType(transferType),
		Body:         body,
		FullTransfer: t,
	}

	return result
}
