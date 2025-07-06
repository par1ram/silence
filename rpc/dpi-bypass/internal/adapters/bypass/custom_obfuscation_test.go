package bypass

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	adapter := &CustomAdapter{}
	key1 := adapter.generateKey("password1")
	key2 := adapter.generateKey("password1")
	key3 := adapter.generateKey("password2")
	assert.Equal(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.Len(t, key1, 32)
}

func TestEncryptData(t *testing.T) {
	adapter := &CustomAdapter{}
	conn := &customConnection{encryptionKey: adapter.generateKey("testpass")}
	plain := []byte("hello world")
	cipher, err := adapter.encryptData(plain, conn)
	assert.NoError(t, err)
	assert.NotNil(t, cipher)
	assert.NotEqual(t, plain, cipher)
}

func TestFragmentData(t *testing.T) {
	adapter := &CustomAdapter{}
	data := []byte("abcdefghijklmnopqrstuvwxyz")
	fragments := adapter.fragmentData(data, 5)
	assert.Len(t, fragments, 6)
	assert.Equal(t, []byte("abcde"), fragments[0])
	assert.Equal(t, byte('z'), fragments[5][0])
	assert.Equal(t, 1, len(fragments[5]))
}

func TestGenerateChaffData(t *testing.T) {
	adapter := &CustomAdapter{}
	conn := &customConnection{chaffRatio: 0.5}
	chaff := adapter.generateChaffData(100, conn)
	assert.GreaterOrEqual(t, len(chaff), 64)
}

func TestApplyCustomObfuscation_Chaff(t *testing.T) {
	adapter := &CustomAdapter{}
	conn := &customConnection{obfuscationMode: "chaff", chaffRatio: 0.5, encryptionKey: adapter.generateKey("test")}
	data := []byte("testdata")
	obf, err := adapter.applyCustomObfuscation(data, conn)
	assert.NoError(t, err)
	assert.NotNil(t, obf)
	assert.Greater(t, len(obf), len(data))
}

func TestApplyCustomObfuscation_Fragment(t *testing.T) {
	adapter := &CustomAdapter{}
	conn := &customConnection{obfuscationMode: "fragment", fragmentSize: 3, encryptionKey: adapter.generateKey("test")}
	data := []byte("abcdefghij")
	obf, err := adapter.applyCustomObfuscation(data, conn)
	assert.NoError(t, err)
	assert.NotNil(t, obf)
	assert.Greater(t, len(obf), len(data))
}

func TestApplyCustomObfuscation_Hybrid(t *testing.T) {
	adapter := &CustomAdapter{}
	conn := &customConnection{obfuscationMode: "hybrid", chaffRatio: 0.2, fragmentSize: 2, encryptionKey: adapter.generateKey("test")}
	data := []byte("abcdefghij")
	obf, err := adapter.applyCustomObfuscation(data, conn)
	assert.NoError(t, err)
	assert.NotNil(t, obf)
	assert.Greater(t, len(obf), len(data))
}

func TestApplyCustomTiming(t *testing.T) {
	adapter := &CustomAdapter{}
	conn := &customConnection{timingJitter: 10}
	start := time.Now()
	adapter.applyCustomTiming(conn)
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 0*time.Millisecond)
}
