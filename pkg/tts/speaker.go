package tts

import (
	"regexp"
	"strings"
)

type Speaker interface {
	Connect(string) error
	Close()
	Speech(text string) ([]byte, error)
}

// RemoveEmoji removes the typical emoji characters from the input string.
func RemoveEmoji(str string) string {
	// Regex pattern to match most emoji characters
	emojiPattern := "[\U0001F600-\U0001F64F\U0001F300-\U0001F5FF\U0001F680-\U0001F6FF\U0001F700-\U0001F77F\U0001F780-\U0001F7FF\U0001F800-\U0001F8FF\U0001F900-\U0001F9FF\U00002702-\U000027B0\U000024C2-\U0001F251]+"
	re := regexp.MustCompile(emojiPattern)
	// Replace matched emoji with an empty string to remove it
	return re.ReplaceAllString(str, "")
}

// RemoveExtraStrings removes any of the removal strings from the input string.
func RemoveExtraStrings(str string, remove []string) string {
	result := str
	for _, s := range remove {
		result = strings.ReplaceAll(result, s, "")
	}

	return result
}
