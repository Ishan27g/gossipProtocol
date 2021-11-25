package gossip
/*
	Partial view & strategies based on http://lpdwww.epfl.ch/upload/documents/publications/neg--1184036295all.pdf
*/
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/emirpasic/gods/containers"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/emirpasic/gods/utils"
)

const LogName = "view"

// view of at max MaxNodesInView nodes in the network
// updated during selectView()
type view struct {
	nodes *sll.List
}

// increaseHopCount for each node in the view
func increaseHopCount(v *view) {
	nodes := sll.New()
	for it := v.nodes.Iterator(); it.Next(); {
		n1 := it.Value().(nodeDescriptor)
		n1.hop++
		nodes.Add(n1)
	}
	v.nodes.Clear()
	v.nodes = nodes
	v.sortNodes()
}

// checkExists checks whether the address exists in the view
func (v *view) checkExists(address string) (bool, int) {
	for it := v.nodes.Iterator(); it.Next(); {
		n1 := it.Value().(nodeDescriptor)
		if n1.address == address {
			return true, n1.hop
		}
	}
	return false, -1
}
// mergeView view2 into view 1, discarding duplicate nodes with higher hop count
func mergeViewExcludeNode(view1, view2 view, n nodeDescriptor) view {
	return mergeMaps(toMap(view1, n.address), toMap(view2, n.address))
}

// mergeView view2 into view 1, discarding duplicate nodes with higher hop count
func mergeView(view1, view2 view) view {
	merged := view{nodes: sll.New()}
	// For duplicate nodes in v2map, merge with v1map with lower hop
	merge(toMap(view1, ""), toMap(view2, ""), merged)
	merged.sortNodes()
	return merged
}

// mergeView view2 into view 1, discarding duplicate nodes with higher hop count
func mergeMaps(v1map, v2map map[string]int) view {
	merged := view{nodes: sll.New()}
	// For duplicate nodes in v2map, merge with v1map with lower hop
	merge(v1map, v2map, merged)
	merged.sortNodes()
	return merged
}
func toMap(view1 view, excludeAddr string) map[string]int {
	// convert view to hashmap
	v1map := make(map[string]int) //address:hop
	for it := view1.nodes.Iterator(); it.Next(); {
		n1 := it.Value().(nodeDescriptor)
		if excludeAddr != "" {
			if strings.Compare(excludeAddr, n1.address) != 0 {
				v1map[n1.address] = n1.hop
			}
		}else {
			v1map[n1.address] = n1.hop
		}
	}
	return v1map
}
// merge For duplicate nodes in `from`, merge the ones with lower hop
func merge(into map[string]int, from map[string]int, merged view) {
	for address, hop := range into {
		if from[address] > hop {
			merged.nodes.Add(nodeDescriptor{
				address: address,
				hop:     hop, // lower hop
			})
		} else {
			merged.nodes.Add(nodeDescriptor{
				address: address,
				hop:     from[address], // lower hop
			})
		}
	}
}
func (v *view) sortByAddr() {
	c := utils.Comparator(func(a, b interface{}) int {
		n1 := a.(nodeDescriptor)
		n2 := b.(nodeDescriptor)
		return strings.Compare(n1.address, n2.address)
	})
	sortedNodes := containers.GetSortedValues(v.nodes, c)
	v.nodes.Clear()
	for _, va := range sortedNodes {
		n := va.(nodeDescriptor)
		v.nodes.Add(n)
	}
}

// sortNodes according to increasing hop count
func (v *view) sortNodes() {
	c := utils.Comparator(func(a, b interface{}) int {
		n1 := a.(nodeDescriptor)
		n2 := b.(nodeDescriptor)
		if n1.hop > n2.hop {
			return 1
		}
		if n1.hop < n2.hop {
			return -1
		}
		return 0
	})
	v.nodes.Sort(c)
	v.sortByAddr()
}

// headNode returns the node with the lowest Hop count
func (v *view) headNode() string {
	node, _ := v.nodes.Get(0)
	return node.(nodeDescriptor).address
}

// tailNode returns the node with the highest Hop count
func (v *view) tailNode() string {
	node, _ := v.nodes.Get(v.nodes.Size() - 1)
	return node.(nodeDescriptor).address
}

// randomNode returns a random node from the list
func (v *view) randomNode() string {
	rand.Seed(time.Now().Unix())
	node, _ := v.nodes.Get(rand.Intn(v.nodes.Size()))
	return node.(nodeDescriptor).address
}

// randomView sets the current view as a random subset of current view
func (v *view) randomView() {
	selection := sll.New()
	for {
		rand.Seed(time.Now().Unix())
		node, _ := v.nodes.Get(rand.Intn(v.nodes.Size()))

		if selection.IndexOf(node) == -1 {
			// don't add self if present in view
			//if strings.Compare(node.(nodeDescriptor).address, Env.envCfg.Hostname +":" + Env.envCfg.UdpPort) != 0{
				selection.Add(node)
			// }
		}
		if selection.Size() == MaxNodesInView {
			break
		}
	}
	v.nodes.Clear()
	v.nodes = selection
	v.sortNodes()
}

// headView sets the current view as the subset of first MaxNodesInView nodes in the current view
func (v *view) headView() {
	selection := sll.New()
	for i := 0; i < MaxNodesInView; i++ {
		node, _ := v.nodes.Get(i)
		selection.Add(node)
	}
	v.nodes.Clear()
	v.nodes = selection
	v.sortNodes()
}

// tailView sets the current view as the subset of last MaxNodesInView nodes in the current view
func (v *view) tailView() {
	selection := sll.New()
	for i := v.nodes.Size() - 1; i >= v.nodes.Size()-MaxNodesInView; i-- {
		node, _ := v.nodes.Get(i)
		selection.Add(node.(nodeDescriptor))
	}
	v.nodes.Clear()
	v.nodes = selection
	v.sortNodes()
}

func viewToBytes(view view) []byte {
	m := make(map[string]int)
	for it := view.nodes.Iterator(); it.Next(); {
		node := it.Value().(nodeDescriptor)
		m[node.address] = node.hop
	}
	b, _ := json.Marshal(m)
	return b
}
func bytesToView(bytes []byte) (view, error) {
	var m map[string]int
	if err := json.Unmarshal(bytes, &m); err != nil {
		fmt.Println("Unmarshal error" + err.Error())
		return view{}, err
	}
	v := view{nodes: sll.New()}
	for addr, hop := range m {
		v.nodes.Add(nodeDescriptor{address: addr, hop: hop})
	}
	return v, nil
}
func printView(view view) string{
	str := "view len - " + strconv.Itoa(view.nodes.Size())
	//mLogger.Get(LogName).Info("view len - " + strconv.Itoa(view.nodes.Size()))
	view.nodes.Each(func(_ int, value interface{}) {
		n := value.(nodeDescriptor)
		str += "\n" + n.address + "[" + strconv.Itoa(n.hop) + "]"
		// mLogger.Get(LogName).Info(n.address + "[" + strconv.Itoa(n.hop) + "]")
	})
	return str
}
