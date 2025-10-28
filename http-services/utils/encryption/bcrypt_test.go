package encryption

import (
	"testing"
)

func TestHashPasswordWithBcrypt(t *testing.T) {
	password := "test-password-123"

	// 测试哈希生成
	hash, err := HashPasswordWithBcrypt(password)
	if err != nil {
		t.Fatalf("HashPasswordWithBcrypt failed: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPasswordWithBcrypt returned empty hash")
	}

	// 哈希值不应该等于原始密码
	if hash == password {
		t.Error("Hash should not equal original password")
	}

	// 检查是否为 bcrypt 格式
	if !IsBcryptHash(hash) {
		t.Errorf("Hash is not in bcrypt format: %s", hash)
	}
}

func TestVerifyBcryptPassword(t *testing.T) {
	password := "correct-password"
	wrongPassword := "wrong-password"

	// 生成哈希
	hash, err := HashPasswordWithBcrypt(password)
	if err != nil {
		t.Fatalf("HashPasswordWithBcrypt failed: %v", err)
	}

	// 测试正确密码验证
	if !VerifyBcryptPassword(password, hash) {
		t.Error("VerifyBcryptPassword should return true for correct password")
	}

	// 测试错误密码验证
	if VerifyBcryptPassword(wrongPassword, hash) {
		t.Error("VerifyBcryptPassword should return false for wrong password")
	}

	// 测试空密码
	if VerifyBcryptPassword("", hash) {
		t.Error("VerifyBcryptPassword should return false for empty password")
	}
}

func TestIsBcryptHash(t *testing.T) {
	tests := []struct {
		name  string
		hash  string
		want  bool
	}{
		{
			name: "valid bcrypt hash with $2a",
			hash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			want: true,
		},
		{
			name: "valid bcrypt hash with $2b",
			hash: "$2b$12$KIXqRbCDRqJhNvz4zXvCIOMhDgHvMh8.pIJvdXVZQfJvKVLDKZvSq",
			want: true,
		},
		{
			name: "invalid hash - too short",
			hash: "$2a$10",
			want: false,
		},
		{
			name: "invalid hash - wrong format",
			hash: "not-a-bcrypt-hash",
			want: false,
		},
		{
			name: "invalid hash - empty",
			hash: "",
			want: false,
		},
		{
			name: "invalid hash - plain text",
			hash: "plain-text-password",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBcryptHash(tt.hash)
			if got != tt.want {
				t.Errorf("IsBcryptHash(%q) = %v, want %v", tt.hash, got, tt.want)
			}
		})
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	password := "same-password"

	// 生成两次哈希
	hash1, err1 := HashPasswordWithBcrypt(password)
	hash2, err2 := HashPasswordWithBcrypt(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("HashPasswordWithBcrypt failed: %v, %v", err1, err2)
	}

	// bcrypt 每次生成的哈希应该不同（因为使用了随机盐）
	if hash1 == hash2 {
		t.Error("Two hashes of the same password should be different (bcrypt uses random salt)")
	}

	// 但两个哈希都应该能验证原始密码
	if !VerifyBcryptPassword(password, hash1) {
		t.Error("First hash should verify the password")
	}
	if !VerifyBcryptPassword(password, hash2) {
		t.Error("Second hash should verify the password")
	}
}
