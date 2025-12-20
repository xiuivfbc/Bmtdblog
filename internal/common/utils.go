package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/snluu/uuid"
	"github.com/xiuivfbc/bmtdblog/internal/system"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

// Md5 计算字符串的md5值（用于签名等非密码用途）
func Md5(source string) string {
	md5h := md5.New()
	md5h.Write([]byte(source))
	return hex.EncodeToString(md5h.Sum(nil))
}

// HashPassword 使用bcrypt生成密码哈希
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword 验证密码是否匹配哈希值
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func Truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

func Len(s string) int {
	return len([]rune(s))
}

func UUID() string {
	return uuid.Rand().Hex()
}

func GetCurrentTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Chongqing")
	return time.Now().In(loc)
}

func GetCurrentDirectory() string {
	dir := system.GetConfiguration().Dir
	return dir
}

// SendMail 发送邮件（简化版本，用于邮件队列）
func SendMail(to, subject, body string) error {
	cfg := system.GetConfiguration()
	if !cfg.Smtp.Enabled {
		return nil
	}

	return SendToMail(cfg.Smtp.Username, cfg.Smtp.Password,
		cfg.Smtp.Host, // Host中包含端口信息
		to, subject, body, "html")
}

func SendToMail(user, password, host, to, subject, body, mailType string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", user)
	m.SetHeader("To", strings.Split(to, ";")...)
	m.SetHeader("Subject", subject)
	if mailType == "html" {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	hp := strings.Split(host, ":")
	port := 25
	if len(hp) > 1 {
		port, _ = strconv.Atoi(hp[1])
	}
	d := gomail.NewDialer(hp[0], port, user, password)
	return d.DialAndSend(m)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, enc := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func Encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}
