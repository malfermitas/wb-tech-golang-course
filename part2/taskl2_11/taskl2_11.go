package main

import (
	"fmt"
	"sort"
	"strings"
)

func groupAnagrams(words []string) map[string][]string {
	anagramGroups := make(map[string][]string)

	for _, word := range words {
		lowerWord := strings.ToLower(word)
		sortedWord := sortLettersInString(lowerWord)
		anagramGroups[sortedWord] = append(anagramGroups[sortedWord], word)
	}

	result := make(map[string][]string)
	for _, group := range anagramGroups {
		if len(group) > 1 {
			sort.Strings(group)
			result[group[0]] = group
		}
	}

	return result
}

func sortLettersInString(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	result := groupAnagrams(words)

	for key, group := range result {
		fmt.Printf("\"%s\": %v\n", key, group)
	}
}
