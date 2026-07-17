package vault

import (
	"path/filepath"
	"testing"
)

func TestVaultRoundTrip(t *testing.T) {
	dir := t.TempDir()
	v := &Vault{path: filepath.Join(dir, "vault.zet")}

	if v.Exists() {
		t.Fatal("new vault should not exist")
	}
	if err := v.Create("hunter2"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !v.Exists() || !v.IsUnlocked() {
		t.Fatal("vault should exist and be unlocked after Create")
	}
	if err := v.Set("api/token", "s3cr3t"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Re-open with a fresh instance to prove decryption works.
	v2 := &Vault{path: v.path}
	if err := v2.Unlock("hunter2"); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	got, err := v2.Get("api/token")
	if err != nil || got != "s3cr3t" {
		t.Fatalf("Get = %q, %v; want s3cr3t", got, err)
	}
}

func TestVaultWrongPassword(t *testing.T) {
	dir := t.TempDir()
	v := &Vault{path: filepath.Join(dir, "vault.zet")}
	if err := v.Create("correct"); err != nil {
		t.Fatal(err)
	}
	v2 := &Vault{path: v.path}
	if err := v2.Unlock("wrong"); err != ErrWrongPassword {
		t.Fatalf("Unlock wrong password: got %v, want ErrWrongPassword", err)
	}
}

func TestVaultLocked(t *testing.T) {
	dir := t.TempDir()
	v := &Vault{path: filepath.Join(dir, "vault.zet")}
	if err := v.Create("pw"); err != nil {
		t.Fatal(err)
	}
	v.Lock()
	if v.IsUnlocked() {
		t.Fatal("vault should be locked")
	}
	if _, err := v.Get("x"); err != ErrLocked {
		t.Fatalf("Get on locked vault: got %v, want ErrLocked", err)
	}
}
