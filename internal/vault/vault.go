// Package vault implements Zet-SSH's local encrypted secret store.
//
// The vault is a single file (vault.zet) holding a JSON envelope:
// Argon2id derives a 32-byte key from the master password, and the
// secret map is sealed with XChaCha20-Poly1305. Nothing sensitive is
// ever written to disk in plaintext.
package vault

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	ErrLocked        = errors.New("vault is locked")
	ErrWrongPassword = errors.New("wrong master password")
	ErrNoVault       = errors.New("vault does not exist yet")
)

type envelope struct {
	Version    int    `json:"version"`
	KDF        string `json:"kdf"`
	Salt       string `json:"salt"`
	ArgonTime  uint32 `json:"argon_time"`
	ArgonMemKB uint32 `json:"argon_mem_kb"`
	ArgonPar   uint8  `json:"argon_par"`
	Cipher     string `json:"cipher"`
	Nonce      string `json:"nonce"`
	Data       string `json:"data"`
}

// Vault is a lockable map of secret name -> secret value.
type Vault struct {
	mu       sync.Mutex
	path     string
	key      []byte // nil while locked
	salt     []byte
	secrets  map[string]string
	lastUsed time.Time
}

var global *Vault

// Global returns the process-wide vault bound to the given config dir.
func Global(configDir string) *Vault {
	if global == nil {
		global = &Vault{path: filepath.Join(configDir, "vault.zet")}
	}
	return global
}

func (v *Vault) Path() string { return v.path }

// Exists reports whether a vault file has been created.
func (v *Vault) Exists() bool {
	_, err := os.Stat(v.path)
	return err == nil
}

// IsUnlocked reports whether secrets are currently accessible.
func (v *Vault) IsUnlocked() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.key != nil
}

func deriveKey(password string, salt []byte, timeCost, memKB uint32, par uint8) []byte {
	return argon2.IDKey([]byte(password), salt, timeCost, memKB, par, chacha20poly1305.KeySize)
}

// Create initializes a brand-new vault protected by the master password.
func (v *Vault) Create(master string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, err := os.Stat(v.path); err == nil {
		return errors.New("vault already exists")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	v.salt = salt
	v.key = deriveKey(master, salt, 3, 64*1024, 4)
	v.secrets = map[string]string{}
	v.lastUsed = time.Now()
	return v.saveLocked()
}

// Unlock decrypts the vault file with the master password.
func (v *Vault) Unlock(master string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	raw, err := os.ReadFile(v.path)
	if err != nil {
		return ErrNoVault
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return err
	}
	salt, err := base64.StdEncoding.DecodeString(env.Salt)
	if err != nil {
		return err
	}
	nonce, err := base64.StdEncoding.DecodeString(env.Nonce)
	if err != nil {
		return err
	}
	data, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		return err
	}

	key := deriveKey(master, salt, env.ArgonTime, env.ArgonMemKB, env.ArgonPar)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return err
	}
	plain, err := aead.Open(nil, nonce, data, nil)
	if err != nil {
		return ErrWrongPassword
	}
	secrets := map[string]string{}
	if err := json.Unmarshal(plain, &secrets); err != nil {
		return err
	}

	v.salt = salt
	v.key = key
	v.secrets = secrets
	v.lastUsed = time.Now()
	return nil
}

// Lock wipes the in-memory key and secrets.
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.key {
		v.key[i] = 0
	}
	v.key = nil
	v.secrets = nil
}

// MaybeAutoLock locks the vault if it has been idle longer than d.
// Returns true if it locked.
func (v *Vault) MaybeAutoLock(d time.Duration) bool {
	v.mu.Lock()
	idle := v.key != nil && d > 0 && time.Since(v.lastUsed) > d
	v.mu.Unlock()
	if idle {
		v.Lock()
	}
	return idle
}

func (v *Vault) touch() { v.lastUsed = time.Now() }

// Get returns a secret by name.
func (v *Vault) Get(name string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.key == nil {
		return "", ErrLocked
	}
	v.touch()
	val, ok := v.secrets[name]
	if !ok {
		return "", errors.New("secret not found: " + name)
	}
	return val, nil
}

// Set stores a secret and persists the vault.
func (v *Vault) Set(name, value string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.key == nil {
		return ErrLocked
	}
	v.touch()
	v.secrets[name] = value
	return v.saveLocked()
}

// Delete removes a secret and persists the vault.
func (v *Vault) Delete(name string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.key == nil {
		return ErrLocked
	}
	v.touch()
	delete(v.secrets, name)
	return v.saveLocked()
}

// List returns secret names, optionally filtered by prefix, sorted.
func (v *Vault) List(prefix string) ([]string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.key == nil {
		return nil, ErrLocked
	}
	v.touch()
	var names []string
	for name := range v.secrets {
		if prefix == "" || strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// ChangeMasterPassword re-encrypts the vault under a new password.
func (v *Vault) ChangeMasterPassword(newMaster string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.key == nil {
		return ErrLocked
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	v.salt = salt
	v.key = deriveKey(newMaster, salt, 3, 64*1024, 4)
	return v.saveLocked()
}

func (v *Vault) saveLocked() error {
	plain, err := json.Marshal(v.secrets)
	if err != nil {
		return err
	}
	aead, err := chacha20poly1305.NewX(v.key)
	if err != nil {
		return err
	}
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	sealed := aead.Seal(nil, nonce, plain, nil)

	env := envelope{
		Version:    1,
		KDF:        "argon2id",
		Salt:       base64.StdEncoding.EncodeToString(v.salt),
		ArgonTime:  3,
		ArgonMemKB: 64 * 1024,
		ArgonPar:   4,
		Cipher:     "xchacha20-poly1305",
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Data:       base64.StdEncoding.EncodeToString(sealed),
	}
	raw, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(v.path), 0o700); err != nil {
		return err
	}
	tmp := v.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, v.path)
}

// Secret key naming helpers keep vault entries consistent.

func ProfilePasswordKey(profileID string) string   { return "profile/" + profileID + "/password" }
func ProfilePassphraseKey(profileID string) string { return "profile/" + profileID + "/passphrase" }
func SnippetVarKey(varName string) string          { return "var/" + varName }
