package gossip

/*
	Partial View & strategies based on http://lpdwww.epfl.ch/upload/documents/publications/neg--1184036295all.pdf
*/
import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/emirpasic/gods/containers"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/emirpasic/gods/utils"
)

const LogName = "View"

// View of at max MaxNodesInView nodes in the network
// updated during selectView()
type View struct {
	Nodes *sll.List
}

type NodeDescriptor struct {
	Address string
	Hop     int
}

// increaseHopCount for each node in the View
func increaseHopCount(v *View) {
	nodes := sll.New()
	for it := v.Nodes.Iterator(); it.Next(); {
		n1 := it.Value().(NodeDescriptor)
		n1.Hop++
		nodes.Add(n1)
	}
	v.Nodes.Clear()
	v.Nodes = nodes
	v.sortNodes()
}

// checkExists checks whether the address exists in the View
func (v *View) checkExists(address string) (bool, int) {
	for it := v.Nodes.Iterator(); it.Next(); {
		n1 := it.Value().(NodeDescriptor)
		if n1.Address == address {
			return true, n1.Hop
		}
	}
	return false, -1
}

// mergeView view2 into View 1, discarding duplicate nodes with higher hop count
func mergeViewExcludeNode(view1, view2 View, n NodeDescriptor) View {
	return mergeMaps(toMap(view1, n.Address), toMap(view2, n.Address))
}

// mergeView view2 into View 1, discarding duplicate nodes with higher hop count
func mergeView(view1, view2 View) View {
	merged := View{Nodes: sll.New()}
	// For duplicate nodes in v2map, merge with v1map with lower hop
	merge(toMap(view1, ""), toMap(view2, ""), merged)
	merged.sortNodes()
	return merged
}

// mergeView view2 into View 1, discarding duplicate nodes with higher hop count
func mergeMaps(v1map, v2map map[string]int) View {
	merged := View{Nodes: sll.New()}
	// For duplicate nodes in v2map, merge with v1map with lower hop
	merge(v1map, v2map, merged)
	merged.sortNodes()
	return merged
}
func toMap(view1 View, excludeAddr string) map[string]int {
	// convert View to hashmap
	v1map := make(map[string]int) //address:hop
	for it := view1.Nodes.Iterator(); it.Next(); {
		n1 := it.Value().(NodeDescriptor)
		if excludeAddr != "" {
			if strings.Compare(excludeAddr, n1.Address) != 0 {
				v1map[n1.Address] = n1.Hop
			}
		} else {
			v1map[n1.Address] = n1.Hop
		}
	}
	return v1map
}

// merge For duplicate nodes in `from`, merge the ones with lower hop
func merge(into map[string]int, from map[string]int, merged View) {
	for address, hop := range into {
		if from[address] > hop {
			merged.Nodes.Add(NodeDescriptor{
				Address: address,
				Hop:     hop, // lower hop
			})
		} else {
			merged.Nodes.Add(NodeDescriptor{
				Address: address,
				Hop:     from[address], // lower hop
			})
		}
	}
}
func (v *View) sortByAddr() {
	c := utils.Comparator(func(a, b interface{}) int {
		n1 := a.(NodeDescriptor)
		n2 := b.(NodeDescriptor)
		return strings.Compare(n1.Address, n2.Address)
	})
	sortedNodes := containers.GetSortedValues(v.Nodes, c)
	v.Nodes.Clear()
	for _, va := range sortedNodes {
		n := va.(NodeDescriptor)
		v.Nodes.Add(n)
	}
}

// sortNodes according to increasing hop count
func (v *View) sortNodes() {
	c := utils.Comparator(func(a, b interface{}) int {
		n1 := a.(NodeDescriptor)
		n2 := b.(NodeDescriptor)
		if n1.Hop > n2.Hop {
			return 1
		}
		if n1.Hop < n2.Hop {
			return -1
		}
		return 0
	})
	v.Nodes.Sort(c)
	v.sortByAddr()
}

// headNode returns the node with the lowest Hop count
func (v *View) headNode() string {
	node, _ := v.Nodes.Get(0)
	return node.(NodeDescriptor).Address
}

// tailNode returns the node with the highest Hop count
func (v *View) tailNode() string {
	node, _ := v.Nodes.Get(v.Nodes.Size() - 1)
	return node.(NodeDescriptor).Address
}

// randomNode returns a random node from the list
func (v *View) randomNode() string {
	rand.Seed(time.Now().Unix())
	node, _ := v.Nodes.Get(rand.Intn(v.Nodes.Size()))
	return node.(NodeDescriptor).Address
}

// RandomView sets the current View as a random subset of current View
func (v *View) RandomView() {
	selection := sll.New()
	rand.Seed(time.Now().Unix())
	for {
		node, _ := v.Nodes.Get(rand.Intn(v.Nodes.Size()))

		if selection.IndexOf(node) == -1 {
			// don't add self if present in View
			//if strings.Compare(node.(NodeDescriptor).address, Env.envCfg.Hostname +":" + Env.envCfg.UdpPort) != 0{
			selection.Add(node)
			// }
		}
		if selection.Size() == MaxNodesInView {
			break
		}
	}
	v.Nodes.Clear()
	v.Nodes = selection
	v.sortNodes()
}

// headView sets the current View as the subset of first MaxNodesInView Nodes in the current View
func (v *View) headView() {
	selection := sll.New()
	for i := 0; i < MaxNodesInView; i++ {
		node, _ := v.Nodes.Get(i)
		selection.Add(node)
	}
	v.Nodes.Clear()
	v.Nodes = selection
	v.sortNodes()
}

// tailView sets the current View as the subset of last MaxNodesInView Nodes in the current View
func (v *View) tailView() {
	selection := sll.New()
	for i := v.Nodes.Size() - 1; i >= v.Nodes.Size()-MaxNodesInView; i-- {
		node, _ := v.Nodes.Get(i)
		selection.Add(node.(NodeDescriptor))
	}
	v.Nodes.Clear()
	v.Nodes = selection
	v.sortNodes()
}

func ViewToBytes(view View) []byte {
	m := make(map[string]int)
	for it := view.Nodes.Iterator(); it.Next(); {
		node := it.Value().(NodeDescriptor)
		m[node.Address] = node.Hop
	}
	b, _ := json.Marshal(m)
	return b
}
func BytesToView(bytes []byte) (View, error) {
	if bytes == nil {
		return View{}, errors.New("empty")
	}
	var m map[string]int
	if err := json.Unmarshal(bytes, &m); err != nil {
		fmt.Println("Unmarshal error" + err.Error())
		return View{}, err
	}
	v := View{Nodes: sll.New()}
	for addr, hop := range m {
		v.Nodes.Add(NodeDescriptor{Address: addr, Hop: hop})
	}
	return v, nil
}
func printView(view View) string {
	str := "View len - " + strconv.Itoa(view.Nodes.Size())
	//mLogger.Get(LogName).Info("View len - " + strconv.Itoa(View.nodes.Size()))
	view.Nodes.Each(func(_ int, value interface{}) {
		n := value.(NodeDescriptor)
		str += "\n" + n.Address + "[" + strconv.Itoa(n.Hop) + "]"
		// mLogger.Get(LogName).Info(n.address + "[" + strconv.Itoa(n.hop) + "]")
	})
	return str
}
