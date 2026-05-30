package handler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strings"

	"wjfcms-go/internal/config"

	"github.com/gin-gonic/gin"
)

type BaiduHandler struct {
	cfg config.Config
}

func NewBaiduHandler(cfg config.Config) *BaiduHandler {
	return &BaiduHandler{cfg: cfg}
}

func (h *BaiduHandler) Serve(c *gin.Context) {
	timestamp := requestValue(c, "timestamp")
	nonce := requestValue(c, "nonce")
	signature := requestValue(c, "signature")
	encrypted := requestValue(c, "encrypt")

	if h.cfg.Baijiahao.AppID == "" || h.cfg.Baijiahao.AppYouToken == "" || h.cfg.Baijiahao.EncodingAESKey == "" {
		c.String(http.StatusBadRequest, "failed")
		return
	}
	if timestamp == "" || nonce == "" || signature == "" || encrypted == "" {
		c.String(http.StatusBadRequest, "failed")
		return
	}
	if baiduSHA1(h.cfg.Baijiahao.AppYouToken, timestamp, nonce) != signature {
		c.String(http.StatusOK, "failed")
		return
	}

	plain, err := decryptBaijiahaoMessage(h.cfg.Baijiahao.AppID, h.cfg.Baijiahao.EncodingAESKey, encrypted)
	if err != nil {
		c.String(http.StatusBadRequest, "failed")
		return
	}
	c.String(http.StatusOK, plain)
}

func requestValue(c *gin.Context, key string) string {
	if value := strings.TrimSpace(c.Query(key)); value != "" {
		return value
	}
	return strings.TrimSpace(c.PostForm(key))
}

func baiduSHA1(token string, timestamp string, nonce string) string {
	values := []string{token, timestamp, nonce}
	sort.Strings(values)
	sum := sha1.Sum([]byte(strings.Join(values, "")))
	return hex.EncodeToString(sum[:])
}

func decryptBaijiahaoMessage(appID string, encodingAESKey string, encrypted string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	if err != nil {
		return "", err
	}
	if len(key) != 32 {
		return "", errors.New("invalid aes key length")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("invalid ciphertext length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := key[:aes.BlockSize]
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)

	plain, err = baiduPKCS7Unpad(plain)
	if err != nil {
		return "", err
	}
	if len(plain) < 20 {
		return "", errors.New("invalid plaintext length")
	}

	content := plain[16:]
	messageLength := int(binary.BigEndian.Uint32(content[:4]))
	if messageLength < 0 || len(content) < 4+messageLength {
		return "", errors.New("invalid message length")
	}

	message := content[4 : 4+messageLength]
	fromAppID := string(content[4+messageLength:])
	if fromAppID != appID {
		return "", errors.New("invalid app id")
	}
	return string(message), nil
}

func baiduPKCS7Unpad(value []byte) ([]byte, error) {
	if len(value) == 0 {
		return nil, errors.New("empty plaintext")
	}
	pad := int(value[len(value)-1])
	if pad < 1 || pad > 32 || pad > len(value) {
		return nil, errors.New("invalid padding")
	}
	for _, b := range value[len(value)-pad:] {
		if int(b) != pad {
			return nil, errors.New("invalid padding")
		}
	}
	return value[:len(value)-pad], nil
}
