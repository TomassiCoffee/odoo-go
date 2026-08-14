package odoo

import (
	"encoding/json"
	"testing"
)

func TestMonetaryJSONKeepsDecimalExact(t *testing.T) {
	var got struct {
		Amount Monetary `json:"amount"`
	}
	if err := json.Unmarshal([]byte(`{"amount":12.34}`), &got); err != nil {
		t.Fatalf("unmarshal monetary: %v", err)
	}
	if got.Amount.String() != "12.34" {
		t.Fatalf("Amount.String() = %q, want 12.34", got.Amount.String())
	}
	if got.Amount.Cents() != 1234 {
		t.Fatalf("Amount.Cents() = %d, want 1234", got.Amount.Cents())
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal monetary: %v", err)
	}
	if string(encoded) != `{"amount":12.34}` {
		t.Fatalf("json.Marshal() = %s, want {\"amount\":12.34}", encoded)
	}
}

func TestMonetaryDivFloat64ForUnitPrice(t *testing.T) {
	amount := MustMonetary("10")
	unit := amount.DivFloat64(3)
	if unit.Cents() != 333 {
		t.Fatalf("10 / 3 as cents = %d, want 333", unit.Cents())
	}
}
