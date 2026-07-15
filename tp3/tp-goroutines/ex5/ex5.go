package main

import (
	"fmt"
	"sync"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		results <- j * j
		fmt.Printf("worker %d processed job %d\n", id, j)
	}
}

func main() {
	const numJobs = 20
	const numWorkers = 4

	jobs := make(chan int)
	results := make(chan int)

	var wg sync.WaitGroup

	wg.Add(numWorkers)
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		for j := 1; j <= numJobs; j++ {
			jobs <- j
		}
		close(jobs)
	}()

	for r := range results {
		fmt.Println("result:", r)
	}
}
