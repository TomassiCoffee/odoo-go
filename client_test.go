package odoo

import (
	"encoding/json"
	"testing"
)

func TestFalseString(t *testing.T) {
	var s FalseString
	if err := json.Unmarshal([]byte(`false`), &s); err != nil {
		t.Fatal(err)
	}
	if s.Valid {
		t.Fatal("false should be invalid")
	}
	if err := json.Unmarshal([]byte(`"abc"`), &s); err != nil {
		t.Fatal(err)
	}
	if !s.Valid || s.String != "abc" {
		t.Fatalf("unexpected value: %#v", s)
	}
}

func TestMany2One(t *testing.T) {
	var m Many2One
	if err := json.Unmarshal([]byte(`[7,"Cash"]`), &m); err != nil {
		t.Fatal(err)
	}
	if !m.Valid || m.ID != 7 || m.DisplayName != "Cash" {
		t.Fatalf("unexpected value: %#v", m)
	}
}
