package main

import "fmt"

func main() {
	var sl = make([]int, 0, 3)

	sl = append(sl, 1)
	fmt.Printf("SL: %s, cap(): %s, len(): %s\n", sl, cap(sl), len(sl))

	sl = append(sl, 2)
	fmt.Printf("SL: %s, cap(): %s, len(): %s\n", sl, cap(sl), len(sl))

	sl = append(sl, 3)
	fmt.Printf("SL: %s, cap(): %s, len(): %s\n", sl, cap(sl), len(sl))

	sl = sl[:0]

	fmt.Printf("SL: %s, cap(): %s, len(): %s\n", sl, cap(sl), len(sl))

	sl = append(sl, 4)
	fmt.Printf("SL: %s, cap(): %s, len(): %s\n", sl, cap(sl), len(sl))

	sl = append(sl, 5)
	fmt.Printf("SL: %s, cap(): %s, len(): %s\n", sl, cap(sl), len(sl))
}
