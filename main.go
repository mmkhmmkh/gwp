package main

import (
	"fmt"
	"math"

	"github.com/montanaflynn/stats"
)

type Node struct {
	X int
	Y int
}

type Nodes []*Node

type Ground struct {
	W        int
	H        int
	gateways Nodes
}

func (n *Node) Distance(other *Node) float64 {
	return math.Sqrt(math.Pow(float64(n.Y)-float64(other.Y), 2) +
		math.Pow(float64(n.X)-float64(other.X), 2))
}

func (g *Ground) CreateGateways(gwcount int) {
	for i := 0; i < gwcount; i++ {
		gw := &Node{}
		g.gateways = append(g.gateways, gw)
		// for j := 0; j < i; j++ {
		// 	g.Increment(gw)
		// }
	}
}

func (g *Ground) GrowAndMeasure(gwIndex, gwCount int) (sums, sds stats.Float64Data) {
	if gwIndex >= len(g.gateways) {
		sum, sd := g.MeasureGround(gwCount)
		// fmt.Printf("sum %v sd %v\n", sum, sd)
		return []float64{sum}, []float64{sd}
	}
	for {
		newSums, newSds := g.GrowAndMeasure(gwIndex+1, gwCount)
		sums = append(sums, newSums...)
		sds = append(sds, newSds...)

		if !g.Increment(g.gateways[gwIndex]) {
			break
		}
		if gwIndex < 5 {
			fmt.Printf("GW %v: %v,%v\n", gwIndex, g.gateways[gwIndex].X, g.gateways[gwIndex].Y)
		}
	}
	return sums, sds
}

// false if finished the travel!
func (g *Ground) Increment(n *Node) bool {
	n.X += 1
	if n.X > g.W {
		n.Y++
		n.X = 0
		if n.Y > g.H {
			n.Y = 0
			return false
		}
	}
	return true
}

func (g *Ground) MeasureNode(n *Node, gwcount int) (sum float64, sd float64) {
	gws := g.AssignGateways(n, gwcount)
	var dists []float64
	for _, gw := range gws {
		dists = append(dists, n.Distance(gw))
	}
	sd, err := stats.StandardDeviation(dists)
	if err != nil {
		panic(err)
	}

	sum, err = stats.Sum(dists)
	if err != nil {
		panic(err)
	}

	return sum, sd
}

func (g *Ground) MeasureGround(gwcount int) (sum float64, sd float64) {
	n := &Node{
		X: -1,
		Y: 0,
	}
	var sums []float64
	var sds []float64
	for g.Increment(n) {
		sum, sd := g.MeasureNode(n, gwcount)
		sums = append(sums, sum)
		sds = append(sds, sd)
	}

	sum, err := stats.Mean(sums)
	if err != nil {
		panic(err)
	}

	sd, err = stats.Mean(sds)
	if err != nil {
		panic(err)
	}

	return sum, sd
}

func (g *Ground) AssignGateways(n *Node, gwcount int) Nodes {
	var gws Nodes
	for i := 0; i < gwcount; i++ {
		gw := g.gateways.Nearest(n, gws)
		if gw == nil {
			panic("No enough gateways!")
		}
		gws = append(gws, gw)
	}
	return gws
}

func (nodes Nodes) Nearest(n *Node, exclude Nodes) *Node {
	minDist := math.MaxFloat64
	var minNode *Node
	minNode = nil
	for _, node := range nodes {
		dist := n.Distance(node)
		if dist < minDist {
			if exclude != nil && exclude.Contains(node) {
				exclude.Remove(node)
				continue
			}
			minDist = dist
			minNode = node
		}
	}
	return minNode
}

func (nodes Nodes) Contains(n *Node) bool {
	for _, node := range nodes {
		if n.X == node.X && n.Y == node.Y {
			return true
		}
	}
	return false
}

func remove(slice *Nodes, s int) {
	*slice = append((*slice)[:s], (*slice)[s+1:]...)
}

func (nodes *Nodes) Remove(n *Node) {
	for i, node := range *nodes {
		if n.X == node.X && n.Y == node.Y {
			remove(nodes, i)
			return
		}
	}
}

func main() {
	g := Ground{
		W: 10,
		H: 10,
	}

	g.CreateGateways(4)

	sums, sds := g.GrowAndMeasure(0, 3)

	minSum, err := sums.Min()
	if err != nil {
		panic(err)
	}
	minSd, err := sds.Min()
	if err != nil {
		panic(err)
	}

	fmt.Printf("FINAL Min sum %v sd %v\n", minSum, minSd)
}
