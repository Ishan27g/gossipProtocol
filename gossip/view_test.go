package gossip

import (
	"strconv"
	"testing"

	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/stretchr/testify/assert"
)

func mockView(hop int) View {
	n := sll.New()
	for i := 1; i <= 9; i++ {
		n.Add(NodeDescriptor{
			Address: "120" + strconv.Itoa(i),
			Hop:     hop,
		})
	}
	return View{Nodes: n}
}
func TestMerge(t *testing.T) {

	lowerHop, higherHop := 0, 2
	v1 := mockView(lowerHop)
	v2 := mockView(higherHop)

	merged := mergeView(v1,v2)

	assert.Equal(t, merged.Nodes.Size(), v1.Nodes.Size())
	merged.Nodes.Each(func(_ int, value interface{}) {
		n := value.(NodeDescriptor)
		assert.Equal(t, n.Hop, lowerHop)
	})
}

func TestViewNodes(t *testing.T) {
	v1 := mockView(0)
	assert.Equal(t, v1.headNode(), "1201")
	assert.Equal(t, v1.tailNode(), "1209")
	assert.NotNil(t, v1.randomNode())
}
func TestRandomView(t *testing.T) {
	v := mockView(0)
	v.RandomView()
	assert.NotNil(t, v.randomNode())
	assert.Equal(t, MaxNodesInView, v.Nodes.Size())
}
func TestHeadView(t *testing.T) {
	v := mockView(0)
	v.headView()
	assert.Equal(t, MaxNodesInView, v.Nodes.Size())
	assert.Equal(t, "1201", v.headNode())
	assert.Equal(t, "1206", v.tailNode())
	exists, hop := v.checkExists("1205")
	assert.Equal(t, true, exists)
	assert.Equal(t, 0, hop)
}
func TestTailView(t *testing.T) {
	v := mockView(0)
	v.tailView()
	v.sortByAddr()
	assert.Equal(t, MaxNodesInView, v.Nodes.Size())
	assert.Equal(t, "1204", v.headNode())
	assert.Equal(t, "1209", v.tailNode())
	exists, hop := v.checkExists("1205")
	assert.Equal(t, true, exists)
	assert.Equal(t, 0, hop)
}

func TestSerialization(t *testing.T) {
	v := mockView(4)
	bytes := ViewToBytes(v)
	v2, e := BytesToView(bytes)
	assert.NoError(t, e)
	v2.Nodes.Each(func(_ int, value interface{}) {
		n := value.(NodeDescriptor)
		assert.NotEqual(t, -1, v.Nodes.IndexOf(n))
	})
}
