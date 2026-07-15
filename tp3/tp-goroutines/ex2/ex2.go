package main

import (
	"fmt"
	"sync"
	"time"
)

func afficherLettre2(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("A")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("B")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("C")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("D")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("E")
}

func afficherChiffre2(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("1")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("2")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("3")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("4")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("5")
}

func main() {
	var wg sync.WaitGroup

	wg.Add(2)
	go afficherLettre2(&wg)
	go afficherChiffre2(&wg)

	wg.Wait()
}
