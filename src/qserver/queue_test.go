package main

import (
	"flag"
	"fmt"
	"sync"
	"testing"
)

var (
	q *queue
	data []string

	count = flag.Int("count", 10000, "number of concurrent read/writes")
)

func contains(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

func TestSingleGoroutine(t *testing.T) {
	q = newQueue()

	got1 := "abcdef"
	got2 := "ghijkl"

	q.enqueue([]byte(got1))
	q.enqueue([]byte(got2))

	if want, ok := q.dequeue(); !ok || string(want) != got1 {
		t.Errorf("want %q, got %q", want, got1)
	}

	if want, ok := q.dequeue(); !ok || string(want) != got2 {
		t.Errorf("want %q, got %q", want, got2)
	}
}

func TestWrite(t *testing.T) {
	var waitgroup sync.WaitGroup
	for i := 0; i < *count; i++ {
		s := fmt.Sprintf("%d", i)
		data = append(data, s)
		waitgroup.Add(1)
		go func() {
			q.enqueue([]byte(s))
			waitgroup.Done()
		}()
	}
	waitgroup.Wait()
}

func BenchmarkWrite(b *testing.B) {
	bmData := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		bmData[i] = fmt.Sprintf("%d", i)
	}
	var waitgroup sync.WaitGroup
	waitgroup.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func(i int) {
			q.enqueue([]byte(bmData[i]))
			waitgroup.Done()
		}(i)
	}
	waitgroup.Wait()
}

func TestRead(t *testing.T) {
	var waitgroup sync.WaitGroup

	var results []*string
	for i := 0; i < *count; i++ {
		s := new(string)
		results = append(results, s)

		waitgroup.Add(1)
		go func(s *string) {
			b, ok := q.dequeue()
			if !ok {
				t.Errorf("Unexpected empty queue")
				return
			}
			*s = string(b)
			waitgroup.Done()
		}(s)
	}

	waitgroup.Wait()
	for i := 0; i < *count; i++ {
		s := *(results[i])
		if !contains(data, *(results[i])) {
			t.Errorf("%q was not enqueue", s)
			return
		}
	}

	if _, ok := q.dequeue(); ok {
		t.Errorf("Expected empty queue")
	}
}

/* TODO
func BenchmarkRead(b *testing.B) {
	bmData := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		bmData[i] = fmt.Sprintf("%d", i)
		q.enqueue([]byte(bmData[i]))
	}

	var waitgroup sync.WaitGroup
	waitgroup.Add(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			q.dequeue()
			waitgroup.Done()
		}()
	}
	waitgroup.Wait()
}
*/

func TestReadWrite(t *testing.T) {
	enq := make(chan int)
	deq := make(chan int)

	for i := 0; i < 2 * *count; i++ {
		go func(i int) {
			if i % 2 == 0 {
				q.enqueue([]byte(data[i/2]))
				enq <- 1
				if (i/2)+1 == *count {
					enq <- -1
				}
			} else {
				if _, ok := q.dequeue(); ok {
					deq <- 1
				} else {
					deq <- 0
				}
				if (i/2)+1 == *count {
					deq <- -1
				}
			}
		}(i)
	}

	var enqCount, deqCount int
	enqEnd, deqEnd := false, false

	for {
		if enqEnd && deqEnd {
			break
		}

		select {
		case i := <-enq:
			if i < 0 {
				enqEnd = true
			} else {
				enqCount += i
			}
		case i := <-deq:
			if i < 0 {
				deqEnd = true
			} else {
				deqCount += i
			}
		}
	}

	if enqCount != *count {
		t.Errorf("%d enqueue operations failed", *count - enqCount)
		return
	}

	if deqCount != *count {
		t.Errorf("%d dequeue operations failed", *count - deqCount)
		return
	}

	if _, ok := q.dequeue(); ok {
		t.Errorf("expected empty queue")
		return
	}
}

func BenchmarkReadWrite(b *testing.B) {
	bmData := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		bmData[i] = fmt.Sprintf("%d", i)
	}
	b.ResetTimer()
	for i := 0; i < 2 * b.N; i++ {
		go func(i int) {
			if i % 2 == 0 {
				q.enqueue([]byte(bmData[i/2]))
			} else {
				q.dequeue()
			}
		}(i)
	}

}

