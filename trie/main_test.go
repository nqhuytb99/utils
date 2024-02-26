package main

import (
	"testing"

	"github.com/google/uuid"
)

var trie = NewTrie()
var hashMap = map[string]interface{}{}
var hashMapWithHashKey = map[uint64]interface{}{}
var cacheSize = 100 * 1000

var sampleData struct {
}

func BenchmarkTrieSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < cacheSize; j++ {
			id := uuid.NewString()
			trie.Insert(id, sampleData)
		}
	}
}

func BenchmarkHashMapSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < cacheSize; j++ {
			id := uuid.NewString()
			hashMap[id] = sampleData
		}
	}
}

func BenchmarkHashMapWithHashKeySet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < cacheSize; j++ {
			id := uuid.NewString()
			hashMapWithHashKey[keyFromString(id)] = sampleData
		}
	}
}

func BenchmarkTrieGet(b *testing.B) {
	for j := 0; j < cacheSize; j++ {
		id := uuid.NewString()
		trie.Insert(id, sampleData)
	}
	for i := 0; i < cacheSize; i++ {
		trie.Get(uuid.NewString())
	}
}

func BenchmarkHashMapGet(b *testing.B) {
	for j := 0; j < cacheSize; j++ {
		id := uuid.NewString()
		hashMap[id] = sampleData
	}
	for i := 0; i < cacheSize; i++ {
		x := hashMap[uuid.NewString()]
		_ = x
	}
}
