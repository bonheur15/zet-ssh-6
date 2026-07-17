package snippets

import (
	"reflect"
	"testing"
)

func TestVarNames(t *testing.T) {
	sn := &Snippet{Command: "psql -h ${HOST} -U ${USER} -d ${HOST}"}
	got := sn.VarNames()
	want := []string{"HOST", "USER"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("VarNames = %v, want %v", got, want)
	}
}

func TestExpand(t *testing.T) {
	sn := &Snippet{Command: "ssh ${USER}@${HOST}"}
	vals := map[string]string{"USER": "root", "HOST": "example.com"}
	out, missing := sn.Expand(func(n string) (string, bool) {
		v, ok := vals[n]
		return v, ok
	})
	if out != "ssh root@example.com" {
		t.Errorf("Expand = %q", out)
	}
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
}

func TestExpandMissing(t *testing.T) {
	sn := &Snippet{Command: "echo ${A} ${B}"}
	out, missing := sn.Expand(func(n string) (string, bool) {
		if n == "A" {
			return "1", true
		}
		return "", false
	})
	if out != "echo 1 ${B}" {
		t.Errorf("Expand = %q", out)
	}
	if !reflect.DeepEqual(missing, []string{"B"}) {
		t.Errorf("missing = %v, want [B]", missing)
	}
}
