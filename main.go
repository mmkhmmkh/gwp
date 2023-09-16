package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"

	"github.com/montanaflynn/stats"
)

type Node struct {
	X int
	Y int
}

type Nodes []Node

type Ground struct {
	W        int
	H        int
	gateways Nodes
}

func (n Node) Distance(other Node) float64 {
	dY := n.Y - other.Y
	dX := n.X - other.X
	return math.Sqrt(float64(dY*dY + dX*dX))
}

func (g *Ground) CreateGateways(gwcount int) {
	for i := 0; i < gwcount; i++ {
		gw := Node{}
		g.gateways = append(g.gateways, gw)
		// for j := 0; j < i; j++ {
		// 	g.Increment(gw)
		// }
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func (g Ground) GrowAndMeasure(gwIndex, gwCount int) (sums, sds stats.Float64Data) {

	if gwIndex >= len(g.gateways) {
		sum, sd := g.MeasureGround(gwCount)
		// fmt.Printf("sum %v sd %v\n", sum, sd)
		return []float64{sum}, []float64{sd}
	}

	if gwIndex == len(g.gateways)-1 {
		// Leaf!
		c := (g.W + 1) * (g.H + 1)
		sums := make(chan float64, c)
		sds := make(chan float64, c)

		for {
			go func(g Ground, gwIndex, gwCount int, sums chan float64, sds chan float64) {
				newSums, newSds := g.GrowAndMeasure(gwIndex+1, gwCount)
				sums <- newSums[0]
				sds <- newSds[0]
			}(g, gwIndex, gwCount, sums, sds)

			if !g.Increment(&g.gateways[gwIndex]) {
				break
			}
		}

		var sumsArr []float64
		var sdsArr []float64

		for i := 0; i < c; i++ {
			sumsArr = append(sumsArr, <-sums)
			sdsArr = append(sdsArr, <-sds)
		}

		return sumsArr, sdsArr
	}

	for {
		newSums, newSds := g.GrowAndMeasure(gwIndex+1, gwCount)
		sums = append(sums, newSums...)
		sds = append(sds, newSds...)

		if !g.Increment(&g.gateways[gwIndex]) {
			break
		}
	}

	return sums, sds
}

// false if finished the travel!
func (g Ground) Increment(n *Node) bool {
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

func (g Ground) MeasureNode(n Node, gwcount int) (sum float64, sd float64) {
	dists := g.GatewaysDist(n, gwcount)

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

func (g Ground) MeasureGround(gwcount int) (sum float64, sd float64) {
	n := Node{}
	// c := (g.H + 1) * (g.W + 1)
	// sums := make(chan float64, c)
	// sds := make(chan float64, c)

	var sumsArr []float64
	var sdsArr []float64
	for {
		sum, sd := g.MeasureNode(n, gwcount)

		sumsArr = append(sumsArr, sum)
		sdsArr = append(sdsArr, sd)

		if !g.Increment(&n) {
			break
		}
	}

	sum, err := stats.Mean(sumsArr)
	if err != nil {
		panic(err)
	}

	sd, err = stats.Mean(sdsArr)
	if err != nil {
		panic(err)
	}

	return sum, sd
}

func (g Ground) GatewaysDist(n Node, gwcount int) []float64 {
	var gwds []float64
	for i := 0; i < gwcount; i++ {
		gwd := g.gateways.NearestDist(n, gwds)
		gwds = append(gwds, gwd)
	}
	return gwds
}

func (nodes Nodes) NearestDist(n Node, exclude Dists) float64 {
	minDist := math.MaxFloat64
	for _, node := range nodes {
		dist := n.Distance(node)
		if dist < minDist {
			if exclude != nil && exclude.Contains(dist) {
				exclude.Remove(dist)
				continue
			}
			minDist = dist
		}
	}
	return minDist
}

type Dists []float64

func (dists Dists) Contains(dist float64) bool {

	for _, d := range dists {
		if d == dist {
			return true
		}
	}
	return false
}

func remove(slice *Dists, s int) {
	*slice = append((*slice)[:s], (*slice)[s+1:]...)
}

func (dists *Dists) Remove(dist float64) {
	for i, d := range *dists {
		if d == dist {
			remove(dists, i)
			break
		}
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(1000 * 1000 * 1000 * 7)

	g := Ground{
		W: 100,
		H: 100,
	}

	N := 8

	g.CreateGateways(N)

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
