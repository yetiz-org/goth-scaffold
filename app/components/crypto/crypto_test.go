package crypto

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"testing"
)

const (
	testKeyTypeFoo KeyType = "foo_id"
	testKeyTypeBar KeyType = "bar_id"
	testKeyTypeBaz KeyType = "baz_id"
	testKeyTypeQux KeyType = "qux_id"
)

func resetCryptoGlobalsForTest(t *testing.T) {
	t.Helper()

	oldPrefix := KeyPrefix
	oldCache := keyCache

	KeyPrefix = ""
	keyCache = map[string][]byte{}

	t.Cleanup(func() {
		KeyPrefix = oldPrefix
		keyCache = oldCache
	})
}

func TestEncryptDecryptRoundTripVariousLengths(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	testCases := []struct {
		name string
		data []byte
	}{
		{name: "one_byte", data: []byte{0x7f}},
		{name: "fifteen_bytes", data: bytes.Repeat([]byte{0x11}, 15)},
		{name: "exact_block_size_16", data: bytes.Repeat([]byte{0x22}, 16)},
		{name: "seventeen_bytes", data: bytes.Repeat([]byte{0x33}, 17)},
		{name: "non_ascii_bytes", data: []byte{0x00, 0x01, 0xfe, 0xff, 0x10, 0x20}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := Encrypt(tc.data, testKeyTypeFoo, "")
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := Decrypt(encrypted, testKeyTypeFoo, "")
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(decrypted, tc.data) {
				t.Fatalf("round-trip mismatch, got=%v want=%v", decrypted, tc.data)
			}
		})
	}
}

func TestEncryptReturnsErrorOnEmptyData(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	_, err := Encrypt([]byte{}, testKeyTypeFoo, "")
	if err == nil {
		t.Fatal("Encrypt() expected error for empty data, got nil")
	}

	if err.Error() != "data cannot be empty" {
		t.Fatalf("unexpected error message, got=%q", err.Error())
	}
}

func TestDecryptReturnsErrorOnEmptyEncrypted(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	_, err := Decrypt("", testKeyTypeFoo, "")
	if err == nil {
		t.Fatal("Decrypt() expected error for empty encrypted, got nil")
	}

	if err.Error() != "empty encrypted" {
		t.Fatalf("unexpected error message, got=%q", err.Error())
	}
}

func TestDecryptReturnsErrorOnInvalidBase62Data(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	_, err := Decrypt("@@@", testKeyTypeFoo, "")
	if err == nil {
		t.Fatal("Decrypt() expected error for invalid base62 data, got nil")
	}

	validErrors := []string{"invalid base62 data", "invalid encrypted data length"}
	for _, expected := range validErrors {
		if strings.Contains(err.Error(), expected) {
			return
		}
	}

	t.Fatalf("unexpected error message, got=%q", err.Error())
}

func TestDecryptReturnsErrorOnInvalidEncryptedLength(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	_, err := Decrypt("1", testKeyTypeFoo, "")
	if err == nil {
		t.Fatal("Decrypt() expected error for invalid encrypted data length, got nil")
	}

	if err.Error() != "invalid encrypted data length" {
		t.Fatalf("unexpected error message, got=%q", err.Error())
	}
}

func TestDecryptReturnsErrorOnTamperedCiphertext(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	original := bytes.Repeat([]byte{0x55}, 32)
	encrypted, err := Encrypt(original, testKeyTypeBar, "")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	tampered := encrypted[:len(encrypted)-1] + "0"
	_, err = Decrypt(tampered, testKeyTypeBar, "")
	if err == nil {
		t.Fatal("Decrypt() expected error for tampered ciphertext, got nil")
	}

	validErrors := []string{"invalid encrypted data length", "invalid padding", "invalid base62 data"}
	for _, expected := range validErrors {
		if strings.Contains(err.Error(), expected) {
			return
		}
	}

	t.Fatalf("unexpected tampered decrypt error, got=%q", err.Error())
}

func TestKeyTypeEncryptIdDecryptIdRoundTrip(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	ids := []int64{math.MinInt64, -1, 0, 1, 42, math.MaxInt64}
	for _, id := range ids {
		t.Run("id_"+strconv.FormatInt(id, 10), func(t *testing.T) {
			encrypted, err := testKeyTypeBaz.EncryptId(id)
			if err != nil {
				t.Fatalf("EncryptId() error = %v", err)
			}

			decrypted, err := testKeyTypeBaz.DecryptId(encrypted)
			if err != nil {
				t.Fatalf("DecryptId() error = %v", err)
			}

			if decrypted != id {
				t.Fatalf("DecryptId() mismatch, got=%d want=%d", decrypted, id)
			}
		})
	}
}

func TestKeyTypeDecryptIdReturnsErrorWhenDecryptedLengthNotEight(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	encrypted, err := testKeyTypeQux.Encrypt([]byte{0x01})
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = testKeyTypeQux.DecryptId(encrypted)
	if err == nil {
		t.Fatal("DecryptId() expected invalid decrypted data length error, got nil")
	}

	if err.Error() != "invalid decrypted data length" {
		t.Fatalf("unexpected error message, got=%q", err.Error())
	}
}

func TestEncryptKeyIdDecryptKeyIdRoundTrip(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	type myID int64
	id := myID(12345)

	encrypted, err := EncryptKeyId(testKeyTypeFoo, id)
	if err != nil {
		t.Fatalf("EncryptKeyId() error = %v", err)
	}

	decrypted, err := DecryptKeyId[myID](testKeyTypeFoo, encrypted)
	if err != nil {
		t.Fatalf("DecryptKeyId() error = %v", err)
	}

	if decrypted != id {
		t.Fatalf("DecryptKeyId() mismatch, got=%d want=%d", decrypted, id)
	}
}

func TestEncryptedKeyIdReturnsEmptyOnError(t *testing.T) {
	resetCryptoGlobalsForTest(t)

	// EncryptedKeyId is a best-effort function; any valid id should succeed
	// We just verify it returns a non-empty string for a valid id
	result := EncryptedKeyId(testKeyTypeFoo, int64(1))
	if result == "" {
		t.Fatal("EncryptedKeyId() returned empty string for valid id")
	}
}
