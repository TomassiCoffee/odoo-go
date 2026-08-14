package typegen

type SupportedFieldType string

const (
	FieldBinary    SupportedFieldType = "binary"
	FieldBoolean   SupportedFieldType = "boolean"
	FieldChar      SupportedFieldType = "char"
	FieldDate      SupportedFieldType = "date"
	FieldDatetime  SupportedFieldType = "datetime"
	FieldFloat     SupportedFieldType = "float"
	FieldHTML      SupportedFieldType = "html"
	FieldInteger   SupportedFieldType = "integer"
	FieldMany2Many SupportedFieldType = "many2many"
	FieldMany2One  SupportedFieldType = "many2one"
	FieldMonetary  SupportedFieldType = "monetary"
	FieldOne2Many  SupportedFieldType = "one2many"
	FieldSelection SupportedFieldType = "selection"
	FieldText      SupportedFieldType = "text"
)

var supportedFieldTypes = map[SupportedFieldType]bool{
	FieldBinary:    true,
	FieldBoolean:   true,
	FieldChar:      true,
	FieldDate:      true,
	FieldDatetime:  true,
	FieldFloat:     true,
	FieldHTML:      true,
	FieldInteger:   true,
	FieldMany2Many: true,
	FieldMany2One:  true,
	FieldMonetary:  true,
	FieldOne2Many:  true,
	FieldSelection: true,
	FieldText:      true,
}

type NormalizedField struct {
	Name     string             `json:"name"`
	Type     SupportedFieldType `json:"type"`
	Label    string             `json:"label,omitempty"`
	Relation string             `json:"relation,omitempty"`
	Required bool               `json:"required,omitempty"`
	Readonly bool               `json:"readonly,omitempty"`
}

type NormalizedModel struct {
	Name   string            `json:"name"`
	Fields []NormalizedField `json:"fields"`
}

type MetadataCache struct {
	Models     []NormalizedModel `json:"models"`
	FieldCount int               `json:"field_count"`
	Source     string            `json:"source"`
}

// FilterCache returns only the requested Odoo models. An empty model list returns cache unchanged.
func FilterCache(cache MetadataCache, modelNames []string) MetadataCache {
	if len(modelNames) == 0 {
		return cache
	}
	wanted := make(map[string]struct{}, len(modelNames))
	for _, name := range modelNames {
		if name != "" {
			wanted[name] = struct{}{}
		}
	}
	filtered := make([]NormalizedModel, 0, len(wanted))
	fieldCount := 0
	for _, model := range cache.Models {
		if _, ok := wanted[model.Name]; !ok {
			continue
		}
		filtered = append(filtered, model)
		fieldCount += len(model.Fields)
	}
	cache.Models = filtered
	cache.FieldCount = fieldCount
	return cache
}
