package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plain := "sk-kimi-87JDBrdkbwV6Hf9lfFPJbVtuxTjHYK4NJ17Kn592qoua0c0mcLBwcWFFk"
	enc, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if !strings.HasPrefix(enc, Prefix) {
		t.Fatalf("密文缺少前缀: %s", enc)
	}
	dec, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if dec != plain {
		t.Fatalf("往返不一致: got %q want %q", dec, plain)
	}
}

func TestEncryptEmpty(t *testing.T) {
	enc, err := Encrypt("")
	if err != nil || enc != "" {
		t.Fatalf("空明文应返回空串: enc=%q err=%v", enc, err)
	}
	if dec, _ := Decrypt(""); dec != "" {
		t.Fatalf("空串解密应返回空串: %q", dec)
	}
}

func TestDecryptPlaintextPassthrough(t *testing.T) {
	// 历史明文（无前缀）原样返回，视为未迁移数据
	dec, err := Decrypt("sk-plaintext-key")
	if err != nil {
		t.Fatalf("明文 passthrough 不应报错: %v", err)
	}
	if dec != "sk-plaintext-key" {
		t.Fatalf("明文应原样返回: %q", dec)
	}
}

func TestDecryptTampered(t *testing.T) {
	enc, _ := Encrypt("secret-key-1234567890")
	if _, err := Decrypt(enc + "x"); err == nil {
		t.Fatal("篡改后的密文应解密失败")
	}
	if _, err := Decrypt(Prefix + "not-base64!!!"); err == nil {
		t.Fatal("非法 base64 应报错")
	}
}

func TestMask(t *testing.T) {
	cases := map[string]string{
		"":                                        "",
		"short":                                   "sh****",
		"sk-kimi-87JDBrdkbwV6Hf9lfFPJbVtuxTjHYK4NJ17Kn592qoua0c0mcLBwcWFFk": "sk-kimi****WFFk",
	}
	for in, want := range cases {
		if got := Mask(in); got != want {
			t.Fatalf("Mask(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEncryptNonDeterministic(t *testing.T) {
	a, _ := Encrypt("same-key")
	b, _ := Encrypt("same-key")
	if a == b {
		t.Fatal("相同明文两次加密应产生不同密文（随机 nonce）")
	}
}
