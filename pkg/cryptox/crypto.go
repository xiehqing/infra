package cryptox

type Crypto interface {
	// Encrypt 加密
	Encrypt(data string) (string, error)
	// Decrypt 解密
	Decrypt(data string) (string, error)
}
