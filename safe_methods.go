package odoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Values map[string]any

type Method string

const (
	MethodSearchRead              Method = "search_read"
	MethodSearch                  Method = "search"
	MethodRead                    Method = "read"
	MethodCreate                  Method = "create"
	MethodWrite                   Method = "write"
	MethodUnlink                  Method = "unlink"
	MethodActionArchive           Method = "action_archive"
	MethodActionUnarchive         Method = "action_unarchive"
	MethodCheckFieldAccess        Method = "check_field_access"
	MethodCopy                    Method = "copy"
	MethodCopyData                Method = "copy_data"
	MethodCopyTranslations        Method = "copy_translations"
	MethodDefaultGet              Method = "default_get"
	MethodExportData              Method = "export_data"
	MethodFieldsGet               Method = "fields_get"
	MethodGetBaseURL              Method = "get_base_url"
	MethodGetExternalID           Method = "get_external_id"
	MethodGetFieldTranslations    Method = "get_field_translations"
	MethodGetMetadata             Method = "get_metadata"
	MethodGetPropertyDefinition   Method = "get_property_definition"
	MethodHasAccess               Method = "has_access"
	MethodHasFieldAccess          Method = "has_field_access"
	MethodLoad                    Method = "load"
	MethodNameCreate              Method = "name_create"
	MethodNameSearch              Method = "name_search"
	MethodOnchange                Method = "onchange"
	MethodSearchCount             Method = "search_count"
	MethodUpdateFieldTranslations Method = "update_field_translations"
)

func CallTyped[R any](
	ctx context.Context,
	caller Caller,
	model string,
	method Method,
	options Payload,
) (R, error) {
	var result R

	if err := caller.Call(ctx, model, method, options, &result); err != nil {
		return result, fmt.Errorf("%s.%s: %w", model, method, err)
	}

	return result, nil
}

func callVoid(
	ctx context.Context,
	caller Caller,
	model string,
	method Method,
	options Payload,
) error {
	_, err := CallTyped[json.RawMessage](
		ctx,
		caller,
		model,
		method,
		options,
	)
	return err
}

func SearchReadTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options SearchReadOptions,
) ([]T, error) {
	if options.Fields == nil {
		options.Fields = model.Fields()
	}

	return CallTyped[[]T](
		ctx,
		caller,
		model.Name(),
		MethodSearchRead,
		options,
	)
}

func SearchTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options SearchOptions,
) ([]ID, error) {
	return CallTyped[[]ID](
		ctx,
		caller,
		model.Name(),
		MethodSearch,
		options,
	)
}

func ReadTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options ReadOptions,
) ([]T, error) {
	if options.Fields == nil {
		options.Fields = model.Fields()
	}

	return CallTyped[[]T](
		ctx,
		caller,
		model.Name(),
		MethodRead,
		options,
	)
}

func CreateTyped[M any, V any](
	ctx context.Context,
	caller Caller,
	model Model[M],
	vals []V,
) ([]ID, error) {
	return CallTyped[[]ID](
		ctx,
		caller,
		model.Name(),
		MethodCreate,
		CreateOptions[V]{
			ValsList: vals,
		},
	)
}

func WriteTyped[M any, V any](
	ctx context.Context,
	caller Caller,
	model Model[M],
	ids []ID,
	vals V,
) (bool, error) {
	return CallTyped[bool](
		ctx,
		caller,
		model.Name(),
		MethodWrite,
		WriteOptions{
			IDs:  ids,
			Vals: vals,
		},
	)
}

func UnlinkTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	ids []ID,
) (bool, error) {
	return CallTyped[bool](
		ctx,
		caller,
		model.Name(),
		MethodUnlink,
		UnlinkOptions{
			IDs: ids,
		},
	)
}

func SearchCountTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options SearchCountOptions,
) (int, error) {
	return CallTyped[int](
		ctx,
		caller,
		model.Name(),
		MethodSearchCount,
		options,
	)
}

func ActionArchiveTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	ids []ID,
) error {
	return callVoid(
		ctx,
		caller,
		model.Name(),
		MethodActionArchive,
		ActionArchiveOptions{
			IDs: ids,
		},
	)
}

func ActionUnarchiveTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	ids []ID,
) error {
	return callVoid(
		ctx,
		caller,
		model.Name(),
		MethodActionUnarchive,
		ActionUnarchiveOptions{
			IDs: ids,
		},
	)
}

func CheckFieldAccessTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options CheckFieldAccessOptions,
) error {
	return callVoid(
		ctx,
		caller,
		model.Name(),
		MethodCheckFieldAccess,
		options,
	)
}

func HasAccessTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options HasAccessOptions,
) (bool, error) {
	return CallTyped[bool](
		ctx,
		caller,
		model.Name(),
		MethodHasAccess,
		options,
	)
}

func HasFieldAccessTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options HasFieldAccessOptions,
) (bool, error) {
	return CallTyped[bool](
		ctx,
		caller,
		model.Name(),
		MethodHasFieldAccess,
		options,
	)
}

func CopyTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options CopyOptions,
) ([]ID, error) {
	return CallTyped[[]ID](
		ctx,
		caller,
		model.Name(),
		MethodCopy,
		options,
	)
}

func CopyDataTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options CopyDataOptions,
) ([]T, error) {
	return CallTyped[[]T](
		ctx,
		caller,
		model.Name(),
		MethodCopyData,
		options,
	)
}

func DefaultGetTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options DefaultGetOptions,
) (Values, error) {
	return CallTyped[Values](
		ctx,
		caller,
		model.Name(),
		MethodDefaultGet,
		options,
	)
}

type ExportDataResult struct {
	Datas [][]any `json:"datas"`
}

func ExportDataTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options ExportDataOptions,
) (ExportDataResult, error) {
	return CallTyped[ExportDataResult](
		ctx,
		caller,
		model.Name(),
		MethodExportData,
		options,
	)
}

type FieldDefinition Values

func FieldsGetTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options FieldsGetOptions,
) (map[string]FieldDefinition, error) {
	return CallTyped[map[string]FieldDefinition](
		ctx,
		caller,
		model.Name(),
		MethodFieldsGet,
		options,
	)
}

func GetBaseURLTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	ids []ID,
) (string, error) {
	return CallTyped[string](
		ctx,
		caller,
		model.Name(),
		MethodGetBaseURL,
		GetBaseUrlOptions{
			IDs: ids,
		},
	)
}

func GetExternalIDTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	ids []ID,
) (map[ID]string, error) {
	return CallTyped[map[ID]string](
		ctx,
		caller,
		model.Name(),
		MethodGetExternalID,
		GetExternalIdOptions{
			IDs: ids,
		},
	)
}

type Metadata Values

func GetMetadataTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	ids []ID,
) ([]Metadata, error) {
	return CallTyped[[]Metadata](
		ctx,
		caller,
		model.Name(),
		MethodGetMetadata,
		GetMetadataOptions{
			IDs: ids,
		},
	)
}

func GetPropertyDefinitionTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options GetPropertyDefinitionOptions,
) (Values, error) {
	return CallTyped[Values](
		ctx,
		caller,
		model.Name(),
		MethodGetPropertyDefinition,
		options,
	)
}

type FieldTranslation struct {
	Lang   string `json:"lang"`
	Source string `json:"source"`
	Value  string `json:"value"`
}

type FieldTranslationsResult struct {
	Translations []FieldTranslation
	Context      map[string]any
}

func (r *FieldTranslationsResult) UnmarshalJSON(data []byte) error {
	var tuple []json.RawMessage

	if err := json.Unmarshal(data, &tuple); err != nil {
		return fmt.Errorf(
			"decode field translations tuple: %w",
			err,
		)
	}

	if len(tuple) != 2 {
		return fmt.Errorf(
			"decode field translations tuple: expected 2 elements, got %d",
			len(tuple),
		)
	}

	if err := json.Unmarshal(tuple[0], &r.Translations); err != nil {
		return fmt.Errorf(
			"decode field translations: %w",
			err,
		)
	}

	if err := json.Unmarshal(tuple[1], &r.Context); err != nil {
		return fmt.Errorf(
			"decode field translations context: %w",
			err,
		)
	}

	return nil
}

func GetFieldTranslationsTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options GetFieldTranslationsOptions,
) (FieldTranslationsResult, error) {
	return CallTyped[FieldTranslationsResult](
		ctx,
		caller,
		model.Name(),
		MethodGetFieldTranslations,
		options,
	)
}

type LoadIDs struct {
	IDs    []ID
	Failed bool
}

func (v *LoadIDs) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)

	if bytes.Equal(data, []byte("false")) {
		v.IDs = nil
		v.Failed = true
		return nil
	}

	var ids []ID

	if err := json.Unmarshal(data, &ids); err != nil {
		return fmt.Errorf(
			"expected ID array or false: %w",
			err,
		)
	}

	v.IDs = ids
	v.Failed = false

	return nil
}

type LoadResult struct {
	IDs LoadIDs `json:"ids"`

	// Runtime documentation says [Message] but does not define Message.
	Messages []json.RawMessage `json:"messages"`

	LastRow *int `json:"lastrow,omitempty"`
}

func LoadTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options LoadOptions,
) (LoadResult, error) {
	return CallTyped[LoadResult](
		ctx,
		caller,
		model.Name(),
		MethodLoad,
		options,
	)
}

type NamePair struct {
	ID   ID
	Name string
}

func (p *NamePair) UnmarshalJSON(data []byte) error {
	var tuple []json.RawMessage

	if err := json.Unmarshal(data, &tuple); err != nil {
		return fmt.Errorf("decode name pair: %w", err)
	}

	if len(tuple) != 2 {
		return fmt.Errorf(
			"decode name pair: expected 2 elements, got %d",
			len(tuple),
		)
	}

	if err := json.Unmarshal(tuple[0], &p.ID); err != nil {
		return fmt.Errorf(
			"decode name pair ID: %w",
			err,
		)
	}

	if err := json.Unmarshal(tuple[1], &p.Name); err != nil {
		return fmt.Errorf(
			"decode name pair name: %w",
			err,
		)
	}

	return nil
}

func NameCreateTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options NameCreateOptions,
) (*NamePair, error) {
	raw, err := CallTyped[json.RawMessage](
		ctx,
		caller,
		model.Name(),
		MethodNameCreate,
		options,
	)
	if err != nil {
		return nil, err
	}

	raw = bytes.TrimSpace(raw)

	if bytes.Equal(raw, []byte("false")) {
		return nil, nil
	}

	var result NamePair

	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf(
			"%s.%s: decode result: %w",
			model.Name(),
			MethodNameCreate,
			err,
		)
	}

	return &result, nil
}

func NameSearchTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options NameSearchOptions,
) ([]NamePair, error) {
	return CallTyped[[]NamePair](
		ctx,
		caller,
		model.Name(),
		MethodNameSearch,
		options,
	)
}

func OnchangeTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options OnchangeOptions,
) (Values, error) {
	return CallTyped[Values](
		ctx,
		caller,
		model.Name(),
		MethodOnchange,
		options,
	)
}

func UpdateFieldTranslationsTyped[T any](
	ctx context.Context,
	caller Caller,
	model Model[T],
	options UpdateFieldTranslationsOptions,
) (bool, error) {
	return CallTyped[bool](
		ctx,
		caller,
		model.Name(),
		MethodUpdateFieldTranslations,
		options,
	)
}
