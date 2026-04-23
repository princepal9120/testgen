package llm

import "testing"

func TestCacheGenerateKeyDeterministicAndSensitiveToInputs(t *testing.T) {
	t.Parallel()

	cache := NewCache(2)
	base := cache.GenerateKey("prompt", "system", "model-a")

	if got := cache.GenerateKey("prompt", "system", "model-a"); got != base {
		t.Fatalf("expected deterministic key, got %q want %q", got, base)
	}
	if got := cache.GenerateKey("prompt-changed", "system", "model-a"); got == base {
		t.Fatal("expected prompt change to alter cache key")
	}
	if got := cache.GenerateKey("prompt", "system-changed", "model-a"); got == base {
		t.Fatal("expected system role change to alter cache key")
	}
	if got := cache.GenerateKey("prompt", "system", "model-b"); got == base {
		t.Fatal("expected model change to alter cache key")
	}
}

func TestCacheGetReturnsCachedCopyAndTracksStats(t *testing.T) {
	t.Parallel()

	cache := NewCache(2)
	key := cache.GenerateKey("prompt", "system", "model")
	original := &CompletionResponse{
		Content:      "generated test",
		TokensInput:  123,
		TokensOutput: 45,
		Model:        "model",
	}
	cache.Set(key, original)

	if _, hit := cache.Get("missing"); hit {
		t.Fatal("expected cache miss for unknown key")
	}

	cached, hit := cache.Get(key)
	if !hit {
		t.Fatal("expected cache hit")
	}
	if !cached.Cached {
		t.Fatal("expected returned response to be marked as cached")
	}
	if cached.Content != original.Content || cached.TokensInput != original.TokensInput || cached.TokensOutput != original.TokensOutput {
		t.Fatalf("unexpected cached response: %#v", cached)
	}
	if original.Cached {
		t.Fatal("expected original response to remain unchanged")
	}

	size, hits, misses, hitRate := cache.Stats()
	if size != 1 {
		t.Fatalf("expected cache size 1, got %d", size)
	}
	if hits != 1 || misses != 1 {
		t.Fatalf("expected 1 hit / 1 miss, got hits=%d misses=%d", hits, misses)
	}
	if hitRate != 0.5 {
		t.Fatalf("expected hit rate 0.5, got %f", hitRate)
	}
}

func TestCacheRespectsMaxSize(t *testing.T) {
	t.Parallel()

	cache := NewCache(1)
	keyA := cache.GenerateKey("prompt-a", "system", "model")
	keyB := cache.GenerateKey("prompt-b", "system", "model")
	cache.Set(keyA, &CompletionResponse{Content: "A"})
	cache.Set(keyB, &CompletionResponse{Content: "B"})

	size, _, _, _ := cache.Stats()
	if size != 1 {
		t.Fatalf("expected cache size to remain capped at 1, got %d", size)
	}
	_, hitA := cache.Get(keyA)
	_, hitB := cache.Get(keyB)
	if hitA == hitB {
		t.Fatalf("expected exactly one cached entry after eviction, got hitA=%v hitB=%v", hitA, hitB)
	}
}
