package main

import (
	"os"
)

type Edge struct {
	src, dst, cost, id int
}

// -------------------- Fast IO -------------------

const Bufsize = 1 << 20

var inputBuffer [Bufsize]byte
var outputBuffer [Bufsize]byte
var inputPtr = 0
var ioPtrLim = Bufsize - 64
var bytesRead = 0
var outputPtr = 0
var stdin *os.File
var stdout *os.File

func fill() {
	n := bytesRead - inputPtr
	if inputPtr > 0 {
		for i := 0; i < n; i++ {
			inputBuffer[i] = inputBuffer[inputPtr]
			inputPtr++
		}
	}
	bytesRead -= inputPtr - n
	n, _ = stdin.Read(inputBuffer[bytesRead:])
	inputPtr = 0
	bytesRead += n
}

func flush() {
	stdout.Write(outputBuffer[:outputPtr])
	outputPtr = 0
}

func readInt() int {
	if inputPtr > ioPtrLim {
		fill()
	}
	var x = 0
	for inputBuffer[inputPtr] < '0' {
		inputPtr++
	}
	for inputBuffer[inputPtr] >= '0' {
		x = x*10 + int(inputBuffer[inputPtr]-'0')
		inputPtr++
	}
	return x
}

func writeInt(x int) {
	if outputPtr > ioPtrLim {
		flush()
	}
	if x == 0 {
		outputBuffer[outputPtr] = '0'
		outputPtr++
	} else {
		var buffer [20]byte
		i := 0
		for x > 0 {
			buffer[i] = byte(x%10 + '0')
			x /= 10
			i++
		}
		for i > 0 {
			i--
			outputBuffer[outputPtr] = buffer[i]
			outputPtr++
		}
	}
	outputBuffer[outputPtr] = ' '
	outputPtr++
}

// -------------------- SkewHeap -------------------

type SkewHeap struct {
	cost, id, offset int
	Left, Right      *SkewHeap
}

var HeapPool = []SkewHeap{}
var HeapPoolPtr = 0

func Merge(sh1, sh2 *SkewHeap) *SkewHeap {
	if sh1 == nil {
		return sh2
	}
	if sh2 == nil {
		return sh1
	}
	if sh1.offset != 0 {
		sh1.propagate()
	}
	if sh2.offset != 0 {
		sh2.propagate()
	}
	if sh1.cost > sh2.cost {
		sh1, sh2 = sh2, sh1
	}
	sh1.Right = Merge(sh1.Right, sh2)
	sh1.Left, sh1.Right = sh1.Right, sh1.Left
	return sh1
}

func Push(heap *SkewHeap, cost, id int) *SkewHeap {
	newNode := &HeapPool[HeapPoolPtr]
	HeapPoolPtr++
	newNode.cost = cost
	newNode.id = id
	return Merge(heap, newNode)
}

func Pop(heap *SkewHeap) *SkewHeap {
	return Merge(heap.Left, heap.Right)
}

func Update(heap *SkewHeap, offset int) {
	if heap == nil {
		return
	}
	heap.cost += offset
	heap.offset += offset
}

func (sh *SkewHeap) propagate() {
	if sh.Left != nil {
		sh.Left.cost += sh.offset
		sh.Left.offset += sh.offset
	}
	if sh.Right != nil {
		sh.Right.cost += sh.offset
		sh.Right.offset += sh.offset
	}
	sh.offset = 0
}

// -------------------- Solution -------------------

func Tarjan(edges []Edge, N, M, S int) (int, []int) {
	twoN := 2 * N
	ch := make(chan bool, 2)
	HeapPool = make([]SkewHeap, M+N-1)

	mins := make([]*SkewHeap, twoN)
	go func() {
		for i, e := range edges[:N] {
			mins[e.dst] = Push(mins[e.dst], e.cost, i)
			if i == S {
				continue
			}
			edges[M] = Edge{i, S, 0, -1}
			M++
		}
		for i := N; i < M; i++ {
			e := edges[i]
			mins[e.dst] = Push(mins[e.dst], e.cost, i)
		}
		ch <- true
	}()

	super := make([]int, twoN)
	parent := make([]int, twoN)
	visited := make([]int, twoN)
	go func() {
		for i := range parent {
			super[i] = i
			parent[i] = -1
			visited[i] = -1
		}
		ch <- true
	}()

	<-ch
	<-ch

	inEdge := func(v int) Edge {
		return edges[mins[v].id]
	}

	findPrev := func(v int) int {
		v = inEdge(v).src
		for super[v] != v {
			v, super[v] = super[v], super[super[v]]
		}
		return v
	}

	pathFront := 0

	for vc := N; mins[pathFront] != nil; vc++ {
		for visited[pathFront] == -1 {
			visited[pathFront] = 0
			pathFront = findPrev(pathFront)
		}
		for pathFront != vc {
			w := mins[pathFront].cost
			min := Pop(mins[pathFront])
			Update(min, -w)
			mins[vc] = Merge(mins[vc], min)
			parent[pathFront] = vc
			super[pathFront] = vc
			pathFront = findPrev(pathFront)
		}
		for mins[pathFront] != nil && findPrev(pathFront) == pathFront {
			mins[pathFront] = Pop(mins[pathFront])
		}
	}

	totalCost := 0
	for v := S; v != -1; v = parent[v] {
		visited[v] = 1
	}
	parent[S] = S
	for v := pathFront; v > -1; v-- {
		if visited[v] == 1 {
			continue
		}

		e := inEdge(v)
		dst := e.dst
		for dst != v {
			visited[dst] = 1
			dst = parent[dst]
		}
		totalCost += e.cost
		parent[e.dst] = e.src
	}
	return totalCost, parent[:N]
}

// -------------------- Main -------------------

func main() {
	stdin = os.Stdin
	stdout = os.Stdout

	fill()

	var n, m, s, src, dst, cost int
	n = readInt()
	m = readInt()
	s = readInt()

	edges := make([]Edge, m+n-1)
	for i := 0; i < m; i++ {
		src = readInt()
		dst = readInt()
		cost = readInt()
		edges[i] = Edge{src, dst, cost, i}
	}

	cost, parents := Tarjan(edges, n, m, s)

	writeInt(cost)
	outputBuffer[outputPtr-1] = '\n'
	for _, parent := range parents {
		writeInt(parent)
	}
	flush()
}
