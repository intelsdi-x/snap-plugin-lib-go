package nstrie

import "sync"

// NSTrie is a trie which holds values via a namespace.
type NSTrie struct {
	sync.RWMutex

	children []*NSTrie
	key      string
	val      interface{}
}

// New returns a root NSTrie
func New() *NSTrie {
	return &NSTrie{
		children: make([]*NSTrie, 0),
	}
}

// Get returns the value at the searched-for namespace, and a bool for whether
// or not the search was successful.
// If the search fails, the return is (nil, false).
func (t *NSTrie) Get(namespace []string) (interface{}, bool) {
	var (
		node *NSTrie
		idx  = 0
	)

	t.RLock()
	defer t.RUnlock()

	node = t.get(namespace, &idx)
	if idx == len(namespace) && node != nil && node.val != nil {
		return node.val, true
	}
	return nil, false
}

// Put writes a value at Namespace.  If the value exists, it will overwrite it.
func (t *NSTrie) Put(namespace []string, val interface{}) {
	var (
		node *NSTrie
		idx  = 0
	)

	t.Lock()
	defer t.Unlock()

	node = t.get(namespace, &idx)
	if idx == len(namespace) {
		node.val = val
		return
	}

	node.put(namespace[idx:], val)
}

// Fetch returns all childen of the searched-for namespace.  If the search is
// unsuccessful, it returns (nil, false)
func (t *NSTrie) Fetch(namespace []string) ([]interface{}, bool) {
	var (
		from     *NSTrie
		children []*NSTrie
		idx      = 0
	)

	t.RLock()
	defer t.RUnlock()

	// test for empty lookup (fetch from root)
	if len(namespace) == 0 || (len(namespace) == 1 && namespace[0] == "") {
		from = t
	} else {
		from = t.get(namespace, &idx)
		if idx != len(namespace) {
			return nil, false
		}
	}

	children = append(children, from)
	from.gatherChildren(&children)

	childVals := make([]interface{}, 0, len(children))
	for i := range children {
		if children[i].val != nil {
			childVals = append(childVals, children[i].val)
		}
	}
	return childVals, true
}

// TestAndPut transactionally checks whether or not the value exists before
// attempting to write the given value.  If there is already a value at
// namespace, it returns false and does not write.
func (t *NSTrie) TestAndPut(namespace []string, val interface{}) bool {
	var (
		node *NSTrie
		idx  = 0
	)

	t.Lock()
	defer t.Unlock()

	node = t.get(namespace, &idx)
	if idx == len(namespace) && node != nil && node.val != nil {
		return false
	}

	node.put(namespace[idx:], val)
	return true
}

func (t *NSTrie) get(namespace []string, idx *int) *NSTrie {
	if len(namespace) == 0 {
		return t
	}

	for _, child := range t.children {
		if namespace[0] == child.key {
			(*idx)++
			t = child.get(namespace[1:], idx)
			return t
		}
	}
	return t
}

func (t *NSTrie) put(namespace []string, val interface{}) {
	if len(namespace) == 0 {
		t.val = val
		return
	}
	newChild := New()
	newChild.key = namespace[0]
	newChild.put(namespace[1:], val)
	t.children = append(t.children, newChild)
}

func (t *NSTrie) gatherChildren(children *[]*NSTrie) {
	for _, child := range t.children {
		*children = append(*children, child)
		child.gatherChildren(children)
	}
}
