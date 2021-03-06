package graphs

import (
	"fmt"
	"log"
	"testing"
)

func TestTree(t *testing.T) {
	g := &Graph{Names: []string{"planetlab2.cs.unc.edu", "pl1.6test.edu.cn", "planetlab1.cs.du.edu", "planetlab02.cs.washington.edu", "planetlab-2.cse.ohio-state.edu", "planetlab2.cs.ubc.ca"}, mem: []float64{0, 213.949, 51.86, 76.716, 2754.531, 81.301, 214.143, 0, 169.744, 171.515, 557.526, 189.186, 51.601, 170.191, 0, 41.418, 2444.206, 31.475, 76.731, 171.43, 41.394, 0, 2470.722, 5.741, 349.881, 520.028, 374.362, 407.282, 0, 392.211, 81.381, 189.386, 31.582, 5.78, 141.273, 0}, Weights: [][]float64{[]float64{0, 213.949, 51.86, 76.716, 2754.531, 81.301}, []float64{214.143, 0, 169.744, 171.515, 557.526, 189.186}, []float64{51.601, 170.191, 0, 41.418, 2444.206, 31.475}, []float64{76.731, 171.43, 41.394, 0, 2470.722, 5.741}, []float64{349.881, 520.028, 374.362, 407.282, 0, 392.211}, []float64{81.381, 189.386, 31.582, 5.78, 141.273, 0}}}
	tree := g.Tree(2)
	log.Println(tree)
}

func TestTreeFromList(t *testing.T) {
	nodeNames := make([]string, 0)
	nodeNames = append(nodeNames, "machine0", "machine1", "machine2")
	hostsPerNode := 2
	bf := 2

	root, usedHosts, err := TreeFromList(nodeNames, hostsPerNode, bf)
	if err != nil {
		panic(err)
	}

	// JSON format
	// b, err := json.Marshal(root)
	// if err != nil {
	// 	t.Error(err)
	// }
	// fmt.Println(string(b))

	if len(usedHosts) != len(nodeNames)*hostsPerNode {
		t.Error("Should have been able to use all hosts")
	}
	fmt.Println("used hosts", usedHosts)
	root.TraverseTree(PrintTreeNode)

	// Output:
	// used hosts [machine0:32600 machine1:32600 machine1:32610 machine0:32610 machine2:32600 machine2:32610]
	// machine0:32600
	// 	 machine1:32600
	// 	 machine1:32610
	// machine1:32600
	// 	 machine0:32610
	// 	 machine2:32600
	// machine0:32610
	// machine2:32600
	// machine1:32610
	// 	 machine2:32610
	// machine2:32610
}
