package main

import "fmt"

//heap struct that hold the array
type maxHeap struct {
	array []int
}

//insert
func (h *maxHeap) insert(key int) {
	h.array = append(h.array, key)
	fmt.Println("Length of array:", len(h.array))
	h.maxHeapifyUp(len(h.array) - 1)
}

func (h *maxHeap) maxHeapifyUp(index int) {
	for h.array[h.getParentIndex(index)] < h.array[index] {
		h.swap(h.getParentIndex(index), index)
		index = h.getParentIndex(index)
	}
}

//get parent index
func (h *maxHeap) getParentIndex(index int) int {
	return (index - 1) / 2
}

//right child index always be even number
func (h *maxHeap) rightIndex(index int) int {
	return 2*index + 2
}

//left child index always be odd number
func (h *maxHeap) leftIndex(index int) int {
	return 2*index + 1
}

func (h *maxHeap) swap(index1, index2 int) {
	h.array[index1], h.array[index2] = h.array[index2], h.array[index1]
}

func main() {
	m := &maxHeap{}

	buildHeap := []int{50, 30, 20, 40, 10, 60, 70}
	for _, v := range buildHeap {
		m.insert(v)
		fmt.Println(m)
	}
}
