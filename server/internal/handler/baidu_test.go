package handler

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"testing"
)

func TestBaiduSHA1(t *testing.T) {
	got := baiduSHA1("token", "123", "nonce")
	if got != baiduSHA1("token", "nonce", "123") {
		t.Fatalf("signature should be order independent")
	}
}

func TestDecryptBaijiahaoMessage(t *testing.T) {
	appID := "app-123"
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	key = strings.TrimRight(key, "=")
	message := "hello baijiahao"

	encrypted := encryptBaijiahaoMessageForTest(t, appID, key, message)
	got, err := decryptBaijiahaoMessage(appID, key, encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if got != message {
		t.Fatalf("unexpected message: %q", got)
	}
}

func encryptBaijiahaoMessageForTest(t *testing.T, appID string, encodingAESKey string, message string) string {
	t.Helper()

	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		t.Fatal(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	body := append(bytes.Repeat([]byte("r"), 16), uint32Bytes(uint32(len(message)))...)
	body = append(body, []byte(message)...)
	body = append(body, []byte(appID)...)
	body = baiduPKCS7PadForTest(body)

	ciphertext := make([]byte, len(body))
	cipher.NewCBCEncrypter(block, key[:aes.BlockSize]).CryptBlocks(ciphertext, body)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func uint32Bytes(value uint32) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, value)
	return buffer
}

func baiduPKCS7PadForTest(value []byte) []byte {
	padding := 32 - len(value)%32
	if padding == 0 {
		padding = 32
	}
	return append(value, bytes.Repeat([]byte{byte(padding)}, padding)...)
}
