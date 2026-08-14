package odoo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type ID int64

type Record map[string]any

type Domain []any

func Clause(field string, operator string, value any) []any {
	return []any{field, operator, value}
}

func Or() string  { return "|" }
func And() string { return "&" }
func Not() string { return "!" }

type SearchReadOptions struct {
	Domain  Domain
	Fields  []string
	Limit   *int
	Offset  *int
	Order   string
	Context map[string]any
}

func (o SearchReadOptions) Payload() map[string]any {
	payload := map[string]any{
		"domain": o.Domain,
	}
	if o.Fields != nil {
		payload["fields"] = o.Fields
	}
	if o.Limit != nil {
		payload["limit"] = *o.Limit
	}
	if o.Offset != nil {
		payload["offset"] = *o.Offset
	}
	if o.Order != "" {
		payload["order"] = o.Order
	}
	if o.Context != nil {
		payload["context"] = o.Context
	}
	return payload
}

type WriteOptions struct {
	IDs  []ID `json:"ids"`
	Vals any  `json:"vals"`
}

func (o WriteOptions) Payload() map[string]any {
	payload := map[string]any{}
	payload["ids"] = o.IDs
	payload["vals"] = o.Vals
	return payload
}

type Model[T any] interface {
	Name() string
	Fields() []string
}

type RawSearchReader interface {
	SearchReadRaw(ctx Context, model string, options SearchReadOptions) ([]json.RawMessage, error)
}

type RawWriter interface {
	WriteRaw(ctx Context, model string, options WriteOptions) (bool, error)
}

// Context is an alias so generated code does not depend on a concrete client type.
type Context interface {
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

type FieldMetadata struct {
	Type     string `json:"type"`
	Label    string `json:"label,omitempty"`
	Relation string `json:"relation,omitempty"`
	Required bool   `json:"required,omitempty"`
	Readonly bool   `json:"readonly,omitempty"`
}

type ModelMetadata map[string]FieldMetadata

type ModelRegistry map[string]ModelMetadata

// FalseString handles Odoo fields that are either a string or false/null.
type FalseString struct {
	String string
	Valid  bool
}

func NewFalseString(value string) FalseString {
	return FalseString{String: value, Valid: true}
}

func (s *FalseString) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "false" || text == "null" || text == "" {
		*s = FalseString{}
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = FalseString{String: value, Valid: true}
	return nil
}

func (s FalseString) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("false"), nil
	}
	return json.Marshal(s.String)
}

func (s FalseString) StringOr(fallback string) string {
	if !s.Valid {
		return fallback
	}
	return s.String
}

// Many2One handles Odoo many2one values: false or [id, display_name].
type Many2One struct {
	ID          ID
	DisplayName string
	Valid       bool
}

func NewMany2One(id ID, displayName string) Many2One {
	return Many2One{ID: id, DisplayName: displayName, Valid: true}
}

func (m *Many2One) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "false" || text == "null" || text == "" {
		*m = Many2One{}
		return nil
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) < 2 {
		return fmt.Errorf("invalid many2one value: %s", string(data))
	}

	var id ID
	if err := json.Unmarshal(raw[0], &id); err != nil {
		var f float64
		if err2 := json.Unmarshal(raw[0], &f); err2 != nil {
			return err
		}
		id = ID(f)
	}

	var name string
	if err := json.Unmarshal(raw[1], &name); err != nil {
		return err
	}

	*m = Many2One{ID: id, DisplayName: name, Valid: true}
	return nil
}

func (m Many2One) MarshalJSON() ([]byte, error) {
	if !m.Valid {
		return []byte("false"), nil
	}
	return json.Marshal([]any{m.ID, m.DisplayName})
}

// Many2OneWrite is the write form: id or false.
type Many2OneWrite struct {
	ID    ID
	Valid bool
}

func LinkMany2One(id ID) Many2OneWrite {
	return Many2OneWrite{ID: id, Valid: true}
}

func (m Many2OneWrite) MarshalJSON() ([]byte, error) {
	if !m.Valid {
		return []byte("false"), nil
	}
	return json.Marshal(m.ID)
}

type RelationCommand [3]any

type One2ManyCommand = RelationCommand
type Many2ManyCommand = RelationCommand

func RelationCreate(values map[string]any) RelationCommand { return RelationCommand{0, 0, values} }
func RelationUpdate(id ID, values map[string]any) RelationCommand {
	return RelationCommand{1, id, values}
}
func RelationDelete(id ID) RelationCommand { return RelationCommand{2, id, 0} }
func RelationUnlink(id ID) RelationCommand { return RelationCommand{3, id, 0} }
func RelationLink(id ID) RelationCommand   { return RelationCommand{4, id, 0} }
func RelationClear() RelationCommand       { return RelationCommand{5, 0, 0} }
func RelationSet(ids []ID) RelationCommand { return RelationCommand{6, 0, ids} }

func FieldNameList(fields ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}

func IntPtr(value int) *int { return &value }

func AsID(value any) ID {
	switch v := value.(type) {
	case ID:
		return v
	case int:
		return ID(v)
	case int64:
		return ID(v)
	case float64:
		return ID(v)
	case json.Number:
		i, _ := strconv.ParseInt(v.String(), 10, 64)
		return ID(i)
	default:
		return 0
	}
}

type OdooErrorPayload struct {
	Name      string         `json:"name"`
	Message   string         `json:"message"`
	Arguments []any          `json:"arguments"`
	Context   map[string]any `json:"context"`
	Debug     string         `json:"debug"`
}

type OdooHTTPError struct {
	Model      string
	Method     string
	StatusCode int
	Status     string
	Payload    *OdooErrorPayload
	RawBody    string
}

func (e *OdooHTTPError) Error() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Odoo JSON-2 request failed\n")
	fmt.Fprintf(&b, "  operation: %s.%s\n", e.Model, e.Method)
	fmt.Fprintf(&b, "  status:    %d %s\n", e.StatusCode, e.Status)

	if e.Payload != nil {
		if e.Payload.Name != "" {
			fmt.Fprintf(&b, "  exception: %s\n", e.Payload.Name)
		}
		if e.Payload.Message != "" {
			fmt.Fprintf(&b, "  message:   %s\n", e.Payload.Message)
		}
	} else if e.RawBody != "" {
		fmt.Fprintf(&b, "  body:      %s\n", e.RawBody)
	}

	return strings.TrimRight(b.String(), "\n")
}
