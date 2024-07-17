package logger

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func PrettyMap(data map[string]interface{}, indent string) string {
	var builder strings.Builder
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	// Sort keys for consistent output
	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]
		builder.WriteString(fmt.Sprintf("%s%s: ", indent, k))

		if v == nil {
			builder.WriteString("nil\n")
			continue
		}

		switch reflect.TypeOf(v).Kind() {
		case reflect.Map:
			// If it's a map, recursively call PrettyLogMap
			subMap, ok := v.(map[string]interface{})
			if !ok {
				builder.WriteString(fmt.Sprintf("(Unsupported map type: %T)\n", v))
				continue
			}
			builder.WriteString("\n")
			builder.WriteString(PrettyMap(subMap, indent+"  "))
		case reflect.Slice, reflect.Array:
			// Handle slices and arrays
			builder.WriteString(fmt.Sprintf("%v\n", v))
		default:
			// For all other types, just print the value
			builder.WriteString(fmt.Sprintf("%v\n", v))
		}
	}
	return builder.String()
}
