package browser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMaskProxyCredentials 校验代理日志脱敏：绝不能把用户名/密码打进日志。
func TestMaskProxyCredentials(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "空字符串", input: "", want: ""},
		{name: "无认证信息原样返回", input: "http://127.0.0.1:8080", want: "http://127.0.0.1:8080"},
		{name: "用户名+密码都脱敏", input: "http://user:pass@host:8080", want: "http://***:***@host:8080"},
		{name: "仅用户名脱敏", input: "http://user@host:8080", want: "http://***@host:8080"},
		{name: "socks5带认证", input: "socks5://alice:secret@127.0.0.1:1080", want: "socks5://***:***@127.0.0.1:1080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, maskProxyCredentials(tt.input))
		})
	}
}

func TestSetBrowserBinPath(t *testing.T) {
	original := configuredBrowserBinPath
	t.Cleanup(func() { configuredBrowserBinPath = original })

	binPath := filepath.Join(t.TempDir(), "chrome")
	assert.NoError(t, os.WriteFile(binPath, []byte("test"), 0o755))

	resolved, err := SetBrowserBinPath(binPath)
	assert.NoError(t, err)
	assert.Equal(t, binPath, resolved)
	assert.Equal(t, binPath, configuredBrowserBinPath)
	got, err := BrowserBinPath()
	assert.NoError(t, err)
	assert.Equal(t, binPath, got)
}

func TestSetBrowserBinPathRejectsInvalidPaths(t *testing.T) {
	original := configuredBrowserBinPath
	t.Cleanup(func() { configuredBrowserBinPath = original })

	_, err := SetBrowserBinPath(filepath.Join(t.TempDir(), "missing"))
	assert.ErrorContains(t, err, "浏览器二进制不可用")

	notExecutable := filepath.Join(t.TempDir(), "chrome")
	assert.NoError(t, os.WriteFile(notExecutable, []byte("test"), 0o644))
	_, err = SetBrowserBinPath(notExecutable)
	assert.ErrorContains(t, err, "浏览器二进制不可执行")
}

// TestOptions 校验 Option 正确写入 browserConfig（New+Option 的接线）。
func TestOptions(t *testing.T) {
	cfg := &browserConfig{}
	WithFingerprintSeed(98759)(cfg)
	WithProxy("http://127.0.0.1:8080")(cfg)

	assert.Equal(t, 98759, cfg.fingerprintSeed)
	assert.Equal(t, "http://127.0.0.1:8080", cfg.proxy)
}

// TestOptions_Defaults 未传 Option 时各字段为零值（回退随机 seed / 不设代理）。
func TestOptions_Defaults(t *testing.T) {
	cfg := &browserConfig{}
	assert.Equal(t, 0, cfg.fingerprintSeed)
	assert.Equal(t, "", cfg.proxy)
}
