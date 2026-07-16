package store

import (
	"strings"
	"testing"
)

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple words",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with punctuation",
			input:    "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with numbers",
			input:    "go 123 test",
			expected: []string{"go", "123", "test"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitWords(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d words, got %d", len(tt.expected), len(result))
				return
			}
			for i, word := range result {
				if word != tt.expected[i] {
					t.Errorf("expected %q at index %d, got %q", tt.expected[i], i, word)
				}
			}
		})
	}
}

func TestCountVowels(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "hello",
			input:    "hello",
			expected: 2,
		},
		{
			name:     "aeiou",
			input:    "aeiou",
			expected: 5,
		},
		{
			name:     "no vowels",
			input:    "bcdfg",
			expected: 0,
		},
		{
			name:     "mixed case",
			input:    "HeLLo",
			expected: 2,
		},
		{
			name:     "with y",
			input:    "yay",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countVowels(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d vowels, got %d", tt.expected, result)
			}
		})
	}
}

func TestCountDigits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "123",
			input:    "123",
			expected: 3,
		},
		{
			name:     "hello world",
			input:    "hello world",
			expected: 0,
		},
		{
			name:     "mixed",
			input:    "abc123def",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countDigits(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d digits, got %d", tt.expected, result)
			}
		})
	}
}

func TestCountUppercase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "ABC",
			input:    "ABC",
			expected: 3,
		},
		{
			name:     "abc",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "mixed",
			input:    "AaBbCc",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countUppercase(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d uppercase, got %d", tt.expected, result)
			}
		})
	}
}

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "no overlap",
			a:        []string{"go", "rust"},
			b:        []string{"python", "java"},
			expected: []string{"go", "java", "python", "rust"},
		},
		{
			name:     "with overlap",
			a:        []string{"go", "rust"},
			b:        []string{"go", "python"},
			expected: []string{"go", "python", "rust"},
		},
		{
			name:     "with duplicates",
			a:        []string{"go", "go"},
			b:        []string{"rust", "rust"},
			expected: []string{"go", "rust"},
		},
		{
			name:     "with spaces",
			a:        []string{" go ", "rust "},
			b:        []string{" python"},
			expected: []string{"go", "python", "rust"},
		},
		{
			name:     "case insensitive",
			a:        []string{"Go", "RUST"},
			b:        []string{"go", "rust"},
			expected: []string{"go", "rust"},
		},
		{
			name:     "empty strings",
			a:        []string{"", "go"},
			b:        []string{"  ", "python"},
			expected: []string{"go", "python"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeTags(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tags, got %d", len(tt.expected), len(result))
				return
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("expected %q at index %d, got %q", tt.expected[i], i, tag)
				}
			}
		})
	}
}

func TestNullString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "non-empty string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUniqueWords(t *testing.T) {
	words := []string{"go", "rust", "go", "python", "go"}
	result := uniqueWords(words)

	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("expected %d unique words, got %d", expectedCount, len(result))
	}

	if _, ok := result["go"]; !ok {
		t.Error("expected 'go' in unique words")
	}
	if _, ok := result["rust"]; !ok {
		t.Error("expected 'rust' in unique words")
	}
	if _, ok := result["python"]; !ok {
		t.Error("expected 'python' in unique words")
	}
}

func TestQueryEmbedding(t *testing.T) {
	result := queryEmbedding("hello world")

	if !strings.HasPrefix(result, "[") || !strings.HasSuffix(result, "]") {
		t.Errorf("expected embedding to be bracketed, got %q", result)
	}

	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(result, "["), "]"), ",")
	if len(parts) != 8 {
		t.Errorf("expected 8 features, got %d", len(parts))
	}

	result2 := queryEmbedding("hello world")
	if result != result2 {
		t.Error("expected deterministic results")
	}
}
