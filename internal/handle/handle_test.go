package handle

import (
	"strings"
	"testing"
)

func TestGenerateShouldCreateValidThreePartHandle(t *testing.T) {
	gen := NewGenerator()

	handle, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	parts := strings.Split(handle, "-")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d: %s", len(parts), handle)
	}

	if len(handle) == 0 {
		t.Error("Generated empty handle")
	}
}

func TestGenerateShouldGenerateMostlyUniqueHandles(t *testing.T) {
	gen := NewGenerator()
	seen := make(map[string]bool)

	// Generate multiple handles and check for variety
	for i := 0; i < 50; i++ {
		handle, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() failed on iteration %d: %v", i, err)
		}
		seen[handle] = true
	}

	// With ~100 words per list, we should see mostly unique handles
	if len(seen) < 40 {
		t.Errorf("Expected at least 40 unique handles, got %d", len(seen))
	}
}

func TestGenerateUniqueShouldGenerateUniqueHandlesUsingExistsFunction(t *testing.T) {
	gen := NewGenerator()
	used := make(map[string]bool)

	exists := func(h string) bool {
		return used[h]
	}

	// Generate a few unique handles
	for i := 0; i < 5; i++ {
		handle, err := gen.GenerateUnique(exists)
		if err != nil {
			t.Fatalf("GenerateUnique() failed: %v", err)
		}

		if used[handle] {
			t.Errorf("Generated duplicate handle: %s", handle)
		}

		used[handle] = true
	}
}

func TestGenerateUniqueShouldReturnErrorWhenAllHandlesExist(t *testing.T) {
	gen := NewGenerator()

	// Make exists always return true to force exhaustion
	exists := func(h string) bool {
		return true
	}

	_, err := gen.GenerateUnique(exists)
	if err == nil {
		t.Error("Expected error when all handles exist, got nil")
	}
}
