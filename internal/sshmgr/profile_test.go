package sshmgr

import (
	"strings"
	"testing"
)

func TestParseSSHCommand(t *testing.T) {
	cases := []struct {
		in       string
		wantHost string
		wantUser string
		wantPort int
		wantKey  string
		wantJump string
	}{
		{"ssh root@example.com", "example.com", "root", 22, "", ""},
		{"ssh -p 2222 deploy@10.0.0.5", "10.0.0.5", "deploy", 2222, "", ""},
		{"ssh -i ~/.ssh/id_ed25519 git@github.com", "github.com", "git", 22, "~/.ssh/id_ed25519", ""},
		{"ssh -J bastion user@internal", "internal", "user", 22, "", "bastion"},
		{"ssh -o ProxyJump=jump user@host -p 22", "host", "user", 22, "", "jump"},
		{"sshpass -p secret ssh admin@box", "box", "admin", 22, "", ""},
	}
	for _, c := range cases {
		p, ok := ParseSSHCommand(c.in)
		if !ok {
			t.Fatalf("ParseSSHCommand(%q) failed to parse", c.in)
		}
		if p.Host != c.wantHost || p.User != c.wantUser || p.Port != c.wantPort {
			t.Errorf("ParseSSHCommand(%q) = host=%q user=%q port=%d; want host=%q user=%q port=%d",
				c.in, p.Host, p.User, p.Port, c.wantHost, c.wantUser, c.wantPort)
		}
		if p.KeyPath != c.wantKey {
			t.Errorf("ParseSSHCommand(%q) key=%q; want %q", c.in, p.KeyPath, c.wantKey)
		}
		if p.JumpHost != c.wantJump {
			t.Errorf("ParseSSHCommand(%q) jump=%q; want %q", c.in, p.JumpHost, c.wantJump)
		}
	}
}

func TestParseSSHCommandRejectsGarbage(t *testing.T) {
	for _, in := range []string{"", "ls -la", "ssh", "ssh -p 22"} {
		if _, ok := ParseSSHCommand(in); ok {
			t.Errorf("ParseSSHCommand(%q) unexpectedly parsed", in)
		}
	}
}

func TestSSHArgv(t *testing.T) {
	p := &Profile{
		Host: "example.com", User: "root", Port: 2222,
		Auth: AuthKey, KeyPath: "/k", JumpHost: "bastion",
		KnownHosts: KnownHostsAccept, Keepalive: 30,
	}
	got := CommandString(p.SSHArgv())
	for _, want := range []string{"ssh", "-p 2222", "-i /k", "-J bastion", "StrictHostKeyChecking=accept-new", "ServerAliveInterval=30", "root@example.com"} {
		if !strings.Contains(got, want) {
			t.Errorf("SSHArgv missing %q in %q", want, got)
		}
	}
}

func TestSCPArgvPortFlag(t *testing.T) {
	p := &Profile{Host: "h", User: "u", Port: 2200}
	up := CommandString(p.SCPArgv("/local", "/remote", true, true))
	if !strings.Contains(up, "-P 2200") || !strings.Contains(up, "-r") {
		t.Errorf("scp upload argv wrong: %q", up)
	}
	if !strings.Contains(up, "/local u@h:/remote") {
		t.Errorf("scp upload order wrong: %q", up)
	}
	down := CommandString(p.SCPArgv("/local", "/remote", false, false))
	if !strings.Contains(down, "u@h:/remote /local") {
		t.Errorf("scp download order wrong: %q", down)
	}
}

func TestWrapSSHPass(t *testing.T) {
	argv := []string{"ssh", "-N", "u@h"}
	// No password: unchanged.
	if got := WrapSSHPass(argv, ""); len(got) != 3 || got[0] != "ssh" {
		t.Errorf("WrapSSHPass with empty password changed argv: %v", got)
	}
	// With password: prefixed with sshpass -e, password not in argv.
	got := WrapSSHPass(argv, "s3cr3t")
	if got[0] != "sshpass" || got[1] != "-e" {
		t.Errorf("WrapSSHPass prefix wrong: %v", got)
	}
	if strings.Contains(strings.Join(got, " "), "s3cr3t") {
		t.Errorf("password leaked into argv: %v", got)
	}
}

func TestExecArgvBatch(t *testing.T) {
	p := &Profile{Host: "h", User: "u"}
	batch := CommandString(p.ExecArgv("/tmp/s", "ls", true))
	if !strings.Contains(batch, "BatchMode=yes") {
		t.Errorf("batch ExecArgv missing BatchMode=yes: %q", batch)
	}
	interactive := CommandString(p.ExecArgv("/tmp/s", "ls", false))
	if strings.Contains(interactive, "BatchMode=yes") {
		t.Errorf("interactive ExecArgv should not force BatchMode=yes: %q", interactive)
	}
}

func TestTunnelArgv(t *testing.T) {
	p := &Profile{Host: "h", User: "u"}
	local := &TunnelSpec{Type: TunnelLocal, BindPort: 8080, TargetHost: "db", TargetPort: 5432}
	if got := CommandString(local.Argv(p)); !strings.Contains(got, "-L localhost:8080:db:5432") {
		t.Errorf("local tunnel argv wrong: %q", got)
	}
	dyn := &TunnelSpec{Type: TunnelDynamic, BindPort: 1080}
	if got := CommandString(dyn.Argv(p)); !strings.Contains(got, "-D localhost:1080") {
		t.Errorf("dynamic tunnel argv wrong: %q", got)
	}
}
