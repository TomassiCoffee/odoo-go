package odoo

import (
	"context"
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

type Payload interface {
	isPayload()
}

type Caller interface {
	Call(
		ctx context.Context,
		model string,
		method Method,
		options Payload,
		out any,
	) error
}

func (o SearchReadOptions) isPayload() {}

type SearchReadOptions struct {
	Domain  Domain         `json:"domain,omitempty"`
	Fields  []string       `json:"fields,omitempty"`
	Limit   *int           `json:"limit,omitempty"`
	Offset  *int           `json:"offset,omitempty"`
	Order   string         `json:"order,omitempty"`
	Context map[string]any `json:"context,omitempty"`
}

func (o SearchOptions) isPayload() {}

type SearchOptions struct {
	Domain Domain `json:"domain"`
	Limit  *int   `json:"limit,omitempty"`
	Offset *int   `json:"offset,omitempty"`
	Order  string `json:"order,omitempty"`
}

func (o ReadOptions) isPayload() {}

type ReadOptions struct {
	IDs    []ID     `json:"ids"`
	Fields []string `json:"fields,omitempty"`

	// Odoo default: "_classic_read"
	//
	// Optional, but exposing it lets the caller override the default.
	Load *string `json:"load,omitempty"`
}

func (o CreateOptions[V]) isPayload() {}

type CreateOptions[V any] struct {
	// Required.
	ValsList []V `json:"vals_list"`
}

func (o UnlinkOptions) isPayload() {}

type UnlinkOptions struct {
	// Required JSON/2 recordset selector.
	IDs []ID `json:"ids"`
}

func (o ActionArchiveOptions) isPayload() {}

type ActionArchiveOptions struct {
	// Required JSON/2 recordset selector.
	IDs []ID `json:"ids"`
}

func (o ActionUnarchiveOptions) isPayload() {}

type ActionUnarchiveOptions struct {
	// Required JSON/2 recordset selector.
	IDs []ID `json:"ids"`
}

type SupportedFieldOps string

const (
	CheckFieldAccessRead  SupportedFieldOps = "read"
	CheckFieldAccessWrite SupportedFieldOps = "write"
)

func (o CheckFieldAccessOptions) isPayload() {}

type CheckFieldAccessOptions struct {
	Field     string            `json:"field"`
	Operation SupportedFieldOps `json:"operation"`
}

func (o CopyOptions) isPayload() {}

type CopyOptions struct {
	// copy() operates on a recordset.
	IDs []ID `json:"ids"`

	// Odoo default: null
	Default *any `json:"default,omitempty"`
}

func (o CopyDataOptions) isPayload() {}

type CopyDataOptions struct {
	// copy_data() operates on a recordset.
	IDs []ID `json:"ids"`

	// Odoo default: null
	Default *any `json:"default,omitempty"`
}

func (o DefaultGetOptions) isPayload() {}

type DefaultGetOptions struct {
	// Required.
	Fields []string `json:"fields"`
}

func (o ExportDataOptions) isPayload() {}

type ExportDataOptions struct {
	// export_data() operates on a recordset.
	IDs []ID `json:"ids"`

	// Required.
	FieldsToExport []string `json:"fields_to_export"`
}

func (o FieldsGetOptions) isPayload() {}

type FieldsGetOptions struct {
	// Odoo names this "allfields", not "all_fields".
	// Default: null
	AllFields *[]string `json:"allfields,omitempty"`

	// Default: null
	Attributes *[]string `json:"attributes,omitempty"`
}

func (o GetBaseUrlOptions) isPayload() {}

type GetBaseUrlOptions struct {
	IDs []ID `json:"ids"`
}

func (o GetExternalIdOptions) isPayload() {}

type GetExternalIdOptions struct {
	IDs []ID `json:"ids"`
}

func (o GetFieldTranslationsOptions) isPayload() {}

type GetFieldTranslationsOptions struct {
	// get_field_translations() operates on a recordset.
	IDs []ID `json:"ids"`

	// Documentation says str, not []string.
	FieldName string `json:"field_name"`

	// Default: null
	Langs []string `json:"langs,omitempty"`
}

func (o GetMetadataOptions) isPayload() {}

type GetMetadataOptions struct {
	IDs []ID `json:"ids"`
}

func (o GetPropertyDefinitionOptions) isPayload() {}

type GetPropertyDefinitionOptions struct {
	FullName string `json:"full_name"`
}

func (o HasAccessOptions) isPayload() {}

type HasAccessOptions struct {
	IDs []ID `json:"ids"`

	// Required.
	Operation string `json:"operation"`
}

func (o HasFieldAccessOptions) isPayload() {}

type HasFieldAccessOptions struct {
	Field string `json:"field"`

	// Documentation:
	// Literal["read", "write"]
	Operation SupportedFieldOps `json:"operation"`
}

func (o LoadOptions) isPayload() {}

type LoadOptions struct {
	// Both arguments are required according to the runtime documentation.
	Data   any      `json:"data"`
	Fields []string `json:"fields"`
}

func (o NameCreateOptions) isPayload() {}

type NameCreateOptions struct {
	Name string `json:"name"`
}

func (o NameSearchOptions) isPayload() {}

type NameSearchOptions struct {
	// Default: null
	Domain *Domain `json:"domain,omitempty"`

	// Default: 100
	Limit *int `json:"limit,omitempty"`

	// Default: ""
	Name *string `json:"name,omitempty"`

	// Default: "ilike"
	Operator *string `json:"operator,omitempty"`
}

func (o OnchangeOptions) isPayload() {}

type OnchangeOptions struct {
	// Required JSON/2 recordset selector.
	IDs []ID `json:"ids"`

	// All three method arguments are required.
	FieldNames []string       `json:"field_names"`
	FieldsSpec map[string]any `json:"fields_spec"`
	Values     map[string]any `json:"values"`
}

func (o SearchCountOptions) isPayload() {}

type SearchCountOptions struct {
	// Required.
	Domain Domain `json:"domain"`

	// Default: null
	Limit *int `json:"limit,omitempty"`
}

func (o UpdateFieldTranslationsOptions) isPayload() {}

type UpdateFieldTranslationsOptions struct {
	// Required JSON/2 recordset selector.
	IDs []ID `json:"ids"`

	// Required.
	FieldName string `json:"field_name"`

	// Default: ""
	SourceLang *string `json:"source_lang,omitempty"`

	// Required.
	Translations map[string]any `json:"translations"`
}

func (o WriteOptions) isPayload() {}

type WriteOptions struct {
	IDs  []ID `json:"ids"`
	Vals any  `json:"vals"`
}

type Model[T any] interface {
	Name() string
	Fields() []string
	RecordType() *T
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
	Method     Method
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
