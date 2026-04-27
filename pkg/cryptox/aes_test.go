package cryptox

import "testing"

func TestAesEncrypt(t *testing.T) {
	aes := NewAes(WithSafe(false), WithKey("www.zorktech.com"))
	decrypt, err := aes.Encrypt("zork.8888")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(decrypt)
}

func TestAesDecrypt(t *testing.T) {
	aes := NewAes(WithSafe(false), WithKey("www.zorktech.com"))
	decrypt, err := aes.Decrypt("jYxNn04OJRrvFjwU+VeD9Q==")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(decrypt))
}
