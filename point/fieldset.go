package point

import "strings"

// TODO: add correct error handling/panic when appropriate
func generateFieldSet(s string) ([]string, []string, []string) {
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
