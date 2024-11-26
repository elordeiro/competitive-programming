#include <stdio.h>
#include <stdlib.h>

// -------------------- Helper Functions -------------------

#define IN_EDGE(v) (edges[mins[(v)]->id])
#define FIND_PREV(v) ({ \
    int _v = v; \
    _v = IN_EDGE(_v).src; \
    while (super[_v] != _v) { \
        int temp = super[_v]; \
        super[_v] = super[super[_v]]; \
        _v = temp; \
    } \
    _v; \
})

// -------------------- Fast IO -------------------

const int BUFSIZE = 1 << 20;

char inputBuffer[BUFSIZE];
char outputBuffer[BUFSIZE];
int inputPtr = 0;
int bytesRead = 0;
int outputPtr = 0;
int ioPtrLimit = BUFSIZE - 64; 

void fill() {
	int n = bytesRead - inputPtr;
	if (inputPtr > 0 ) {
		for (int i = 0; i < n; i++) {
			inputBuffer[i] = inputBuffer[inputPtr++];
		}
	}
	bytesRead -= inputPtr - n;
	n = fread(inputBuffer + bytesRead, 1, sizeof(inputBuffer) - bytesRead, stdin);
	inputPtr = 0;
	bytesRead += n;
}

void flush() {
	while (outputPtr > 0) {
		outputPtr -= fwrite(outputBuffer, 1, outputPtr, stdout);
	}
}

int readInt() {
	if (inputPtr > ioPtrLimit) {
		fill();
	}
	int x = 0;
	while (inputBuffer[inputPtr] < '0') { 
		inputPtr++;
	}
	while (inputBuffer[inputPtr] >= '0') {
		x = x * 10 + inputBuffer[inputPtr] - '0';
		inputPtr++;
	}
	return x;
}

void writeLong(long x) {
	if (outputPtr > ioPtrLimit) {
		flush();
	}
	if (x == 0) {
		outputBuffer[outputPtr++] = '0';
	} else {
		char buffer[20];
		int i = 0;
		while (x > 0) {
			buffer[i++] = x % 10 + '0';
			x /= 10;
		}
		while (i > 0) {
			outputBuffer[outputPtr++] = buffer[--i];	
		}
	}
	outputBuffer[outputPtr++] = ' ';
}



// -------------------- Data Structures -------------------

typedef struct {
	int src, dst, cost, id;
} Edge;


typedef struct SkewHeap {
	int cost, id, offset;
	struct SkewHeap *Left, *Right; 
} SkewHeap;

// -------------------- SkewHeap -------------------

SkewHeap* POOL;
int POOL_INDEX = 0;

void propagate(SkewHeap* sh) {
	if (sh->Left != NULL) {
		sh->Left->cost += sh->offset;
		sh->Left->offset += sh->offset;
	}
	if (sh->Right != NULL) {
		sh->Right->cost += sh->offset;
		sh->Right->offset += sh->offset;
	}
	sh->offset = 0; 
}

SkewHeap* Merge(SkewHeap* sh1, SkewHeap* sh2)  {
	if (sh1 == NULL) {
		return sh2; 
	}
	if (sh2 == NULL) {
		return sh1;
	}
	if (sh1->offset != 0) {
		propagate(sh1);
	}
	if (sh2->offset != 0) {
		propagate(sh2);
	}
    SkewHeap* temp;
	if (sh1->cost > sh2->cost) {
		temp = sh1;
        sh1 = sh2;
        sh2 = temp;
	}
	sh1->Right = Merge(sh1->Right, sh2);
    temp = sh1->Left;
    sh1->Left = sh1->Right;
    sh1->Right = temp;
	return sh1;
}

SkewHeap* Push(SkewHeap* heap, int cost, int id) {
	SkewHeap* newNode = &POOL[POOL_INDEX++];
    newNode->cost = cost;
    newNode->id = id;
	return Merge(heap, newNode);
}

SkewHeap* Pop(SkewHeap* heap) {
    return Merge(heap->Left, heap->Right);
}

void Update(SkewHeap* heap, int offset) {
	if (heap == NULL) {
		return;
	}
	heap->cost += offset;
	heap->offset += offset;
}


// -------------------- Solution -------------------

long* Tarjan(Edge* edges, int n, int m, int s) {
	int N2 = 2 * n;
	int *super = (int*)malloc(N2 * sizeof(int));
    int *visited = (int*)malloc(N2 * sizeof(int));
    long *parent = (long*)malloc(N2 * sizeof(long));
	POOL = (SkewHeap*)malloc((m + n - 1) * sizeof(SkewHeap));
    SkewHeap **mins = (SkewHeap**)malloc(N2 * sizeof(SkewHeap*));
 
	#pragma omp parallel
	{
		#pragma omp for
		for (int i = 0; i < n; i++) {
			Edge e = edges[i];
			mins[e.dst] = Push(mins[e.dst], e.cost, i);
			if (i == s) {
				continue;
			}
 	    	edges[m++] = (Edge){i, s, 0, -1};
		}
		#pragma omp for
		for (int i = n; i < m; i++) {
			Edge e = edges[i];
			mins[e.dst] = Push(mins[e.dst], e.cost, i);
		}
		#pragma omp for
		for (int i = 0; i < N2; i++) {
			super[i] = i; 
			parent[i] = -1;
			visited[i] = -1;
		}
	}

	int pathFront = 0;
	for (int vc = n; mins[pathFront] != NULL; vc++) {
		while (visited[pathFront] == -1) {
			visited[pathFront] = 0;
			pathFront = FIND_PREV(pathFront);
		}
		while (pathFront != vc) {
			int w = mins[pathFront]->cost;
			SkewHeap* min = Pop(mins[pathFront]);
			Update(min, -w);
			mins[vc] = Merge(mins[vc], min);
			parent[pathFront] = vc;
			super[pathFront] = vc;
			pathFront = FIND_PREV(pathFront);
		}
		while (mins[pathFront] != NULL && FIND_PREV(pathFront) == pathFront) {
			mins[pathFront] = Pop(mins[pathFront]);
		}
	}

	long totalCost = 0;
	for (int v = s; v != -1; v = parent[v]) {
		visited[v] = 1;
	}
	parent[s] = s;
	for (int v = pathFront; v > -1; v--) {
		if (visited[v] == 1) {
			continue;
		}
		Edge e = IN_EDGE(v);
		int dst = e.dst;
		while (dst != v) {
			visited[dst] = 1;
			dst = parent[dst];
		}
		totalCost += e.cost;
		parent[e.dst] = e.src;
	}
    parent[n] = totalCost;
	return parent;
}


// -------------------- Main -------------------

int main() {
	int n, m, s, src, dst, cost;

    fill();
	
	n = readInt();
	m = readInt();
	s = readInt();

    Edge* edges = (Edge*)malloc((m + n - 1) * sizeof(Edge));
	for (int i = 0; i < m; i++) {
		src = readInt();
		dst = readInt();
		cost = readInt();
		edges[i] = (Edge){src, dst, cost, i};
	}

	long* parents = Tarjan(edges, n, m, s);

	writeLong(parents[n]);
	outputBuffer[outputPtr - 1] = '\n';
	for (int i = 0; i < n; i++) {
		writeLong(parents[i]);
	}
	flush();
}
