package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/telegram-go/backend/pkg/crypto"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	// Hash the password
	hash, err := crypto.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Check that the hash is different from the original password
	assert.NotEqual(t, password, hash)
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"

	// Hash the password
	hash, err := crypto.HashPassword(password)
	assert.NoError(t, err)

	// Check with correct password
	assert.True(t, crypto.CheckPassword(password, hash))

	// Check with wrong password
	assert.False(t, crypto.CheckPassword("wrongpassword", hash))
}

func TestCheckPassword_EmptyHash(t *testing.T) {
	result := crypto.CheckPassword("password", "")
	assert.False(t, result)
}
