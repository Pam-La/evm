package preconflicts

import (
	"container/heap"
	"sort"
)

type GroupResult struct {
	Groups    map[int][]int // 그룹 ID -> 트랜잭션 인덱스들
	TxToGroup map[int]int   // 트랜잭션 인덱스 -> 그룹 ID
	NumGroups int           // 총 그룹 수
}

type Vertex struct {
	Index          int          // 트랜잭션 인덱스 번호
	Diversity      int          // 이웃한 정점의 서로 다른 그룹 수
	Degree         int          // 연결된 간선의 수 (초기 순서용)
	Group          int          // 할당된 그룹 번호 (-1: 할당되지 않음)
	NeighborGroups map[int]bool // 이웃한 정점의 그룹 번호들
	Neighbors      []*Vertex    // 이웃한 정점들의 포인터 (빠른 접근)
	HeapIndex      int          // heap에서의 인덱스 (효율적인 업데이트용)
}

func NewVertex(index int) *Vertex {
	return &Vertex{
		Index:          index,
		Diversity:      0,
		Degree:         0,
		Group:          -1,
		NeighborGroups: make(map[int]bool),
		Neighbors:      []*Vertex{},
		HeapIndex:      -1,
	}
}

func (v *Vertex) AddNeighbor(neighbor *Vertex) {
	v.Neighbors = append(v.Neighbors, neighbor)
	v.Degree = len(v.Neighbors)
}

func (v *Vertex) UpdateSaturation() {
	v.NeighborGroups = make(map[int]bool)
	for _, neighbor := range v.Neighbors {
		if neighbor.Group != -1 {
			v.NeighborGroups[neighbor.Group] = true
		}
	}
	v.Diversity = len(v.NeighborGroups)
}

type ConflictGraph struct {
	Vertices    map[int]*Vertex // 정점들
	NextGroupID int             // 다음에 할당할 그룹 ID
}

func NewConflictGraph() *ConflictGraph {
	return &ConflictGraph{
		Vertices:    make(map[int]*Vertex),
		NextGroupID: 1, // 그룹 0은 독립 트랜잭션용으로 예약
	}
}

func (ocg *ConflictGraph) AddVertex(txIndex int) {
	if _, exists := ocg.Vertices[txIndex]; !exists {
		ocg.Vertices[txIndex] = NewVertex(txIndex)
	}
}

func (ocg *ConflictGraph) AddEdge(tx1Index, tx2Index int) {
	ocg.AddVertex(tx1Index)
	ocg.AddVertex(tx2Index)

	v1 := ocg.Vertices[tx1Index]
	v2 := ocg.Vertices[tx2Index]

	v1.AddNeighbor(v2)
	v2.AddNeighbor(v1)
}

func (ocg *ConflictGraph) BuildFromConflictResult(conflictMap map[int][]int) {
	// 충돌하는 트랜잭션들만 그래프에 추가
	for txIdx := range conflictMap {
		ocg.AddVertex(txIdx)
	}

	// 충돌 관계를 간선으로 추가
	for txIndex, conflictingTxs := range conflictMap {
		for _, conflictTxIndex := range conflictingTxs {
			if txIndex < conflictTxIndex {
				ocg.AddEdge(txIndex, conflictTxIndex)
			}
		}
	}
}

type PriorityQueue []*Vertex

func (opq PriorityQueue) Len() int { return len(opq) }

func (opq PriorityQueue) Less(i, j int) bool {
	if opq[i].Diversity != opq[j].Diversity {
		return opq[i].Diversity > opq[j].Diversity
	}
	if opq[i].Degree != opq[j].Degree {
		return opq[i].Degree > opq[j].Degree
	}
	return opq[i].Index < opq[j].Index
}

func (opq PriorityQueue) Swap(i, j int) {
	opq[i], opq[j] = opq[j], opq[i]
	opq[i].HeapIndex = i
	opq[j].HeapIndex = j
}

func (opq *PriorityQueue) Push(x interface{}) {
	n := len(*opq)
	vertex := x.(*Vertex)
	vertex.HeapIndex = n
	*opq = append(*opq, vertex)
}

func (opq *PriorityQueue) Pop() interface{} {
	old := *opq
	n := len(old)
	vertex := old[n-1]
	vertex.HeapIndex = -1
	*opq = old[0 : n-1]
	return vertex
}

func (opq *PriorityQueue) Update(vertex *Vertex) {
	if vertex.HeapIndex >= 0 && vertex.HeapIndex < len(*opq) {
		heap.Fix(opq, vertex.HeapIndex)
	}
}

func (ocg *ConflictGraph) AssignGroups(independentTxs []int) *GroupResult {
	result := &GroupResult{
		Groups:    make(map[int][]int),
		TxToGroup: make(map[int]int),
		NumGroups: 1,
	}

	// 1. 그룹 0에 독립 트랜잭션들 할당
	if len(independentTxs) > 0 {
		result.Groups[0] = make([]int, len(independentTxs))
		copy(result.Groups[0], independentTxs)
		for _, txIndex := range independentTxs {
			result.TxToGroup[txIndex] = 0
		}
	}

	// 충돌 트랜잭션이 없으면 종료
	if len(ocg.Vertices) == 0 {
		return result
	}

	pq := make(PriorityQueue, 0, len(ocg.Vertices))
	for _, vertex := range ocg.Vertices {
		vertex.UpdateSaturation() // 초기에는 모두 0
		heap.Push(&pq, vertex)
	}

	for pq.Len() > 0 {
		// 최고 우선순위 정점 선택
		vertex := heap.Pop(&pq).(*Vertex)

		// 할당 가능한 최소 그룹 찾기
		assignedGroup := ocg.findSmallestAvailableGroup(vertex)

		// 그룹 할당
		vertex.Group = assignedGroup
		result.TxToGroup[vertex.Index] = assignedGroup

		if result.Groups[assignedGroup] == nil {
			result.Groups[assignedGroup] = []int{}
		}
		result.Groups[assignedGroup] = append(result.Groups[assignedGroup], vertex.Index)

		// 그룹 수 업데이트
		if assignedGroup >= result.NumGroups {
			result.NumGroups = assignedGroup + 1
		}

		// 이웃 정점들의 diversity 업데이트
		for _, neighbor := range vertex.Neighbors {
			if neighbor.Group == -1 { // 아직 미할당인 이웃만
				oldDiversity := neighbor.Diversity
				neighbor.UpdateSaturation()

				// Diversity가 변경된 경우에만 힙 업데이트
				if neighbor.Diversity != oldDiversity {
					pq.Update(neighbor)
				}
			}
		}
	}

	for groupID := range result.Groups {
		sort.Ints(result.Groups[groupID])
	}

	return result
}

func (ocg *ConflictGraph) findSmallestAvailableGroup(vertex *Vertex) int {
	for groupID := 1; groupID < ocg.NextGroupID; groupID++ {
		if !vertex.NeighborGroups[groupID] {
			return groupID
		}
	}

	// 사용 가능한 기존 그룹이 없으면 새 그룹 생성
	newGroupID := ocg.NextGroupID
	ocg.NextGroupID++
	return newGroupID
}
