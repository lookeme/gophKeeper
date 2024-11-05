package security

import (
	db "GophKeeper/internal/storage"
	"GophKeeper/utils"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"sync"
	"time"
)

const (
	// byte12nonce is the size of the nonce used in AES-GCM encryption, set to 12 bytes.
	byte12nonce int = 12
	// rotationTime defines the interval at which keys will be rotated, set to 5 minutes. (should be out to config)
	rotationTime time.Duration = 5 * time.Minute
)

type SecureService struct {
	storage *db.Storage
	logger  *zap.Logger
	kek     []byte
	dek     []byte
	mu      sync.Mutex
}

func NewSecureService(storage *db.Storage, logger *zap.Logger) *SecureService {
	return &SecureService{
		storage: storage,
		logger:  logger,
		kek:     []byte{},
		dek:     []byte{},
		mu:      sync.Mutex{},
	}
}

// generateKEK создает новый KEK
func generateKey() []byte {
	key := make([]byte, 32) // 256-битный ключ
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalf("Failed to generate KEK: %v", err)
	}
	return key
}

func (s *SecureService) GetKek(ctx context.Context) (string, error) {
	key, err := s.storage.SettingsRepository.FindSettingsByKey(ctx, utils.SettingKeyKek)
	if err != nil {
		return "", err
	}
	return key.Value, nil
}

func (s *SecureService) GetDek(ctx context.Context) (string, error) {
	key, err := s.storage.SettingsRepository.FindSettingsByKey(ctx, utils.SettingKeyDek)
	if err != nil {
		return "", err
	}
	return key.Value, nil
}

func (s *SecureService) getAllKeys(ctx context.Context) ([]byte, []byte, error) {
	kek, err := s.GetKek(ctx)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	dek, err := s.GetDek(ctx)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	return []byte(kek), []byte(dek), nil
}

func (s *SecureService) Init(ctx context.Context) error {
	kek, dek, keyErr := s.getAllKeys(ctx)
	if keyErr == nil {
		decodedKek, decErr := base64.StdEncoding.DecodeString(string(kek))
		if decErr != nil {
			return decErr
		}
		decodedDek, decErr := base64.StdEncoding.DecodeString(string(dek))
		if decErr != nil {
			return decErr
		}
		s.kek = decodedKek
		s.dek = decodedDek
		s.logger.Info("dek and kek loaded")
		return nil
	}
	kek = generateKey()
	dek = generateKey()
	encryptedDEK, err := s.encryptDEK(kek, dek)
	if err != nil {
		return err
	}

	err = s.storage.SettingsRepository.SaveKeys(ctx,
		base64.StdEncoding.EncodeToString(kek),
		base64.StdEncoding.EncodeToString(dek))
	if err != nil {
		return err
	}
	s.kek = kek
	s.dek = encryptedDEK
	return nil
}

// EncryptData2 encrypts plaintext using DEK
func (s *SecureService) EncryptData2(plaintext []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	decryptedDEK, err := s.decryptDEK(s.kek, s.dek)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(decryptedDEK)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, byte12nonce) // AES-GCM nonce size
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// EncryptData encrypts plaintext using DEK
func (s *SecureService) EncryptData(plaintext []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	decryptedDEK, err := s.decryptDEK(s.kek, s.dek)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(decryptedDEK)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, byte12nonce) // AES-GCM nonce size
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptData decrypts the base64 encoded ciphertext using DEK
func (s *SecureService) DecryptData(encrypted string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}
	decryptedDEK, err := s.decryptDEK(s.kek, s.dek)
	if err != nil {
		return []byte(""), err
	}
	block, err := aes.NewCipher(decryptedDEK)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < byte12nonce {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:byte12nonce]
	ciphertext = ciphertext[byte12nonce:]

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// encryptDEK шифрует DEK с использованием текущего KEK
func (s *SecureService) encryptDEK(kek []byte, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(dek))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], dek)
	return ciphertext, nil
}

// decryptDEK дешифрует DEK с использованием текущего KEK
func (s *SecureService) decryptDEK(kek []byte, encryptedDEK []byte) ([]byte, error) {
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}
	iv := encryptedDEK[:aes.BlockSize]
	encryptedDEK = encryptedDEK[aes.BlockSize:]
	dek := make([]byte, len(encryptedDEK))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(dek, encryptedDEK)
	return dek, nil
}

func (s *SecureService) StartTickerRotation(ctx context.Context) {
	ticker := time.NewTicker(rotationTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.rotateKeys(ctx)
		case <-ctx.Done():
			break
		}
	}
}

func (s *SecureService) rotateKeys(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newKey := generateKey()
	dek, _ := s.decryptDEK(s.kek, s.dek)
	newEncryptedDEK, _ := s.encryptDEK(newKey, dek)
	err := s.storage.SettingsRepository.SaveKeys(ctx,
		base64.StdEncoding.EncodeToString(newKey),
		base64.StdEncoding.EncodeToString(newEncryptedDEK))
	if err != nil {
		log.Fatalf("Failed to update keys: %v", err)
	}
	s.kek = newKey
	s.dek = newEncryptedDEK
	s.logger.Info("keys rotated")
}
func getSecretKeyToken() string {
	return "secret-key"
}

func EncodePass(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
}

func CheckPass(pass string, hashedPass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass))
	return err == nil
}

type EncryptBuffer struct {
	buffer  *bytes.Buffer
	encrypt func([]byte) (string, error)
	decrypt func(text string) ([]byte, error)
}

func (se *EncryptBuffer) Read(p []byte) (n int, err error) {
	return se.buffer.Read(p)
}

func NewEncryptBuffer(size int, encrypt func([]byte) (string, error), decrypt func(text string) ([]byte, error)) *EncryptBuffer {
	buffer := bytes.NewBuffer(nil)
	if size > 0 {
		buffer = bytes.NewBuffer(make([]byte, size))
	}
	return &EncryptBuffer{
		buffer:  buffer,
		encrypt: encrypt,
		decrypt: decrypt,
	}
}
func (se *EncryptBuffer) Write(p []byte) (int, error) {
	return se.buffer.Write(p)
}

func (se *EncryptBuffer) Encrypt() error {
	encrypted, err := se.encrypt(se.buffer.Bytes())
	if err != nil {
		return err
	}
	se.buffer.Reset()
	se.buffer.Write([]byte(encrypted))
	return err
}

func (se *EncryptBuffer) Decrypt() error {
	decrypted, err := se.decrypt(se.buffer.String())
	if err != nil {
		return err
	}
	se.buffer.Reset()
	se.buffer.Write(decrypted)
	return err
}
func (se *EncryptBuffer) Buffer() *bytes.Buffer {
	return se.buffer
}

func (se *EncryptBuffer) Len() int {
	return se.buffer.Len()
}

func (se *EncryptBuffer) Reset() {
	se.buffer.Reset()
}
