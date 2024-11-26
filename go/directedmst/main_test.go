package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"math/rand/v2"

	"github.com/elordeiro/goext/containers/graph"
	"github.com/elordeiro/goext/containers/hashgraph"
	"github.com/elordeiro/goext/seqs"
)

var start, end time.Time

type UnionFind struct {
	parent []int
	rank   []int
}

func NewUnionFind(size int) *UnionFind {
	parent := make([]int, size)
	rank := make([]int, size)
	for i := 1; i < size; i++ {
		parent[i] = i
	}
	return &UnionFind{parent: parent, rank: rank}
}

func (uf *UnionFind) MakeSet(x int) {
	uf.parent = append(uf.parent, x)
	uf.rank = append(uf.rank, 0)
}

func (uf *UnionFind) Find(x int) int {
	for x != uf.parent[x] {
		x, uf.parent[x] = uf.parent[x], uf.parent[uf.parent[x]]
	}
	return x
}

func (uf *UnionFind) Union(x, y int) {
	rootX, rootY := uf.Find(x), uf.Find(y)
	if rootX == rootY {
		return
	}
	if uf.rank[rootX] < uf.rank[rootY] {
		rootX, rootY = rootY, rootX
	}
	uf.parent[rootY] = rootX
	if uf.rank[rootX] == uf.rank[rootY] {
		uf.rank[rootX]++
	}
}

func (uf *UnionFind) Connected(x, y int) bool {
	return uf.Find(x) == uf.Find(y)
}

func checker(cost, s int, parents []int, edges []Edge) error {
	uf := NewUnionFind(len(parents))
	edgeMap := map[[2]int]int{}

	for _, e := range edges {
		edge := [2]int{e.src, e.dst}
		cost := e.cost
		edgeMap[edge] = cost
	}

	total := 0
	for v, u := range parents {
		if v == s {
			if u != v {
				return fmt.Errorf("parent of root %d is %d", s, u)
			}
			continue
		}
		if uf.Connected(u, v) {
			return errors.New("output isn't a tree")
		}
		uf.Union(u, v)
		e := [2]int{u, v}
		weight, ok := edgeMap[e]
		if !ok {
			return fmt.Errorf("%d->%d isn't an edge", u, v)
		}
		total += weight
	}
	if total != cost {
		return fmt.Errorf("cost = %d, want %d", cost, total)
	}
	return nil
}

func mainRunner(edges []Edge, n, m, s int) (int, []int, error) {
	inputBuffer = [Bufsize]byte{}
	outputBuffer = [Bufsize]byte{}
	inputPtr = 0
	bytesRead = 0
	outputPtr = 0
	HeapPool = []SkewHeap{}
	HeapPoolPtr = 0

	stdin := os.Stdin
	stdout := os.Stdout

	in, _ := os.OpenFile("./stdin.txt", os.O_RDONLY|os.O_WRONLY|os.O_TRUNC, 0644)
	inWriter := bufio.NewWriter(in)
	fmt.Fprintf(inWriter, "%d %d %d\n", n, m, s)
	for _, e := range edges {
		fmt.Fprintf(inWriter, "%d %d %d\n", e.src, e.dst, e.cost)
	}
	inWriter.Flush()
	in.Close()

	in, _ = os.Open("./stdin.txt")
	os.Stdin = in

	out, _ := os.OpenFile("./stdout.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	os.Stdout = out

	{
		f, _ := os.Create("cpu.prof")
		pprof.StartCPUProfile(f)
		start = time.Now()
		main()
		end = time.Now()
		pprof.StopCPUProfile()
	}

	out.Close()
	out, _ = os.Open("./stdout.txt")
	outReader := bufio.NewReader(out)

	var cost int
	parents := make([]int, n)
	fmt.Fscan(outReader, &cost)
	for i := range n {
		fmt.Fscan(outReader, &parents[i])
	}

	os.Stdin = stdin
	os.Stdout = stdout

	err := checker(cost, s, parents, edges)
	return cost, parents, err
}

func GraphMaker(N_MAX int) (int, int, int, []Edge) {
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
	counter := 0
	uniform := func(min, max int) int {
		counter++
		if counter == 10000 {
			counter = 0
			r = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
		}
		return r.IntN(max-min+1) + min
	}

	for {
		C_MAX := 1000

		n := uniform(1, N_MAX)
		m := uniform(n-1, min(N_MAX, n*(n-1)))

		used := make(map[[2]int]bool)
		var edges []Edge

		addEdge := func(a, b, c int) bool {
			if used[[2]int{a, b}] {
				return false
			}
			used[[2]int{a, b}] = true
			edges = append(edges, Edge{a, b, c, 0})
			return true
		}

		idx := rand.Perm(n)
		s := idx[0]

		addEdge(s, uniform(0, n-1), C_MAX)
		addEdge(s, 0, C_MAX)

		for i := 1; i < n; i++ {
			addEdge(uniform(0, i-1), i, C_MAX)
		}

		for i := n - 1; i < m; i++ {
			var a, b int
			for range 1000 {
				a = uniform(0, n-1)
				b = uniform(0, n-1)
				if a != b && !used[[2]int{a, b}] {
					break
				}
			}
			c := uniform(0, C_MAX)
			addEdge(a, b, c)
		}

		m = len(edges)
		if m > N_MAX*(N_MAX-1) {
			continue
		}

		g := hashgraph.New[int, int](true)
		for _, e := range edges {
			g.AddEdge(e.src, e.dst, e.cost)
		}

		if seqs.Len(graph.BFS(g, s)) != n-1 {
			continue
		}

		return n, m, s, edges
	}
}

func TestTarjan1(t *testing.T) {
	testOnly := 12
	tests := []struct {
		n, m, s int
		edges   []Edge
		cost    int
		parents []int
	}{
		{
			// tc = 1
			4, 4, 0,
			[]Edge{
				{0, 1, 10, 0},
				{0, 2, 10, 0},
				{0, 3, 3, 0},
				{3, 2, 4, 0},
			},
			17, []int{0, 0, 3, 0},
		},
		{
			// tc = 2
			7, 8, 3,
			[]Edge{
				{3, 1, 10, 0},
				{1, 2, 1, 0},
				{2, 0, 1, 0},
				{0, 1, 1, 0},
				{2, 6, 10, 0},
				{6, 4, 1, 0},
				{4, 5, 1, 0},
				{5, 6, 1, 0},
			},
			24, []int{2, 3, 1, 3, 6, 4, 2},
		},
		{
			// tc = 3
			4, 4, 0,
			[]Edge{
				{0, 1, 1, 0},
				{1, 2, 1, 0},
				{2, 3, 1, 0},
				{3, 1, 1, 0},
			},
			3, []int{0, 0, 1, 2},
		},
		{
			// tc = 4
			7, 9, 6,
			[]Edge{
				{1, 2, 1, 0},
				{2, 3, 2, 0},
				{3, 0, 1, 0},
				{3, 1, 1, 0},
				{3, 4, 1, 0},
				{4, 5, 2, 0},
				{5, 3, 1, 0},
				{2, 5, 1, 0},
				{6, 2, 1, 0},
			},
			6, []int{3, 3, 6, 5, 3, 2, 6},
		},
		{
			// tc = 5
			8, 11, 0,
			[]Edge{
				{0, 1, 1, 0},
				{1, 2, 2, 0},
				{1, 3, 1, 0},
				{2, 3, 1, 0},
				{4, 2, 1, 0},
				{3, 4, 1, 0},
				{3, 7, 1, 0},
				{4, 5, 1, 0},
				{5, 6, 1, 0},
				{6, 4, 1, 0},
				{6, 0, 1, 0},
				{7, 2, 1, 0},
			},
			7, []int{0, 0, 4, 1, 3, 4, 5, 3},
		},
		{
			// tc = 6
			8, 11, 0,
			[]Edge{
				{0, 1, 1, -1},
				{1, 2, 1, -1},
				{1, 5, 1, -1},
				{2, 3, 2, -1},
				{3, 4, 1, -1},
				{4, 1, 1, -1},
				{4, 6, 2, -1},
				{5, 7, 1, -1},
				{6, 5, 1, -1},
				{7, 3, 1, -1},
				{7, 6, 1, -1},
			},
			7, []int{0, 0, 1, 7, 3, 1, 7, 5},
		},
		{
			// tc = 7
			7, 10, 0,
			[]Edge{
				{0, 1, 1, 0},
				{1, 2, 1, 0},
				{2, 3, 1, 0},
				{3, 1, 1, 0},
				{4, 5, 1, 0},
				{5, 6, 1, 0},
				{6, 4, 1, 0},
				{3, 4, 1, 0},
				{4, 2, 1, 0},
				{5, 2, 1, 0},
			},
			6, []int{0, 0, 1, 2, 3, 4, 5},
		},
		{
			// tc = 8
			10, 21, 3,
			[]Edge{
				{3, 4, 1, 0},
				{9, 0, 1, 0},
				{0, 1, 1, 0},
				{1, 2, 1, 0},
				{2, 4, 1, 0},
				{4, 5, 1, 0},
				{1, 6, 1, 0},
				{1, 7, 1, 0},
				{4, 8, 1, 0},
				{7, 9, 1, 0},
				{9, 2, 1, 0},
				{2, 8, 1, 0},
				{0, 4, 1, 0},
				{2, 9, 2, 0},
				{2, 6, 1, 0},
				{4, 1, 1, 0},
				{0, 5, 1, 0},
				{7, 5, 1, 0},
				{6, 5, 1, 0},
				{6, 9, 2, 0},
				{0, 6, 1, 0},
			},
			9, []int{9, 4, 1, 3, 3, 4, 1, 1, 4, 7},
		},
		{
			// tc = 9
			5, 6, 0,
			[]Edge{
				{0, 1, 1, 0},
				{4, 2, 1, 0},
				{4, 1, 1, 0},
				{1, 2, 1, 0},
				{2, 3, 1, 0},
				{3, 4, 1, 0},
			},
			4, []int{0, 0, 1, 2, 3},
		},
		{
			// tc = 10
			6, 7, 4,
			[]Edge{
				{0, 1, 1, 0},
				{1, 2, 1, 0},
				{1, 5, 1, 0},
				{2, 3, 1, 0},
				{3, 0, 1, 0},
				{5, 3, 1, 0},
				{4, 1, 2, 0},
			},
			6, []int{3, 4, 1, 5, 4, 1},
		},
	}

	for i, tc := range tests {
		if testOnly > 0 && testOnly != i+1 {
			continue
		}
		cost, parents, err := mainRunner(tc.edges, tc.n, tc.m, tc.s)
		if err == nil {
			continue
		}
		t.Error(err)
		t.Error(cost)
		t.Errorf("tc = %d", i+1)
		str := "\nparents = \n"
		for v, u := range parents {
			str += fmt.Sprintf("%d>%d\n", u, v)
		}
		str += "want =\n"
		for v, u := range tc.parents {
			str += fmt.Sprintf("%d>%d\n", u, v)
		}
		t.Error(str)
	}

}

func TestTarjan2(t *testing.T) {
	edges := []Edge{
		{4, 7, 10, -1},
		{4, 0, 10, -1},
		{0, 1, 10, -1},
		{0, 2, 10, -1},
		{1, 3, 10, -1},
		{2, 5, 10, -1},
		{3, 6, 10, -1},
		{0, 7, 10, -1},
		{6, 8, 10, -1},
		{6, 9, 10, -1},
		{0, 10, 10, -1},
		{10, 11, 10, -1},
		{7, 12, 10, -1},
		{11, 13, 10, -1},
		{10, 14, 10, -1},
		{8, 5, 0, -1},
		{1, 11, 8, -1},
		{5, 11, 2, -1},
		{3, 14, 9, -1},
		{8, 14, 5, -1},
		{8, 6, 7, -1},
		{9, 5, 6, -1},
		{3, 11, 10, -1},
		{7, 14, 6, -1},
		{13, 3, 3, -1},
	}

	cost, parents, err := mainRunner(edges, 15, 25, 4)
	if err == nil {
		return
	}
	t.Error(err)
	fmt.Println(cost)
	for v, u := range parents {
		fmt.Printf("%d>%d\n", u, v)
	}
}

func TestTarjan3(t *testing.T) {
	edges := []Edge{
		{8, 9, 10, -1},
		{8, 0, 10, -1},
		{0, 1, 10, -1},
		{1, 2, 10, -1},
		{2, 3, 10, -1},
		{0, 4, 10, -1},
		{4, 5, 10, -1},
		{2, 6, 10, -1},
		{6, 7, 10, -1},
		{7, 9, 10, -1},
		{9, 10, 10, -1},
		{10, 11, 10, -1},
		{4, 12, 10, -1},
		{12, 13, 10, -1},
		{7, 14, 10, -1},
		{1, 15, 10, -1},
		{4, 16, 10, -1},
		{1, 17, 10, -1},
		{3, 18, 10, -1},
		{2, 19, 10, -1},
		{11, 20, 10, -1},
		{15, 4, 10, -1},
		{13, 16, 10, -1},
		{16, 9, 10, -1},
		{14, 13, 2, -1},
		{18, 10, 8, -1},
		{20, 16, 8, -1},
		{6, 3, 3, -1},
		{19, 3, 6, -1},
		{14, 2, 1, -1},
		{13, 14, 8, -1},
	}

	cost, parents, err := mainRunner(edges, 21, 31, 8)
	if err == nil {
		return
	}
	t.Error(err)
	fmt.Println(cost)
	for v, u := range parents {
		fmt.Printf("%d>%d\n", u, v)
	}
}

func TestTarjan4(t *testing.T) {
	edges := []Edge{
		{5, 4, 10, -1},
		{5, 0, 10, -1},
		{0, 1, 10, -1},
		{0, 2, 10, -1},
		{2, 3, 10, -1},
		{3, 4, 10, -1},
		{4, 3, 2, -1},
		{0, 4, 7, -1},
		{3, 1, 3, -1},
		{2, 0, 6, -1},
		{4, 2, 10, -1},
	}

	cost, parents, err := mainRunner(edges, 6, 11, 5)
	if err == nil {
		return
	}
	t.Error(err)
	fmt.Println(cost)
	for v, u := range parents {
		fmt.Printf("%d>%d\n", u, v)
	}
}

func TestTarjan5(t *testing.T) {
	edges := []Edge{
		{1, 5, 10, -1},
		{1, 0, 10, -1},
		{0, 2, 10, -1},
		{1, 3, 10, -1},
		{2, 4, 10, -1},
		{3, 5, 10, -1},
		{5, 6, 10, -1},
		{0, 5, 7, -1},
		{2, 6, 9, -1},
		{5, 0, 2, -1},
		{6, 5, 1, -1},
		{4, 6, 7, -1},
		{1, 2, 2, -1},
		{0, 3, 7, -1},
		{4, 2, 6, -1},
		{0, 6, 3, -1},
		{4, 0, 0, -1},
		{4, 5, 10, -1},
		{2, 5, 2, -1},
		{5, 3, 7, -1},
		{6, 3, 8, -1},
	}

	cost, parents, err := mainRunner(edges, 7, 21, 1)
	if err == nil {
		return
	}
	t.Error(err)
	fmt.Println(cost)
	for v, u := range parents {
		fmt.Printf("%d>%d\n", u, v)
	}
}

func TestTarjan6(t *testing.T) {
	edges := []Edge{
		{11, 10, 10, 0},
		{11, 0, 10, 0},
		{0, 1, 10, 0},
		{1, 2, 10, 0},
		{2, 3, 10, 0},
		{3, 4, 10, 0},
		{3, 5, 10, 0},
		{3, 6, 10, 0},
		{6, 7, 10, 0},
		{0, 8, 10, 0},
		{0, 9, 10, 0},
		{9, 10, 10, 0},
		{11, 7, 0, 0},
		{5, 4, 7, 0},
		{0, 3, 6, 0},
		{2, 5, 1, 0},
		{9, 2, 7, 0},
		{8, 3, 3, 0},
		{9, 0, 1, 0},
		{5, 6, 10, 0},
		{4, 9, 6, 0},
	}

	cost, parents, err := mainRunner(edges, 12, 21, 11)
	if err == nil {
		return
	}
	t.Error(err)
	fmt.Println(cost)
	for v, u := range parents {
		fmt.Printf("%d>%d\n", u, v)
	}
}

func TestTarjan7(t *testing.T) {
	edges := []Edge{
		{9, 0, 0, 0},
		{1, 0, 8, 0},
		{5, 0, 10, 0},
		{2, 1, 4, 0},
		{0, 1, 10, 0},
		{4, 2, 0, 0},
		{7, 2, 7, 0},
		{0, 2, 10, 0},
		{2, 3, 10, 0},
		{7, 4, 2, 0},
		{8, 4, 8, 0},
		{1, 4, 10, 0},
		{8, 6, 4, 0},
		{3, 6, 4, 0},
		{0, 6, 7, 0},
		{4, 6, 10, 0},
		{5, 6, 10, 0},
		{2, 7, 10, 0},
		{6, 8, 1, 0},
		{10, 8, 3, 0},
		{3, 8, 10, 0},
		{11, 9, 1, 0},
		{1, 9, 3, 0},
		{8, 9, 10, 0},
		{0, 10, 0, 0},
		{2, 10, 4, 0},
		{3, 10, 10, 0},
		{0, 11, 10, 0},
	}

	cost, parents, err := mainRunner(edges, 12, 28, 5)
	if err == nil {
		return
	}

	t.Error(err)
	fmt.Println(cost)
	for v, u := range parents {
		fmt.Printf("%d>%d\n", u, v)
	}
}

func TestTarjan8(t *testing.T) {
	edges := []Edge{
		{10, 9, 100, 0},
		{10, 0, 100, 0},
		{0, 1, 100, 0},
		{1, 2, 100, 0},
		{2, 3, 100, 0},
		{1, 4, 100, 0},
		{3, 5, 100, 0},
		{0, 6, 100, 0},
		{5, 7, 100, 0},
		{1, 8, 100, 0},
		{1, 9, 100, 0},
		{1, 10, 100, 0},
		{5, 3, 24, 0},
		{6, 4, 8, 0},
		{10, 2, 46, 0},
		{5, 10, 39, 0},
		{7, 0, 56, 0},
		{0, 5, 57, 0},
		{6, 3, 9, 0},
		{7, 9, 21, 0},
		{8, 4, 11, 0},
		{6, 1, 70, 0},
		{8, 0, 55, 0},
		{9, 10, 44, 0},
		{2, 5, 93, 0},
		{6, 5, 20, 0},
		{9, 8, 36, 0},
		{9, 7, 66, 0},
		{8, 5, 74, 0},
		{10, 5, 56, 0},
		{0, 4, 7, 0},
		{5, 4, 1, 0},
		{8, 1, 33, 0},
		{10, 6, 36, 0},
		{0, 8, 9, 0},
		{9, 1, 54, 0},
		{6, 9, 34, 0},
		{10, 4, 77, 0},
		{4, 8, 0, 0},
		{6, 2, 83, 0},
		{1, 0, 85, 0},
		{3, 0, 62, 0},
		{2, 6, 22, 0},
		{9, 3, 24, 0},
		{9, 2, 39, 0},
		{1, 3, 42, 0},
		{1, 7, 59, 0},
		{10, 8, 94, 0},
		{1, 5, 24, 0},
		{5, 8, 63, 0},
		{3, 7, 22, 0},
		{8, 2, 67, 0},
	}

	cost, parents, err := mainRunner(edges, 11, 52, 10)
	if err == nil {
		return
	}

	t.Error(err)
	fmt.Println(cost)
	for v, u := range parents {
		fmt.Printf("%d>%d\n", u, v)
	}
}

func TestMany(t *testing.T) {
	for i := range 100 {
		n, m, s, edges := GraphMaker(10000)

		printFailing := func() {
			for _, e := range edges {
				fmt.Printf("%d>%d\n", e.src, e.dst)
			}
			for _, e := range edges {
				fmt.Printf("{%d, %d, %d, 0},\n", e.src, e.dst, e.cost)
			}
		}

		defer func() {
			if r := recover(); r != nil {
				t.Error(r)
				printFailing()
			}
		}()

		var cost int
		var parents []int
		var err error
		done := make(chan bool, 1)
		go func() {
			fmt.Printf("Graph made: %d %d %d\n", n, m, s)
			cost, parents, err = mainRunner(edges, n, m, s)
			done <- true
		}()

		select {
		case <-time.After(30 * time.Second):
			printFailing()
			panic("timeout")
		case <-done:
			if err == nil {
				fmt.Println("Pass", i, "time:", end.UnixMicro()-start.UnixMicro())
				continue
			}
			t.Error(err)
			fmt.Println("Fail", i)
			fmt.Printf("cost: %d\n", cost)
			for v, u := range parents {
				fmt.Printf("%d>%d\n", u, v)
			}
			printFailing()
		}
	}
}

func TestGraphMaker(t *testing.T) {
	n, m, s, edges := GraphMaker(200000)
	cost, parents, err := mainRunner(edges, n, m, s)
	if err == nil {
		return
	}
	t.Error(err)
	t.Error(cost)
	t.Error(parents)
}

func BenchmarkLarge(b *testing.B) {
	stdin := os.Stdin
	stdout := os.Stdout

	in, _ := os.Open("large.txt")
	os.Stdin = in
	defer in.Close()

	out, _ := os.OpenFile("stdout.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	os.Stdout = out
	defer out.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		main()
		in.Seek(0, 0)
		out.Seek(0, 0)

		inputBuffer = [Bufsize]byte{}
		outputBuffer = [Bufsize]byte{}
		HeapPool = []SkewHeap{}
		HeapPoolPtr = 0
		inputPtr = 0
		bytesRead = 0
		outputPtr = 0
	}

	os.Stdin = stdin
	os.Stdout = stdout
}
