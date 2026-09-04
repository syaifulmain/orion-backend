package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const (
	APIKeyPrefix        = "cps_"
	APIKeyRandomLength  = 22
	MinimumSecretLength = 32
	MaximumTTLDays      = 365
	MaximumKeyCount     = 10000
	MaximumKeyStoreSize = 1000000 // 1MB
)

var (
	keyPattern       = regexp.MustCompile(`^cps_[A-Za-z0-9_-]{22}$`)
	keyDigestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
	namePattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

	ErrInvalidSecret    = errors.New("CORPUS_API_SECRET must contain at least 32 bytes")
	ErrInvalidKeyName   = errors.New("key name must be 1-64 characters and use only letters, numbers, dot, underscore, or hyphen")
	ErrInvalidTTL       = fmt.Errorf("expiration must be between 1 and %d days", MaximumTTLDays)
	ErrInvalidAPIKey    = errors.New("Invalid or Expired API Key")
	ErrRegistryLimit    = errors.New("API key registry has reached its limit")
	ErrRegistryRead     = errors.New("API key registry cannot be read safely")
	ErrRegistryWrite    = errors.New("API key registry cannot be written safely")
	ErrRegistryFormat   = errors.New("API key registry has an unsupported format")
)

type Claims struct {
	Name      string `json:"name"`
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}

type KeyRecord struct {
	Digest    string `json:"digest"`
	Name      string `json:"name"`
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}

type KeyStorePayload struct {
	Version int         `json:"version"`
	Keys    []KeyRecord `json:"keys"`
}

func ValidateSecret(secret string) error {
	if len([]byte(secret)) < MinimumSecretLength {
		return ErrInvalidSecret
	}
	return nil
}

func ValidateKeyName(name string) error {
	if !namePattern.MatchString(name) {
		return ErrInvalidKeyName
	}
	return nil
}

func KeyDigest(apiKey, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(apiKey))
	return hex.EncodeToString(mac.Sum(nil))
}

func GenerateAPIKey(secret, name, keyStorePath string, expiresInDays int, customNow ...time.Time) (string, *Claims, error) {
	if err := ValidateSecret(secret); err != nil {
		return "", nil, err
	}
	if err := ValidateKeyName(name); err != nil {
		return "", nil, err
	}
	if expiresInDays < 1 || expiresInDays > MaximumTTLDays {
		return "", nil, ErrInvalidTTL
	}

	now := time.Now()
	if len(customNow) > 0 {
		now = customNow[0]
	}

	issuedAt := now.Unix()
	expiresAt := issuedAt + int64(expiresInDays*24*60*60)

	existingRecords, err := loadKeyStore(keyStorePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", nil, err
	}

	// Filter unexpired records
	validRecords := make([]KeyRecord, 0, len(existingRecords))
	existingDigests := make(map[string]bool)
	for _, rec := range existingRecords {
		if rec.ExpiresAt > issuedAt {
			validRecords = append(validRecords, rec)
			existingDigests[rec.Digest] = true
		}
	}

	if len(validRecords) >= MaximumKeyCount {
		return "", nil, ErrRegistryLimit
	}

	var apiKey, digest string
	for attempt := 0; attempt < 5; attempt++ {
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err != nil {
			return "", nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}
		apiKey = APIKeyPrefix + base64.RawURLEncoding.EncodeToString(randomBytes)
		digest = KeyDigest(apiKey, secret)
		if !existingDigests[digest] {
			break
		}
	}

	validRecords = append(validRecords, KeyRecord{
		Digest:    digest,
		Name:      name,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	})

	if err := writeKeyStore(keyStorePath, validRecords); err != nil {
		return "", nil, err
	}

	claims := &Claims{
		Name:      name,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}

	return apiKey, claims, nil
}

func ValidateAPIKey(apiKey, secret, keyStorePath string, customNow ...time.Time) (*Claims, error) {
	if err := ValidateSecret(secret); err != nil {
		return nil, err
	}
	if !keyPattern.MatchString(apiKey) {
		return nil, ErrInvalidAPIKey
	}

	expectedDigest := KeyDigest(apiKey, secret)

	records, err := loadKeyStore(keyStorePath)
	if err != nil {
		return nil, ErrRegistryRead
	}

	now := time.Now().Unix()
	if len(customNow) > 0 {
		now = customNow[0].Unix()
	}

	for _, rec := range records {
		if hmac.Equal([]byte(rec.Digest), []byte(expectedDigest)) {
			if now >= rec.ExpiresAt {
				return nil, ErrInvalidAPIKey
			}
			return &Claims{
				Name:      rec.Name,
				IssuedAt:  rec.IssuedAt,
				ExpiresAt: rec.ExpiresAt,
			}, nil
		}
	}

	return nil, ErrInvalidAPIKey
}

func loadKeyStore(path string) ([]KeyRecord, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []KeyRecord{}, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrRegistryRead, err)
	}

	if info.Size() > MaximumKeyStoreSize {
		return nil, fmt.Errorf("%w: file exceeds size limit", ErrRegistryRead)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRegistryRead, err)
	}

	var payload KeyStorePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("%w: invalid json: %v", ErrRegistryFormat, err)
	}

	if payload.Version != 1 {
		return nil, ErrRegistryFormat
	}

	for _, rec := range payload.Keys {
		if !keyDigestPattern.MatchString(rec.Digest) || !namePattern.MatchString(rec.Name) || rec.ExpiresAt <= rec.IssuedAt {
			return nil, ErrRegistryFormat
		}
	}

	return payload.Keys, nil
}

func writeKeyStore(path string, records []KeyRecord) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %v", ErrRegistryWrite, err)
	}

	payload := KeyStorePayload{
		Version: 1,
		Keys:    records,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: marshal error: %v", ErrRegistryWrite, err)
	}
	data = append(data, '\n')

	// Atomic write via temporary file
	tempFile, err := os.CreateTemp(dir, ".api_keys.*.tmp")
	if err != nil {
		return fmt.Errorf("%w: temp file creation error: %v", ErrRegistryWrite, err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("%w: write error: %v", ErrRegistryWrite, err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("%w: close error: %v", ErrRegistryWrite, err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("%w: rename error: %v", ErrRegistryWrite, err)
	}

	return nil
}
