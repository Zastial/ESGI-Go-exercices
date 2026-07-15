package main

import "fmt"

func sumSlice(slice []int, resultChan chan int) {
	sum := 0
	for _, num := range slice {
		sum += num
	}
	resultChan <- sum
}

func main() {
	slice := make([]int, 1000)
	for i := range 1000 {
		slice[i] = i + 1
	}

	resultChan := make(chan int)

	chunkSize := len(slice) / 4
	go sumSlice(slice[0:chunkSize], resultChan)
	go sumSlice(slice[chunkSize:2*chunkSize], resultChan)
	go sumSlice(slice[2*chunkSize:3*chunkSize], resultChan)
	go sumSlice(slice[3*chunkSize:], resultChan)

	totalSum := 0
	for range 4 {
		totalSum += <-resultChan
	}

	expected := len(slice) * (len(slice) + 1) / 2
	fmt.Println("Expected:", expected)
	fmt.Println("Got:", totalSum)
}
