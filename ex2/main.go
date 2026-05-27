package main

import (
	"fmt"
	"sort"
)

var args = []string{"go", "api", "backend", "go", "rest", "go"}

func main() {
	argMap := make(map[string]int)
	for _, arg := range args {
		argMap[arg] = argMap[arg] + 1
	}

	type tagCount struct {
		tag   string
		count int
	}

	var tagCounts []tagCount
	for key, value := range argMap {
		tagCounts = append(tagCounts, tagCount{tag: key, count: value})
	}

	sort.Slice(tagCounts, func(i, j int) bool {
		return tagCounts[i].count < tagCounts[j].count
	})

	fmt.Println("Tags triés par fréquence décroissante")
	for _, tagCountValue := range tagCounts {
		fmt.Println(tagCountValue.tag, tagCountValue.count)
	}

	fmt.Println("-----------------------------------------------------------")

	fmt.Println("Tags apparaissant plus d'1 fois")
	for _, tagCountValue := range tagCounts {
		if tagCountValue.count > 1 {
			fmt.Println(tagCountValue.tag, tagCountValue.count)
		}
	}
}
