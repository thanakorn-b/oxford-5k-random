package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Sense is one part-of-speech line with CEFR level (B2 / C1).
type sense struct {
	PartOfSpeech string `json:"part_of_speech"`
	Level        string `json:"level"`
}

// Entry is one headword in the Oxford list JSON.
type entry struct {
	Index  int     `json:"index"`
	Word   string  `json:"word"`
	Senses []sense `json:"senses"`
}

type doc struct {
	Entries []entry `json:"entries"`
}

func formatSensesLine(e entry) string {
	parts := make([]string, 0, len(e.Senses))
	for _, s := range e.Senses {
		parts = append(parts, fmt.Sprintf("%s (%s)", s.PartOfSpeech, s.Level))
	}
	return strings.Join(parts, ", ")
}

func formatEntry(e entry) string {
	return e.Word + "\n" + formatSensesLine(e)
}

// loadLexicon reads the word list JSON and assigns stable slice indices.
func loadLexicon(path string) (*doc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d doc
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	if len(d.Entries) == 0 {
		return nil, fmt.Errorf("lexicon has no entries")
	}
	for i := range d.Entries {
		d.Entries[i].Index = i
	}
	return &d, nil
}
