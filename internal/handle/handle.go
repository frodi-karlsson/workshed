package handle

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type GeneratorOption func(*Generator)

func WithAdjectives(adjs []string) GeneratorOption {
	return func(g *Generator) {
		g.adjectives = adjs
	}
}

func WithNouns(ns []string) GeneratorOption {
	return func(g *Generator) {
		g.nouns = ns
	}
}

func WithVerbs(vs []string) GeneratorOption {
	return func(g *Generator) {
		g.verbs = vs
	}
}

// Generator creates random handles in the format adjective-noun-verb
type Generator struct {
	adjectives []string
	nouns      []string
	verbs      []string
}

// NewGenerator creates a new handle generator with default word lists.
func NewGenerator(opts ...GeneratorOption) *Generator {
	g := &Generator{
		adjectives: adjectives,
		nouns:      nouns,
		verbs:      verbs,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Generate creates a random handle in the format adjective-noun-verb.
func (g *Generator) Generate() (string, error) {
	adj, err := g.randomWord(g.adjectives)
	if err != nil {
		return "", fmt.Errorf("selecting adjective: %w", err)
	}

	noun, err := g.randomWord(g.nouns)
	if err != nil {
		return "", fmt.Errorf("selecting noun: %w", err)
	}

	verb, err := g.randomWord(g.verbs)
	if err != nil {
		return "", fmt.Errorf("selecting verb: %w", err)
	}

	return fmt.Sprintf("%s-%s-%s", adj, noun, verb), nil
}

// GenerateUnique creates a unique handle by retrying until the exists check returns false
func (g *Generator) GenerateUnique(exists func(string) bool) (string, error) {
	const maxAttempts = 100

	for i := 0; i < maxAttempts; i++ {
		handle, err := g.Generate()
		if err != nil {
			return "", err
		}

		if !exists(handle) {
			return handle, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique handle after %d attempts", maxAttempts)
}

func (g *Generator) randomWord(words []string) (string, error) {
	if len(words) == 0 {
		return "", fmt.Errorf("word list is empty")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
	if err != nil {
		return "", fmt.Errorf("generating random number: %w", err)
	}

	return words[n.Int64()], nil
}
