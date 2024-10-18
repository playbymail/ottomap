// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// hexReportToNodes converts a hex report into a linked list of nodes
// where each node contains all the arguments for each component of
// the hex report.
func hexReportToNodes(hexReport []byte, debugNodes bool, experimentalUnitSplit bool) (root *node) {
	if debugNodes {
		log.Printf("parser: root: before split %s\n", string(hexReport))
	}

	var tail *node
	for _, component := range bytes.Split(hexReport, []byte{','}) {
		if component = bytes.TrimSpace(component); len(component) != 0 {
			if root == nil {
				// there's a bug in fleet movement reports where the direction-terrain is not followed
				// by a comma if the only substep is a settlement. try to tease that out here.
				if isDirDashTerrain(component) && bytes.IndexByte(component, ' ') != -1 {
					dirTerrain, maybeSettlement, _ := bytes.Cut(component, []byte{' '})
					//log.Printf("parser: root: maybe dirTerrain %s\n", string(dirTerrain))
					//log.Printf("parser: root: maybe settlement %s\n", string(maybeSettlement))
					root = &node{
						text: dirTerrain,
						next: &node{
							text: bytes.TrimSpace(maybeSettlement),
						},
					}
					tail = root.next
				} else {
					root = &node{text: component}
					tail = root
				}
			} else { // tail can't be nil if root is set
				tail.next = &node{text: component}
				tail = tail.next
			}
		}
	}

	if debugNodes {
		log.Printf("parser: root: after split %s\n", printNodes(root))
	}

	// experimental: if the last node in a list is a unit, split it out
	if experimentalUnitSplit {
		foundUnits := 0
		for tmp := root; tmp != nil; {
			// no good solution for patrols
			if bytes.HasPrefix(tmp.text, []byte("Patrolled and found")) {
				tmp = tmp.next
				continue
			}
			// does this node start with text and end with a unit?
			matches := rxTextUnitId.FindSubmatch(tmp.text)
			// move on to the next if it doesn't
			if len(matches) != 4 {
				tmp = tmp.next
				continue
			}
			foundUnits++
			// otherwise, create a new node with the text and unit
			text, unit := matches[1], matches[2]
			newNode := &node{text: unit, next: tmp.next}
			// update this node with the trimmed text
			tmp.text = text
			// insert the new node after this one
			tmp.next = newNode
			// stay on this node because we may have multiple units at the end
		}
		if foundUnits != 0 {
			log.Printf("parser: experiment: %d units split %s\n", foundUnits, printNodes(root))
		}
	}

	// splitting like that has broke some things.
	// there are components that use commas as separators internally.
	// we need to find them and splice them back together. brute force it.
	for tmp := root; tmp != nil && tmp.next != nil; {
		if tmp.isFindQuantityItem() {
			for tmp.next.isQuantityItem() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isFordEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isLakeEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isLowConiferMountainsEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isLowSnowyMountainsEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isHighSnowyMountainsEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isLowJungleMountainsEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isOceanEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isPassEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isPatrolledAndFound() {
			for tmp.next.isUnitId() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isRiverEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isStoneRoadEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isUnitId() && tmp.next.isUnitId() {
			for tmp.next.isUnitId() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		}
		tmp = tmp.next
	}

	if debugNodes {
		log.Printf("parser: root: after consolidating %s\n", printNodes(root))
	}

	return root
}

func nodesToSteps(n *node) ([][]byte, error) {
	if n == nil {
		return nil, nil
	}
	var steps [][]byte
	for n != nil {
		text := bytes.TrimSpace(n.text)
		if len(text) != 0 {
			steps = append(steps, text)
		}
		n = n.next
	}
	return steps, nil
}

func printNodes(root *node) string {
	if root == nil {
		return "<nil>"
	}
	sb := &strings.Builder{}
	sb.WriteString("(")
	for n := root; n != nil; n = n.next {
		sb.WriteString("\n\t")
		sb.WriteString(n.String())
	}
	sb.WriteString("\n)")
	return sb.String()
}

type node struct {
	text []byte
	next *node
}

func (n *node) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("(%s)", string(n.text))
}

func (n *node) addText(t *node) {
	if t == nil && len(t.text) == 0 {
		return
	}
	if len(n.text) != 0 {
		n.text = append(n.text, ' ')
	}
	n.text = append(n.text, t.text...)
}

// isDirection returns true if the node is a direction.
// we have to do a case-insensitive comparison because the grammar is case-insensitive,
// but we don't coerce the returned tokens to upper case.
func (n *node) isDirection() bool {
	if n == nil {
		return false
	} else if bytes.EqualFold(n.text, []byte{'N'}) {
		return true
	} else if bytes.EqualFold(n.text, []byte{'N', 'E'}) {
		return true
	} else if bytes.EqualFold(n.text, []byte{'S', 'E'}) {
		return true
	} else if bytes.EqualFold(n.text, []byte{'S'}) {
		return true
	} else if bytes.EqualFold(n.text, []byte{'S', 'W'}) {
		return true
	} else if bytes.EqualFold(n.text, []byte{'N', 'W'}) {
		return true
	}
	return false
}

func (n *node) isFindQuantityItem() bool {
	if n == nil {
		return false
	}
	return rxFindQuantityItem.Match(n.text)
}

func (n *node) isFordEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'F', 'o', 'r', 'd', ' '})
}

func (n *node) isLakeEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'L', ' '})
}

func (n *node) isLowConiferMountainsEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'L', 'c', 'm', ' '})
}

func (n *node) isLowSnowyMountainsEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'L', 's', 'm', ' '})
}

func (n *node) isHighSnowyMountainsEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'H', 's', 'm', ' '})
}

func (n *node) isLowJungleMountainsEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(bytes.ToUpper(n.text), []byte{'L', 'J', 'M', ' '})
}

func (n *node) isOceanEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'O', ' '})
}

func (n *node) isPassEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'P', 'a', 's', 's', ' '})
}

func (n *node) isPatrolledAndFound() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'P', 'a', 't', 'r', 'o', 'l', 'l', 'e', 'd', ' ', 'a', 'n', 'd', ' ', 'f', 'o', 'u', 'n', 'd', ' '})
}

func (n *node) isQuantityItem() bool {
	if n == nil {
		return false
	}
	return rxQuantityItem.Match(n.text)
}

func (n *node) isRiverEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'R', 'i', 'v', 'e', 'r', ' '})
}

func (n *node) isStoneRoadEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'S', 't', 'o', 'n', 'e', ' ', 'R', 'o', 'a', 'd', ' '})
}

func (n *node) isUnitId() bool {
	if n == nil {
		return false
	}
	return rxUnitId.Match(n.text)
}

var (
	rxFindQuantityItem = regexp.MustCompile(`^Find [0-9]+ [a-zA-Z][a-zA-Z ]+`)
	rxQuantityItem     = regexp.MustCompile(`^[0-9]+ [a-zA-Z][a-zA-Z ]+`)
	rxUnitId           = regexp.MustCompile(`^[0-9][0-9][0-9][0-9]([cefg][0-9])?`)
	rxTextUnitId       = regexp.MustCompile(`^(.*)\s+([0-9][0-9][0-9][0-9]([cefg][0-9])?)$`)
)

func isDirDashTerrain(text []byte) bool {
	if bytes.HasPrefix(text, []byte{'N', '-'}) {
		return true
	} else if bytes.HasPrefix(text, []byte{'N', 'E', '-'}) {
		return true
	} else if bytes.HasPrefix(text, []byte{'S', 'E', '-'}) {
		return true
	} else if bytes.HasPrefix(text, []byte{'S', '-'}) {
		return true
	} else if bytes.HasPrefix(text, []byte{'S', 'W', '-'}) {
		return true
	} else if bytes.HasPrefix(text, []byte{'N', 'W', '-'}) {
		return true
	}
	return false
}

// trimUnit is a placeholder for your function that returns the updated text and the unit
func trimUnit(input []byte) (text, unit []byte) {
	matches := rxTextUnitId.FindSubmatch(input)
	if len(matches) != 3 {
		return input, nil
	}
	return matches[1], matches[2]
}
