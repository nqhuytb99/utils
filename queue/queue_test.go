package queue

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

var noOfMessage = 128

func TestCombination(t *testing.T) {
	testContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := NewQueue[string](testContext, WithSizeLimit(32))
	input := make(chan string)
	go func() {
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()
		defer close(input)

		count := 0
		for range ticker.C {
			count++
			input <- "Hello"
			if count >= noOfMessage {
				break
			}
		}

		log.Println("Done sending", count)
	}()

	go func() {
		var count int
		for value := range input {
			count++
			err := q.Enqueue(value)
			if err != nil {
				log.Println("Error:", err)
				t.Fail()
			}
		}
		log.Println("Done enqueue", count)
		cancel()
	}()

	var count int
	for data := range q.Receive() {
		count += len(data)
		t.Log("Received:", len(data))
		for _, d := range data {
			if d != "Hello" {
				t.Fail()
			}
		}
		fmt.Println("Received:", len(data), "Total:", count)
	}

	if count != noOfMessage {
		log.Println("Expected:", noOfMessage, "Received:", count)
		t.Fail()
	}
}
