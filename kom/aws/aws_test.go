package aws

import (
	"context"
	"testing"
	"time"

	"k8s.io/client-go/tools/clientcmd/api"
)

// TestBuildEnvVariables 测试环境变量构建逻辑
// 该测试确保 ExecConfig 能正确生成包含 AWS 凭证与自定义环境变量的列表
func TestBuildEnvVariables(t *testing.T) {
	ec := &ExecConfig{
		Env: map[string]string{
			"FOO": "BAR",
		},
		AccessKey:       "AKIA_TEST",
		SecretAccessKey: "SECRET_TEST",
		Region:          "us-east-1",
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		SessionName:     "test-session",
	}

	vars := ec.BuildEnvVariables()

	expectContains := []string{
		"FOO=BAR",
		"AWS_ACCESS_KEY_ID=AKIA_TEST",
		"AWS_SECRET_ACCESS_KEY=SECRET_TEST",
		"AWS_DEFAULT_REGION=us-east-1",
		"AWS_ROLE_ARN=arn:aws:iam::123456789012:role/TestRole",
		"AWS_ROLE_SESSION_NAME=test-session",
	}
	for _, s := range expectContains {
		found := false
		for _, v := range vars {
			if v == s {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("环境变量未包含期望项: %s", s)
		}
	}
}

// TestValidateCommandSuccess 测试命令校验成功路径
// 使用系统自带的 echo 命令作为可用命令
func TestValidateCommandSuccess(t *testing.T) {
	exec := NewExecExecutor()
	if err := exec.ValidateCommand(&ExecConfig{Command: "echo"}); err != nil {
		t.Fatalf("ValidateCommand 失败: %v", err)
	}
}

// TestValidateCommandNotFound 测试命令校验失败路径
// 使用一个不存在的命令名，期望返回错误
func TestValidateCommandNotFound(t *testing.T) {
	exec := NewExecExecutor()
	if err := exec.ValidateCommand(&ExecConfig{Command: "cmd_not_exists_xyz"}); err == nil {
		t.Fatalf("ValidateCommand 应该失败，但返回了 nil")
	}
}

// TestExecuteCommandEcho 测试通过 echo 获取并解析 token
// 模拟 aws eks get-token 的返回结构，通过 echo 输出 JSON
func TestExecuteCommandEcho(t *testing.T) {
	exec := NewExecExecutor()
	// 构造一个未来时间，确保 token 有效
	exp := time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)
	json := `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec":{"interactive":false},"status":{"expirationTimestamp":"` + exp + `","token":"TEST_TOKEN"}}`
	cfg := &ExecConfig{
		Command: "echo",
		Args:    []string{json},
		Env:     map[string]string{},
	}

	resp, err := exec.ExecuteCommand(context.Background(), cfg)
	if err != nil {
		t.Fatalf("ExecuteCommand 失败: %v", err)
	}
	if resp.Status.Token != "TEST_TOKEN" {
		t.Fatalf("解析 token 错误，期望 TEST_TOKEN，得到 %s", resp.Status.Token)
	}
	if time.Until(resp.Status.ExpirationTimestamp) <= 0 {
		t.Fatalf("过期时间应在未来，得到 %v", resp.Status.ExpirationTimestamp)
	}
}

// TestGetTokenWithRetry 测试带重试的获取逻辑
// 使用 echo 保证第一次即可成功，重试逻辑不应触发错误
func TestGetTokenWithRetry(t *testing.T) {
	exec := NewExecExecutor()
	exp := time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)
	json := `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec":{"interactive":false},"status":{"expirationTimestamp":"` + exp + `","token":"RETRY_TOKEN"}}`
	cfg := &ExecConfig{
		Command: "echo",
		Args:    []string{json},
		Env:     map[string]string{},
	}

	resp, err := exec.GetTokenWithRetry(context.Background(), cfg, 1)
	if err != nil {
		t.Fatalf("GetTokenWithRetry 失败: %v", err)
	}
	if resp.Status.Token != "RETRY_TOKEN" {
		t.Fatalf("解析 token 错误，期望 RETRY_TOKEN，得到 %s", resp.Status.Token)
	}
}

// TestTokenManagerRefreshAndCache 测试刷新与缓存行为
// 通过 echo 模拟获取 token，验证缓存写入与读取
func TestTokenManagerRefreshAndCache(t *testing.T) {
	eks := &EKSAuthConfig{
		Region:     "us-east-1",
		ExecConfig: &ExecConfig{Command: "echo"},
		TokenCache: &TokenCache{},
	}
	// 构造 echo 输出 JSON
	exp := time.Now().Add(45 * time.Minute).UTC().Format(time.RFC3339)
	json := `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec":{"interactive":false},"status":{"expirationTimestamp":"` + exp + `","token":"CACHE_TOKEN"}}`
	eks.ExecConfig.Args = []string{json}

	tm, err := NewTokenManager(eks)
	if err != nil {
		t.Fatalf("NewTokenManager 失败: %v", err)
	}

	// 刷新并写入缓存
	if err := tm.RefreshToken(context.Background()); err != nil {
		t.Fatalf("RefreshToken 失败: %v", err)
	}

	// 校验缓存读取
	token, expires, valid := tm.GetTokenInfo()
	if token != "CACHE_TOKEN" || !valid {
		t.Fatalf("缓存信息不正确: token=%s valid=%v", token, valid)
	}
	if time.Until(expires) <= 0 {
		t.Fatalf("缓存过期时间应在未来，得到 %v", expires)
	}

	// ClearCache 清理缓存
	tm.ClearCache()
	if tok, _, v := tm.GetTokenInfo(); tok != "" || v {
		t.Fatalf("清理缓存失败，token=%s valid=%v", tok, v)
	}
}

// TestAuthProviderIntegration 测试 AuthProvider 与 TokenManager 集成
// 校验 GetToken / IsTokenValid / GetTokenExpiry 等方法行为
func TestAuthProviderIntegration(t *testing.T) {
	eks := &EKSAuthConfig{
		Region:     "us-west-2",
		ExecConfig: &ExecConfig{Command: "echo"},
		TokenCache: &TokenCache{},
		AccessKey:  "AKIA_TEST",
		SecretAccessKey: "SECRET_TEST",
	}
	exp := time.Now().Add(20 * time.Minute).UTC().Format(time.RFC3339)
	json := `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec":{"interactive":false},"status":{"expirationTimestamp":"` + exp + `","token":"AP_TOKEN"}}`
	eks.ExecConfig.Args = []string{json}

	tm, err := NewTokenManager(eks)
	if err != nil {
		t.Fatalf("NewTokenManager 失败: %v", err)
	}

	ap := NewAuthProvider()
	ap.SetEKSConfig(eks)
	ap.SetTokenManager(tm)

	// 获取 token
	token, expiry, err := ap.GetToken(context.Background())
	if err != nil {
		t.Fatalf("AuthProvider.GetToken 失败: %v", err)
	}
	if token == "" || time.Until(expiry) <= 0 {
		t.Fatalf("获取到的 token 或过期时间不正确: token=%s expiry=%v", token, expiry)
	}

	// 校验有效性
	if !ap.IsTokenValid() {
		t.Fatalf("IsTokenValid 返回 false，但期望 true")
	}
	if ap.GetTokenExpiry().IsZero() {
		t.Fatalf("GetTokenExpiry 不应为零值")
	}

	// 清理缓存
	ap.ClearTokenCache()
	if ap.IsTokenValid() {
		t.Fatalf("清理缓存后 IsTokenValid 应为 false")
	}

	// 触发刷新（不校验异步，仅调用以覆盖路径）
	ap.TriggerRefresh()
}

// TestSetEKSExecProvider 测试从 kubeconfig Exec 映射到内部 ExecConfig
// 验证 Env 转换为 map，及基本字段映射
func TestSetEKSExecProvider(t *testing.T) {
	ap := NewAuthProvider()
	eks := &EKSAuthConfig{
		AccessKey:       "AKIA_X",
		SecretAccessKey: "SECRET_X",
		Region:          "ap-southeast-1",
		RoleARN:         "arn:aws:iam::123456789012:role/RoleX",
		SessionName:     "sess-x",
	}
	ap.SetEKSConfig(eks)

	execCfg := &api.ExecConfig{
		Command: "aws",
		Args:    []string{"eks", "get-token"},
		Env: []api.ExecEnvVar{
			{Name: "FOO", Value: "BAR"},
			{Name: "BAZ", Value: "QUX"},
		},
	}

	if err := ap.SetEKSExecProvider(execCfg); err != nil {
		t.Fatalf("SetEKSExecProvider 失败: %v", err)
	}
	if ap.eksConfig.ExecConfig == nil {
		t.Fatalf("ExecConfig 未设置")
	}
	if ap.eksConfig.ExecConfig.Env["FOO"] != "BAR" || ap.eksConfig.ExecConfig.Env["BAZ"] != "QUX" {
		t.Fatalf("Env 转换为 map 失败: %+v", ap.eksConfig.ExecConfig.Env)
	}
	if ap.eksConfig.ExecConfig.AccessKey != "AKIA_X" || ap.eksConfig.ExecConfig.Region != "ap-southeast-1" {
		t.Fatalf("基本字段映射失败: %+v", ap.eksConfig.ExecConfig)
	}
}

