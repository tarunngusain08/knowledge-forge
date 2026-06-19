package codeintel

import "testing"

func TestRepositoryPolicyRejectsLocalPathByDefault(t *testing.T) {
	policy := NewRepositoryPolicy(false, nil)

	if err := policy.ValidateRegistration("", "/srv/repos/private"); err == nil {
		t.Fatal("expected local_path registration to be disabled")
	}
}

func TestRepositoryPolicyAllowsConfiguredLocalPath(t *testing.T) {
	policy := NewRepositoryPolicy(true, nil)

	if err := policy.ValidateRegistration("", "/srv/repos/trusted"); err != nil {
		t.Fatalf("expected trusted local_path to be allowed: %v", err)
	}
}

func TestRepositoryPolicyAllowsApprovedHTTPSHosts(t *testing.T) {
	policy := NewRepositoryPolicy(false, nil)

	if err := policy.ValidateRegistration("https://github.com/example/service.git", ""); err != nil {
		t.Fatalf("expected github remote to be allowed: %v", err)
	}
	if err := policy.ValidateRegistration("https://gitlab.com/example/service.git", ""); err != nil {
		t.Fatalf("expected gitlab remote to be allowed: %v", err)
	}
}

func TestRepositoryPolicyRejectsUnsafeRemotes(t *testing.T) {
	policy := NewRepositoryPolicy(false, []string{"github.com", "git.example.com"})
	tests := []string{
		"file:///etc/passwd",
		"ssh://github.com/example/service.git",
		"https://localhost/example/service.git",
		"https://127.0.0.1/example/service.git",
		"https://10.0.0.5/example/service.git",
		"https://172.16.0.5/example/service.git",
		"https://192.168.1.10/example/service.git",
		"https://169.254.169.254/latest/meta-data",
		"https://metadata.google.internal/computeMetadata/v1/",
		"https://evil.example.com/repo.git",
		"https://user:pass@github.com/example/service.git",
	}
	for _, remote := range tests {
		t.Run(remote, func(t *testing.T) {
			if err := policy.ValidateRegistration(remote, ""); err == nil {
				t.Fatal("expected remote to be rejected")
			}
		})
	}
}

func TestRepositoryPolicyAllowsConfiguredEnterpriseHost(t *testing.T) {
	policy := NewRepositoryPolicy(false, []string{"git.example.com"})

	if err := policy.ValidateRegistration("https://git.example.com/platform/service.git", ""); err != nil {
		t.Fatalf("expected configured enterprise host to be allowed: %v", err)
	}
}
