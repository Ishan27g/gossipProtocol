package gossip

import (
	"strconv"
	"testing"

	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/stretchr/testify/assert"
)

func mockView(hop int) view {
	n := sll.New()
	for i := 1; i <= 9; i++ {
		n.Add(nodeDescriptor{
			address: "120" + strconv.Itoa(i),
			hop:     hop,
		})
	}
	return view{nodes: n}
}
func TestMerge(t *testing.T) {

	lowerHop, higherHop := 0, 2
	v1 := mockView(lowerHop)
	v2 := mockView(higherHop)

	merged := mergeView(v1,v2)

	assert.Equal(t, merged.nodes.Size(), v1.nodes.Size())
	merged.nodes.Each(func(_ int, value interface{}) {
		n := value.(nodeDescriptor)
		assert.Equal(t, n.hop, lowerHop)
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
	v.randomView()
	assert.NotNil(t, v.randomNode())
	assert.Equal(t, MaxNodesInView, v.nodes.Size())
}
func TestHeadView(t *testing.T) {
	v := mockView(0)
	v.headView()
	assert.Equal(t, MaxNodesInView, v.nodes.Size())
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
	assert.Equal(t, MaxNodesInView, v.nodes.Size())
	assert.Equal(t, "1204", v.headNode())
	assert.Equal(t, "1209", v.tailNode())
	exists, hop := v.checkExists("1205")
	assert.Equal(t, true, exists)
	assert.Equal(t, 0, hop)
}

func TestSerialization(t *testing.T) {
	v := mockView(4)
	bytes := viewToBytes(v)
	v2, e := bytesToView(bytes)
	assert.NoError(t, e)
	v2.nodes.Each(func(_ int, value interface{}) {
		n := value.(nodeDescriptor)
		assert.NotEqual(t, -1, v.nodes.IndexOf(n))
	})
}
