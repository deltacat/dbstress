package fieldset

import "strings"

// GenerateFieldSet ...
func GenerateFieldSet(s string) ([]string, []string, []string) {
	ints := []string{}
	floats := []string{}
	strs := []string{}

	parts := strings.Split(s, ",")

	for _, part := range parts {
		if strings.HasSuffix(part, "i") {
			ints = append(ints, strings.Split(part, "=")[0])
			continue
		}
		if strings.HasSuffix(part, "str") {
			strs = append(strs, strings.Split(part, "=")[0])
			continue
		}
		floats = append(floats, strings.Split(part, "=")[0])
	}

	return ints, floats, strs
}

// GenerateTagsSet ...
func GenerateTagsSet(tagsTmpl string) [][]string {
	parts := strings.Split(tagsTmpl, ",")
	tags := [][]string{}
	for _, part := range parts {
		tags = append(tags, strings.Split(part, "="))
	}
	return tags
}
