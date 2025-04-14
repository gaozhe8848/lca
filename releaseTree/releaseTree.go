package releaseTree

import (
	"errors"
	"fmt" // Import reflect for DeepEqual
	"sort"
	"strconv" // Import strings for error comparison
	"sync"
)

// Chg represents a single change item (Exported).
type Chg struct {
	ID string // Exported field
}

// ReleaseInput represents the raw data for a release node (Exported Type).
// Used as input for NewReleaseTree and InsertNode.
type ReleaseInput struct {
	Ver     string // Exported field
	FromVer string // Exported field
	Changes []Chg  // Exported field (slice of Exported type)
}

// node represents a node in the N-ary release tree (unexported).
type node struct {
	version  string  // unexported field
	changes  []Chg   // unexported field (slice of Exported type)
	parent   *node   // unexported field
	children []*node // unexported field
}

// ReleaseTree holds the entire tree structure.
// Provides exported methods for interaction.
type ReleaseTree struct {
	nodes map[string]*node // map version string to internal node pointer (unexported)
	root  *node            // pointer to the internal root node (unexported)
	mu    sync.RWMutex     // Reader/Writer Mutex for concurrent safety
}

// NewReleaseTree builds the n-ary tree from a slice of input release data.
// Takes []ReleaseInput (Exported type) and returns *ReleaseTree or error.
func NewReleaseTree(inputs []ReleaseInput) (*ReleaseTree, error) {
	tree := &ReleaseTree{
		nodes: make(map[string]*node),
	}

	// Pass 1: Create nodes
	for _, input := range inputs {
		// Use Exported field names (Ver)
		if _, exists := tree.nodes[input.Ver]; exists {
			return nil, fmt.Errorf("NewReleaseTree: duplicate version detected: %s", input.Ver)
		}
		// Copy changes slice (using Exported type Chg and field Changes)
		changesCopy := make([]Chg, len(input.Changes))
		copy(changesCopy, input.Changes)
		// Create internal node
		newNode := &node{
			version:  input.Ver,   // Use Exported field Ver
			changes:  changesCopy, // Use Exported type Chg
			children: []*node{},
		}
		tree.nodes[input.Ver] = newNode // Use Exported field Ver
	}

	// Pass 2: Link nodes
	foundRoots := 0
	for _, input := range inputs {
		newNode := tree.nodes[input.Ver] // Use Exported field Ver
		// Use Exported field FromVer
		if input.FromVer == "" {
			if tree.root == nil {
				tree.root = newNode
			} else if tree.root != newNode {
				fmt.Printf("Warning (NewReleaseTree): Multiple nodes (%s, %s) have empty FromVer. Using first encountered ('%s') as root.\n",
					tree.root.version, newNode.version, tree.root.version)
			}
			foundRoots++
			continue
		}
		// Use Exported field FromVer
		parent, exists := tree.nodes[input.FromVer]
		if !exists {
			// Use Exported fields Ver, FromVer in error message
			return nil, fmt.Errorf("NewReleaseTree: parent version '%s' for node '%s' not found in input data", input.FromVer, input.Ver)
		}
		newNode.parent = parent
		parent.children = append(parent.children, newNode)
	}

	// Final checks
	if len(inputs) > 0 && tree.root == nil {
		if foundRoots == 0 {
			return nil, errors.New("NewReleaseTree: no root node detected (no node has empty FromVer)")
		}
		return nil, errors.New("NewReleaseTree: tree construction failed, root node is nil despite inputs existing")
	}
	return tree, nil
}

// InsertNode adds a single new release node to the tree concurrently safely.
// Takes ReleaseInput struct (Exported type).
func (tree *ReleaseTree) InsertNode(input ReleaseInput) error {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	// Use Exported field Ver
	if _, exists := tree.nodes[input.Ver]; exists {
		return fmt.Errorf("InsertNode: node with version '%s' already exists", input.Ver)
	}

	var parent *node = nil
	// Use Exported field FromVer
	if input.FromVer == "" {
		if tree.root != nil {
			// Use Exported field Ver
			return fmt.Errorf("InsertNode: cannot insert node '%s' with empty FromVer; tree already has a root ('%s')", input.Ver, tree.root.version)
		}
	} else {
		// Use Exported field FromVer
		p, exists := tree.nodes[input.FromVer]
		if !exists {
			// Use Exported fields Ver, FromVer
			return fmt.Errorf("InsertNode: parent version '%s' for node '%s' not found", input.FromVer, input.Ver)
		}
		parent = p
	}

	// Use Exported field Changes and type Chg
	changesCopy := make([]Chg, len(input.Changes))
	copy(changesCopy, input.Changes)
	newNode := &node{
		version:  input.Ver,   // Use Exported field Ver
		changes:  changesCopy, // Use Exported type Chg
		children: []*node{},
		parent:   parent,
	}

	// Use Exported field Ver
	tree.nodes[newNode.version] = newNode

	if parent != nil {
		parent.children = append(parent.children, newNode)
	} else {
		tree.root = newNode
	}
	return nil
}

// findLCA is the internal (unexported) implementation without locking.
func (tree *ReleaseTree) findLCA(version1, version2 string) (*node, error) {
	node1, exists1 := tree.nodes[version1]
	if !exists1 {
		return nil, fmt.Errorf("findLCA internal: version '%s' not found in tree", version1)
	}
	node2, exists2 := tree.nodes[version2]
	if !exists2 {
		return nil, fmt.Errorf("findLCA internal: version '%s' not found in tree", version2)
	}

	if node1 == node2 {
		return node1, nil
	}

	ancestors := make(map[*node]bool)
	curr := node1
	for curr != nil {
		ancestors[curr] = true
		curr = curr.parent
	}
	curr = node2
	for curr != nil {
		if ancestors[curr] {
			return curr, nil
		}
		curr = curr.parent
	}
	return nil, fmt.Errorf("findLCA internal: logic error: no common ancestor for '%s' and '%s'", version1, version2)
}

// FindLCA finds the version string of the LCA concurrently safely (Exported).
func (tree *ReleaseTree) FindLCA(version1, version2 string) (string, error) {
	tree.mu.RLock()
	defer tree.mu.RUnlock()

	lcaNode, err := tree.findLCA(version1, version2)
	if err != nil {
		return "", fmt.Errorf("FindLCA: failed for versions '%s' and '%s': %w", version1, version2, err)
	}
	return lcaNode.version, nil // Return unexported field
}

// CalcChgs calculates the net changes concurrently safely (Exported).
// Returns []Chg (Exported type).
func (tree *ReleaseTree) CalcChgs(endVersion, startVersion string) ([]Chg, error) {
	tree.mu.RLock()
	defer tree.mu.RUnlock()

	lcaNode, err := tree.findLCA(endVersion, startVersion)
	if err != nil {
		return nil, fmt.Errorf("CalcChgs: failed to find LCA for '%s' and '%s': %w", endVersion, startVersion, err)
	}

	endNode := tree.nodes[endVersion]
	startNode := tree.nodes[startVersion]

	netChanges := make(map[string]Chg) // Use Exported type Chg

	// Accumulate End Path Changes
	curr := endNode
	for curr != nil && curr != lcaNode {
		// Use unexported field changes, containing Exported type Chg
		for _, change := range curr.changes {
			// Use Exported field ID
			netChanges[change.ID] = change
		}
		curr = curr.parent
	}

	// Subtract Start Path Changes (with Subset Check)
	curr = startNode
	for curr != nil && curr != lcaNode {
		// Use unexported field changes, containing Exported type Chg
		for _, change := range curr.changes {
			// Use Exported field ID
			if _, exists := netChanges[change.ID]; !exists {
				// Use unexported field version
				return nil, fmt.Errorf("CalcChgs: change ID '%s' from start path (node '%s', version '%s') not found in end path changes (version '%s' to LCA)",
					change.ID, curr.version, startVersion, endVersion)
			}
			// Use Exported field ID
			delete(netChanges, change.ID)
		}
		curr = curr.parent
	}

	// Format Output
	result := make([]Chg, 0, len(netChanges)) // Use Exported type Chg
	for _, change := range netChanges {
		result = append(result, change)
	}

	// Sort Output
	sort.Slice(result, func(i, j int) bool {
		// Use Exported field ID
		idNumI, errI := strconv.Atoi(result[i].ID)
		idNumJ, errJ := strconv.Atoi(result[j].ID)
		if errI == nil && errJ == nil {
			return idNumI < idNumJ
		}
		// Use Exported field ID
		return result[i].ID < result[j].ID
	})

	return result, nil // Return slice of Exported type Chg
}
