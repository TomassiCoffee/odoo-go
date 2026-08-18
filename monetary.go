package odoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Monetary stores Odoo monetary values as an exact decimal number.
//
// Odoo sends monetary fields as JSON numbers. Decoding them into float64 can
// introduce binary floating-point rounding errors before the value reaches the
// application. Monetary keeps the original decimal value exact while still
// marshaling back to a plain JSON number for Odoo writes.
type Monetary struct {
	value *big.Rat
}

func (m *Monetary) GetRat() big.Rat{
	return *m.value
}

func (m *Monetary) Abs() Monetary {
	tmp := new(big.Rat)
	return Monetary{value: tmp.Abs(m.value)}
}

func NewMonetary(value string) (Monetary, error) {
	var m Monetary
	if err := m.setString(value); err != nil {
		return Monetary{}, err
	}
	return m, nil
}

func MustMonetary(value string) Monetary {
	m, err := NewMonetary(value)
	if err != nil {
		panic(err)
	}
	return m
}

func (m *Monetary) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || bytes.Equal(data, []byte("false")) || len(data) == 0 {
		*m = Monetary{}
		return nil
	}

	var text string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
	} else {
		text = string(data)
	}
	return m.setString(text)
}

func (m Monetary) MarshalJSON() ([]byte, error) {
	if m.value == nil {
		return []byte("0"), nil
	}
	return []byte(m.String()), nil
}

func (m Monetary) String() string {
	if m.value == nil {
		return "0"
	}
	return ratDecimalString(m.value)
}

func (m Monetary) IsZero() bool {
	return m.value == nil || m.value.Sign() == 0
}

func (m Monetary) Float64() float64 {
	if m.value == nil {
		return 0
	}
	f, _ := m.value.Float64()
	return f
}

func (m Monetary) Cents() int64 {
	return m.MinorUnits(2)
}

func (m Monetary) MinorUnits(scale int) int64 {
	if m.value == nil {
		return 0
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	scaled := new(big.Rat).Mul(m.value, new(big.Rat).SetInt(factor))
	return roundRatToInt64(scaled)
}

func (m Monetary) DivFloat64(divisor float64) Monetary {
	if m.value == nil || divisor == 0 {
		return Monetary{}
	}
	divisorRat, ok := new(big.Rat).SetString(strconv.FormatFloat(divisor, 'f', -1, 64))
	if !ok || divisorRat.Sign() == 0 {
		return Monetary{}
	}
	return Monetary{value: new(big.Rat).Quo(m.value, divisorRat)}
}

func (m *Monetary) setString(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		*m = Monetary{}
		return nil
	}
	rat, ok := new(big.Rat).SetString(value)
	if !ok {
		return fmt.Errorf("invalid monetary value %q", value)
	}
	m.value = rat
	return nil
}

func ratDecimalString(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	for scale := 0; scale <= 18; scale++ {
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		scaled := new(big.Int).Mul(r.Num(), factor)
		if new(big.Int).Rem(scaled, r.Denom()).Sign() == 0 {
			quotient := new(big.Int).Quo(scaled, r.Denom())
			return formatScaledInt(quotient, scale)
		}
	}
	return strings.TrimRight(strings.TrimRight(r.FloatString(18), "0"), ".")
}

func formatScaledInt(value *big.Int, scale int) string {
	if scale == 0 {
		return value.String()
	}
	negative := value.Sign() < 0
	text := new(big.Int).Abs(value).String()
	for len(text) <= scale {
		text = "0" + text
	}
	text = text[:len(text)-scale] + "." + text[len(text)-scale:]
	text = strings.TrimRight(strings.TrimRight(text, "0"), ".")
	if negative {
		text = "-" + text
	}
	return text
}

func roundRatToInt64(r *big.Rat) int64 {
	num := new(big.Int).Set(r.Num())
	den := r.Denom()
	quotient, remainder := new(big.Int).QuoRem(num, den, new(big.Int))
	twiceRemainder := new(big.Int).Abs(remainder)
	twiceRemainder.Mul(twiceRemainder, big.NewInt(2))
	if twiceRemainder.Cmp(den) >= 0 {
		if r.Sign() >= 0 {
			quotient.Add(quotient, big.NewInt(1))
		} else {
			quotient.Sub(quotient, big.NewInt(1))
		}
	}
	return quotient.Int64()
}
