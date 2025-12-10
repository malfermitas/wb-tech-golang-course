package main

func moveZeroes(nums []int) {
	count := 0 // Count of zeroes

	// Traverse through the array elements.
	for i, num := range nums {
		if num == 0 {
			count++
		} else if count > 0 {
			// Move non-zero to current position - count places
			nums[i-count] = num

			// And place zero at the end of array
			nums[i] = 0
		}
	}
}

func main() {
	moveZeroes([]int{1, 0, 7, 0, 9, 0, 12, 14, 15, 0})
}
