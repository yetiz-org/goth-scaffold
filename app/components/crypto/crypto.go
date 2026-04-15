package crypto

import (
	"crypto/aes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	base62 "github.com/yetiz-org/goth-base62"
)

type KeyId uint64

func (k KeyId) UInt64() uint64 {
	return uint64(k)
}

type KeyType string

// EncryptKeyId encrypts a key ID of any type whose underlying type is uint64.
func EncryptKeyId[T ~uint64](keyType KeyType, id T) (string, error) {
	return keyType.EncryptId(uint64(id))
}

// EncryptedKeyId encrypts a key ID, returning an empty string on error.
func EncryptedKeyId[T ~uint64](keyType KeyType, id T) string {
	encrypted, err := keyType.EncryptId(uint64(id))
	if err != nil {
		return ""
	}

	return encrypted
}

// DecryptKeyId decrypts an encrypted ID string into the target key ID type T.
func DecryptKeyId[T ~uint64](keyType KeyType, encryptedId string) (T, error) {
	id, err := keyType.DecryptId(encryptedId)
	if err != nil {
		return 0, err
	}

	return T(id), nil
}

const DefaultVersion = "1.0"

var KeyPrefix = ""

var (
	keyCache   = map[string][]byte{}
	keyCacheMu sync.RWMutex
)

func getKey(version string, keyType KeyType) []byte {
	cacheKey := fmt.Sprintf("%s_%s:%s", KeyPrefix, version, keyType.String())
	keyCacheMu.RLock()
	cachedKey, ok := keyCache[cacheKey]
	keyCacheMu.RUnlock()
	if ok {
		return cachedKey
	}

	sum256 := sha256.Sum256([]byte(cacheKey))
	key := make([]byte, 16)
	copy(key, sum256[:16])

	keyCacheMu.Lock()
	if cachedKey, ok := keyCache[cacheKey]; ok {
		keyCacheMu.Unlock()
		return cachedKey
	}
	keyCache[cacheKey] = key
	keyCacheMu.Unlock()

	return key
}

func (k KeyType) String() string {
	return string(k)
}

func (k KeyType) Encrypt(data []byte) (string, error) {
	return Encrypt(data, k, "")
}

func (k KeyType) EncryptId(id uint64) (string, error) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, id)
	return k.Encrypt(data)
}

func (k KeyType) Decrypt(encryptedId string) ([]byte, error) {
	return Decrypt(encryptedId, k, "")
}

func (k KeyType) DecryptId(encryptedId string) (uint64, error) {
	decryptedData, err := k.Decrypt(encryptedId)
	if err != nil {
		return 0, err
	}

	if len(decryptedData) != 8 {
		return 0, errors.New("invalid decrypted data length")
	}

	idValue := binary.BigEndian.Uint64(decryptedData)

	return idValue, nil
}

// Encrypt encrypts data using AES-ECB with PKCS7 padding and base62 encoding.
func Encrypt(data []byte, keyType KeyType, version string) (string, error) {
	if len(data) == 0 {
		return "", errors.New("data cannot be empty")
	}

	if version == "" {
		version = DefaultVersion
	}

	// Apply PKCS7 padding
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	paddedData := append(data, padtext...)

	key := getKey(version, keyType)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// ECB mode - encrypt all blocks
	ciphertext := make([]byte, len(paddedData))
	for i := 0; i < len(paddedData); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], paddedData[i:i+aes.BlockSize])
	}

	return base62.StdEncoding.EncodeToString(ciphertext), nil
}

func decodeBase62Safe(input string) (decoded []byte, err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			decoded = nil
			err = errors.New("invalid base62 data")
		}
	}()

	decoded = base62.StdEncoding.DecodeString(input)
	return decoded, nil
}

// Decrypt decrypts an AES-ECB base62-encoded ciphertext.
func Decrypt(encrypted string, keyType KeyType, version string) ([]byte, error) {
	if encrypted == "" {
		return nil, errors.New("empty encrypted")
	}

	if version == "" {
		version = DefaultVersion
	}

	encryptedData, err := decodeBase62Safe(encrypted)
	if err != nil {
		return nil, err
	}

	if len(encryptedData)%aes.BlockSize != 0 {
		return nil, errors.New("invalid encrypted data length")
	}

	key := getKey(version, keyType)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// ECB mode - decrypt all blocks
	decrypted := make([]byte, len(encryptedData))
	for i := 0; i < len(encryptedData); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], encryptedData[i:i+aes.BlockSize])
	}

	if len(decrypted) == 0 {
		return nil, errors.New("invalid decrypted data")
	}

	padding := int(decrypted[len(decrypted)-1])
	if padding > aes.BlockSize || padding == 0 || padding > len(decrypted) {
		return nil, errors.New("invalid padding")
	}

	// Validate padding bytes
	for i := len(decrypted) - padding; i < len(decrypted); i++ {
		if decrypted[i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}

	return decrypted[:len(decrypted)-padding], nil
}
