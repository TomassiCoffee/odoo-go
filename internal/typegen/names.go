package typegen

import (
	"hash/fnv"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var wordRegexp = regexp.MustCompile(`[A-Za-z0-9]+`)

func ToPascalCase(value string) string {
	words := wordRegexp.FindAllString(value, -1)
	if len(words) == 0 {
		return "Unknown"
	}

	var builder strings.Builder
	for _, word := range words {
		if word == "" {
			continue
		}
		runes := []rune(strings.ToLower(word))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		builder.WriteString(string(runes))
	}

	name := builder.String()
	if name == "" {
		return "Unknown"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		return "_" + name
	}
	return name
}

func StableHash(value string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	alphabet := "0123456789abcdefghijklmnopqrstuvwxyz"
	n := h.Sum32()
	if n == 0 {
		return "0"
	}
	var out []byte
	for n > 0 {
		out = append([]byte{alphabet[n%36]}, out...)
		n /= 36
	}
	if len(out) > 6 {
		out = out[:6]
	}
	return string(out)
}

func ModelTypeNames(models []NormalizedModel) map[string]string {
	counts := map[string]int{}
	for _, model := range models {
		counts[ToPascalCase(model.Name)]++
	}

	out := map[string]string{}
	for _, model := range models {
		base := ToPascalCase(model.Name)
		if counts[base] > 1 {
			base += "_" + StableHash(model.Name)
		}
		out[model.Name] = base
	}
	return out
}

func FieldGoNames(fields []NormalizedField) map[string]string {
	baseByField := map[string]string{}
	counts := map[string]int{}
	for _, field := range fields {
		base := ToPascalCase(field.Name)
		baseByField[field.Name] = base
		counts[base]++
	}

	names := map[string]string{}
	keys := make([]string, 0, len(baseByField))
	for key := range baseByField {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, fieldName := range keys {
		base := baseByField[fieldName]
		if counts[base] > 1 {
			base += "_" + StableHash(fieldName)
		}
		names[fieldName] = base
	}
	return names
}
