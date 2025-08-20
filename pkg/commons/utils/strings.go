package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func Sensitive(s string, mask string, left int, right int) string {
	if len(s) <= left+right {
		return strings.Repeat(mask, len(s))
	}
	return s[:left] + mask + s[len(s)-right:]
}

// GenerateUUID creates a random UUID-like string
func GenerateUUID() string {
	return uuid.New().String()
}

// Idents append idents
func Idents(s string, idents int) string {
	var lines []string
	for line := range strings.Lines(s) {
		lines = append(lines, fmt.Sprintf("%s%s", strings.Repeat(" ", idents), line))
	}
	return strings.Join(lines, "")
}

func HashString(s string) string {
	return strings.ReplaceAll(uuid.NewSHA1(uuid.NameSpaceDNS, []byte(s)).String(), "-", "")
}

func JSONString(obj interface{}, indent int) string {
	s, _ := json.MarshalIndent(obj, "", strings.Repeat(" ", indent))
	return string(s)
}

func ContainString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func WrapString(s, r string) string {
	return fmt.Sprintf("%s\n%s\n%s", r, s, r)
}
