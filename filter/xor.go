package filter

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/FastFilter/xorfilter"
)

type Filter struct {
	xorfilter *xorfilter.BinaryFuse8
}

func HashText(text string) uint64 {
	hash := uint64(0)
	for _, char := range text {
		// Convert character to ASCII code and fold 8 bits into 1
		foldedChar := uint64(char) & 0xFF
		hash = hash<<5 + hash + foldedChar
		// Apply bitwise AND to ensure 64-bit output
		hash &= 0xFFFFFFFFFFFFFFFF
	}
	return hash
}

func NewFilter() *Filter {
	return &Filter{}
}

// BuildFilter creates a new xorfilter.BinaryFuse8 from a list of keys.
//
// Caution: This method can use a large amount of RAM. Consider using BuildFilterFromHashes instead.
func (f *Filter) BuildFilter(keys []string) {
	hashes := make([]uint64, len(keys))
	for i, key := range keys {
		hashes[i] = HashText(key)
	}
	filter, _ := xorfilter.PopulateBinaryFuse8(hashes)
	f.xorfilter = filter
}

func (f *Filter) BuildFilterFromHashes(hashes []uint64) {
	filter, _ := xorfilter.PopulateBinaryFuse8(hashes)
	f.xorfilter = filter
}

func (f *Filter) Contains(key string) bool {
	if f.xorfilter == nil {
		return false
	}
	return f.xorfilter.Contains(HashText(key))
}

func (f *Filter) SaveToFile(fp string) error {
	// Encode the struct to a byte buffer
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(*f.xorfilter)
	if err != nil {
		return fmt.Errorf("encoding data: %s", err.Error())
	}

	// Write the byte buffer to a file
	err = os.WriteFile(fp, buffer.Bytes(), 0o644)
	if err != nil {
		return fmt.Errorf("writing to file: %s", err.Error())
	}

	return nil
}

func (f *Filter) LoadFromFile(fp string) error {
	// Read the file into a byte buffer
	file, err := os.OpenFile(fp, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening file: %s", err.Error())
	}
	defer file.Close()

	// Decode the byte buffer into a struct
	decoder := gob.NewDecoder(file)
	filter := &xorfilter.BinaryFuse8{}
	err = decoder.Decode(filter)
	if err != nil {
		return fmt.Errorf("decoding data: %s", err.Error())
	}

	f.xorfilter = filter
	return nil
}
