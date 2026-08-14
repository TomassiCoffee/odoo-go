package typegen

import (
	"context"
	"sort"

	"github.com/TomassiCoffee/odoo-go"
)

type staticModel[T any] struct {
	name   string
	fields []string
}

func (m staticModel[T]) Name() string     { return m.name }
func (m staticModel[T]) Fields() []string { return m.fields }

type irModelRecord struct {
	ID    int64            `json:"id"`
	Model odoo.FalseString `json:"model"`
	Name  odoo.FalseString `json:"name"`
}

type irModelFieldRecord struct {
	Name             odoo.FalseString `json:"name"`
	FieldDescription odoo.FalseString `json:"field_description"`
	Model            odoo.FalseString `json:"model"`
	ModelID          odoo.Many2One    `json:"model_id"`
	Relation         odoo.FalseString `json:"relation"`
	Required         bool             `json:"required"`
	Readonly         bool             `json:"readonly"`
	TType            odoo.FalseString `json:"ttype"`
}

func Fetch(ctx context.Context, client *odoo.Client, pageSize int, modelNames []string) (MetadataCache, error) {
	modelDomain := odoo.Domain{}
	if len(modelNames) > 0 {
		modelDomain = odoo.Domain{odoo.Clause("model", "in", modelNames)}
	}

	models, err := readAllTyped[irModelRecord](ctx, client, staticModel[irModelRecord]{
		name:   "ir.model",
		fields: []string{"id", "model", "name"},
	}, odoo.SearchReadOptions{
		Domain: modelDomain,
		Order:  "model asc",
	}, pageSize)
	if err != nil {
		return MetadataCache{}, err
	}

	modelIDs := make([]any, 0, len(models))
	for _, model := range models {
		if model.ID != 0 {
			modelIDs = append(modelIDs, model.ID)
		}
	}
	if len(modelIDs) == 0 {
		return MetadataCache{Models: nil, FieldCount: 0, Source: "odoo"}, nil
	}

	fields, err := readAllTyped[irModelFieldRecord](ctx, client, staticModel[irModelFieldRecord]{
		name: "ir.model.fields",
		fields: []string{
			"name",
			"field_description",
			"model",
			"model_id",
			"relation",
			"required",
			"readonly",
			"ttype",
		},
	}, odoo.SearchReadOptions{
		Domain: odoo.Domain{odoo.Clause("model_id", "in", modelIDs)},
		Order:  "model asc, name asc",
	}, pageSize)
	if err != nil {
		return MetadataCache{}, err
	}

	fieldsByModel := map[string][]irModelFieldRecord{}
	for _, field := range fields {
		modelName := field.Model.StringOr("")
		if modelName == "" && field.ModelID.Valid {
			modelName = field.ModelID.DisplayName
		}
		if modelName == "" {
			continue
		}
		fieldsByModel[modelName] = append(fieldsByModel[modelName], field)
	}

	return MetadataCache{
		Models:     normalizeModels(models, fieldsByModel),
		FieldCount: len(fields),
		Source:     "odoo",
	}, nil
}

func readAllTyped[T any](ctx context.Context, client *odoo.Client, model odoo.Model[T], base odoo.SearchReadOptions, pageSize int) ([]T, error) {
	if pageSize <= 0 {
		pageSize = 500
	}

	var out []T
	for offset := 0; ; offset += pageSize {
		limit := pageSize
		currentOffset := offset
		base.Limit = &limit
		base.Offset = &currentOffset
		page, err := odoo.SearchReadTyped[T](ctx, client, model, base)
		if err != nil {
			return nil, err
		}
		out = append(out, page...)
		if len(page) < pageSize {
			return out, nil
		}
	}
}

func normalizeModels(models []irModelRecord, fieldsByModel map[string][]irModelFieldRecord) []NormalizedModel {
	out := make([]NormalizedModel, 0, len(models)+1)

	for _, model := range models {
		modelName := model.Model.StringOr("")
		if modelName == "" {
			continue
		}

		fields := make([]NormalizedField, 0, len(fieldsByModel[modelName]))
		for _, field := range fieldsByModel[modelName] {
			fieldName := field.Name.StringOr("")
			fieldType := SupportedFieldType(field.TType.StringOr(""))
			if fieldName == "" || !supportedFieldTypes[fieldType] {
				continue
			}

			fields = append(fields, NormalizedField{
				Name:     fieldName,
				Type:     fieldType,
				Label:    field.FieldDescription.StringOr(""),
				Relation: field.Relation.StringOr(""),
				Required: field.Required,
				Readonly: field.Readonly,
			})
		}

		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		if len(fields) > 0 {
			out = append(out, NormalizedModel{Name: modelName, Fields: fields})
		}
	}

	// out = append(out, NormalizedModel{
	// 	Name: "_unknown",
	// 	Fields: []NormalizedField{
	// 		{Name: "id", Type: FieldInteger, Label: "ID", Readonly: true},
	// 		{Name: "display_name", Type: FieldChar, Label: "Display Name", Readonly: true},
	// 	},
	// })

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
