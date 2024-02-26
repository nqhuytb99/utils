package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

// Node represents a node in the trie
type Node struct {
	value    interface{} // Optional value to store at the end of a word
	children map[rune]*Node
	mutex    sync.Mutex // Mutex for thread safety
}

// Trie represents the trie data structure
type Trie struct {
	root *Node
}

// NewTrie creates a new trie
func NewTrie() *Trie {
	return &Trie{root: &Node{children: make(map[rune]*Node)}}
}

// Insert inserts a key-value pair into the trie
func (t *Trie) Insert(key string, value interface{}) {
	t.root.mutex.Lock()
	defer t.root.mutex.Unlock()

	node := t.root
	for _, r := range key {
		if _, ok := node.children[r]; !ok {
			node.children[r] = &Node{children: make(map[rune]*Node)}
		}
		node = node.children[r]
	}
	node.value = value
}

// Get retrieves the value associated with a key from the trie
func (t *Trie) Get(key string) interface{} {
	t.root.mutex.Lock()
	defer t.root.mutex.Unlock()

	node := t.root
	for _, r := range key {
		if _, ok := node.children[r]; !ok {
			return nil
		}
		node = node.children[r]
	}
	return node.value
}

// PrefixSearch searches for keys starting with a given prefix
func (t *Trie) PrefixSearch(prefix string) []string {
	t.root.mutex.Lock()
	defer t.root.mutex.Unlock()

	node := t.root
	for _, r := range prefix {
		if _, ok := node.children[r]; !ok {
			return nil
		}
		node = node.children[r]
	}

	var results []string
	dfs(node, prefix, &results)
	return results
}

func dfs(node *Node, current string, results *[]string) {
	if node.value != nil {
		*results = append(*results, current)
	}
	for r, child := range node.children {
		dfs(child, current+string(r), results)
	}
}

func keyFromString(key string) (hashKey uint64) {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func main() {
	trie := NewTrie()
	trie.Insert("apple", "apple value")
	trie.Insert("appliance", "appliance value")
	trie.Insert("banana", "banana value")

	fmt.Println(trie.Get("apple"))        // Output: apple value
	fmt.Println(trie.Get("ban"))          // Output: nil
	fmt.Println(trie.PrefixSearch("app")) // Output: [apple appliance]
}
