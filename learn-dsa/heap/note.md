# Data Structure: Heaps

## 1. What is a Heap?

A **Heap** is a specialized tree-based data structure that satisfies the **Heap Property**. It is commonly implemented as a **Complete Binary Tree**.

- **Complete Binary Tree:** A binary tree in which all the levels are completely filled except possibly the lowest one, which is filled from the left.

## 2. Types of Heaps

### Max-Heap

- **Property:** The value of the **Root** node must be the **greatest** among all its descendants. The same is true for all sub-trees (Parent $\ge$ Children).
- **Root:** Contains the maximum element.

### Min-Heap

- **Property:** The value of the **Root** node must be the **smallest** among all its descendants. The same is true for all sub-trees (Parent $\le$ Children).
- **Root:** Contains the minimum element.

---

## 3. Common Misconception: Sibling Order

**Question:** Is the left child always larger than the right child?
**Answer:** **NO.**

Unlike a Binary Search Tree (BST), where `Left < Parent < Right`, a Heap **does not enforce any order between the left and right children**.

As long as the parent is larger than _both_ children (in a Max-Heap), the heap is valid.

**Example of Valid Max-Heaps:**

**Case A (Left > Right):**

```text
      10
     /  \
    8    6
```

````

_Valid because 10 8 and 10 6._

**Case B (Right > Left):**

```text
      10
     /  \
    6    8

```

_Also Valid because 10 6 and 10 8._

---

## 4. Array Representation

Heaps are efficient because they can be stored as arrays without pointers.

For a node at index `i`:

- **Parent:** `(i - 1) / 2`
- **Left Child:** `(2 * i) + 1`
- **Right Child:** `(2 * i) + 2`

## 5. Time Complexity

| Operation          | Time Complexity | Description                                                                  |
| ------------------ | --------------- | ---------------------------------------------------------------------------- |
| **Get Max/Min**    |                 | Access the root element.                                                     |
| **Insert**         |                 | Add to end, then "bubble up" (heapify up).                                   |
| **Delete Max/Min** |                 | Swap root with last element, remove last, then "bubble down" (heapify down). |

## 6. Applications

1. **Priority Queues:** Efficiently retrieving the highest or lowest priority task.
2. **Heap Sort:** An efficient sorting algorithm.
3. **Graph Algorithms:** Used in Dijkstra’s Shortest Path and Prim’s Minimum Spanning Tree.

```
````
