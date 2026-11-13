package model

import (
	"strings"
	"testing"
)

func TestCryptoDeterministicEncryption(t *testing.T) {
	setupTestDB(t)

	plain := "alice@example.com"
	enc1, err := EncryptStringDeterministically(plain)
	if err != nil {
		t.Fatalf("unexpected error encrypting: %v", err)
	}
	if !strings.HasPrefix(enc1, encPrefix) {
		t.Fatalf("expected encrypted string to have prefix %s, got: %s", encPrefix, enc1)
	}

	enc2, err := EncryptStringDeterministically(plain)
	if err != nil {
		t.Fatalf("unexpected error encrypting again: %v", err)
	}
	if enc1 != enc2 {
		t.Fatalf("expected deterministic encryption to produce identical ciphertexts, got %s and %s", enc1, enc2)
	}

	dec, err := DecryptStringDeterministically(enc1)
	if err != nil {
		t.Fatalf("unexpected error decrypting: %v", err)
	}
	if dec != plain {
		t.Fatalf("expected decrypted string %s, got %s", plain, dec)
	}
}

func TestEncryptedUserAtRestInDatabase(t *testing.T) {
	setupTestDB(t)

	user := NewUser(
		"SecretMoniker",
		"secretname",
		"secret@example.com",
		"1234567890",
		"hashed_password",
		"",
	)
	if err := user.Insert(); err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// 1. Raw database query to verify fields are encrypted at rest
	var rawMoniker, rawName, rawEmail string
	err := Sqlx.QueryRow("SELECT moniker, name, email FROM app_users WHERE id = ?", user.Id).Scan(&rawMoniker, &rawName, &rawEmail)
	if err != nil {
		t.Fatalf("failed to query raw app_users: %v", err)
	}

	if !strings.HasPrefix(rawMoniker, encPrefix) || rawMoniker == "SecretMoniker" {
		t.Fatalf("expected moniker to be encrypted at rest in DB, got: %s", rawMoniker)
	}
	if !strings.HasPrefix(rawName, encPrefix) || rawName == "secretname" {
		t.Fatalf("expected name to be encrypted at rest in DB, got: %s", rawName)
	}
	if !strings.HasPrefix(rawEmail, encPrefix) || rawEmail == "secret@example.com" {
		t.Fatalf("expected email to be encrypted at rest in DB, got: %s", rawEmail)
	}

	// 2. Querying through model transparently decrypts
	byID, err := NewUserFromId(user.Id)
	if err != nil {
		t.Fatalf("failed to fetch user by ID: %v", err)
	}
	if byID.Moniker != "SecretMoniker" {
		t.Errorf("expected decrypted Moniker 'SecretMoniker', got '%s'", byID.Moniker)
	}
	if byID.Name != "secretname" {
		t.Errorf("expected decrypted Name 'secretname', got '%s'", byID.Name)
	}
	if byID.Email != "secret@example.com" {
		t.Errorf("expected decrypted Email 'secret@example.com', got '%s'", byID.Email)
	}

	// 3. Exact lookups using Name and Email
	byName, err := NewUserFromName("secretname")
	if err != nil || byName.Id != user.Id {
		t.Fatalf("failed to find user by Name, got err: %v, user: %+v", err, byName)
	}

	byEmail, err := NewUserFromEmail("secret@example.com")
	if err != nil || byEmail.Id != user.Id {
		t.Fatalf("failed to find user by Email, got err: %v, user: %+v", err, byEmail)
	}

	// 4. Search users
	searchRes, err := SearchAllUsers("secret", 10)
	if err != nil {
		t.Fatalf("failed to search users: %v", err)
	}
	found := false
	for _, u := range searchRes {
		if u.Id == user.Id {
			found = true
			if u.Moniker != "SecretMoniker" || u.Name != "secretname" || u.Email != "secret@example.com" {
				t.Errorf("search returned user with un-decrypted fields: %+v", u)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected to find user in SearchAllUsers results, found %d results", len(searchRes))
	}

	// 5. Test empty search limit pushdown
	emptySearchRes, err := SearchAllUsers("", 1)
	if err != nil {
		t.Fatalf("failed to run empty search: %v", err)
	}
	if len(emptySearchRes) > 1 {
		t.Errorf("expected at most 1 result from empty search with limit 1, got %d", len(emptySearchRes))
	}

	// 6. Test search limit enforcement with query
	limitedSearchRes, err := SearchAllUsers("e", 1)
	if err != nil {
		t.Fatalf("failed to run limited search: %v", err)
	}
	if len(limitedSearchRes) > 1 {
		t.Errorf("expected at most 1 result from query search with limit 1, got %d", len(limitedSearchRes))
	}
}
