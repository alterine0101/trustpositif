package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

type Node struct {
	char             string
	children         NodeChildren
	optionalChildren NodeChildren
	parent           *Node
	weight           int
}

type NodeChildren [40]*Node

func NewNode(char string, parent *Node) *Node {
	node := &Node{char: char, parent: parent, weight: 1}
	for i := 0; i < 40; i++ {
		node.children[i] = nil
	}
	return node
}

func (node *Node) Insert(insert string) {
	if node == nil {
		node = NewNode(insert, nil)
	}
	if len(insert) == 0 {
		return
	}
	first := insert[0:1]
	lastIndex := -1
	found := false
	for i, child := range &node.children {
		lastIndex = i
		if child == nil {
			break
		}
		if child.char == first {
			found = true
			break
		}
	}
	if !found {
		node.children[lastIndex] = NewNode(first, node)
	}
	if len(insert) > 1 {
		node.children[lastIndex].Insert(insert[1:])
	}
}

func OptimizeSubtrie(node *Node) *Node {
	if node == nil {
		return nil
	}
	cursor := *node
DO:
	childrenSize := cursor.children.Len()
	currentWeight := cursor.weight
	if childrenSize == 1 && cursor.children[0].char == node.char {
		cursor = *cursor.children[0]
		currentWeight += cursor.weight
		cursor.weight = currentWeight
		goto DO
	} else {
		goto DONE
	}
DONE:
	node = &cursor
	for i := 0; i < 40; i++ {
		node.children[i] = OptimizeSubtrie(node.children[i])
	}
	sort.Sort(node.children)
	return node
}

func (node *Node) GenerateRegex() string {
	if node == nil {
		return ""
	}

	res := ""

	optionalChildrenSize := node.optionalChildren.Len()
	childrenSize := node.children.Len()
	shouldPrintContainerBracket := node.parent != nil && node.parent.children.Len() > 1

	if shouldPrintContainerBracket {
		res += "("
	}

	if node.char != "\000" {
		if node.char == "." {
			res += "\\."
		} else if node.char == "?" {
			res += "\\?"
		} else if node.char == "*" {
			res += "."
		} else {
			res += node.char
		}
		if node.weight > 1 {
			res += fmt.Sprintf("{%d}", node.weight)
		}
	}

	if optionalChildrenSize > 0 {
		if optionalChildrenSize > 1 {
			res += "("
		}
		for i, child := range &node.optionalChildren {
			if child == nil {
				break
			}
			if i > 0 {
				res += "|"
			}
			res += child.GenerateRegex()
		}
		if optionalChildrenSize > 1 {
			res += ")"
		}
		res += "{0,1}"
	}

	if childrenSize > 1 {
		res += "("
	}
	for i, child := range &node.children {
		if child == nil {
			break
		}
		if i > 0 {
			res += "|"
		}
		res += child.GenerateRegex()
	}
	if childrenSize > 1 {
		res += ")"
	}

	if shouldPrintContainerBracket {
		res += ")"
	}

	return res
}

var root *Node = NewNode("\000", nil)

/* Sorting utilities */
func (children NodeChildren) Len() int {
	res := 0
	for _, child := range &children {
		if child != nil {
			res++
		}
	}
	return res
}

func (children NodeChildren) Less(i, j int) bool {
	return children[i].char > children[j].char
}

func (children NodeChildren) Swap(i, j int) {
	temp := *children[i]
	children[i] = children[j]
	children[j] = &temp
}

func main() {
	/* Stage 1: Import all domains list */
	inputFile, err := os.Open("input/domains")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	scanner := bufio.NewScanner(inputFile)
	domains := 0

	for scanner.Scan() {
		domains++
		text := scanner.Text()
		text = strings.ToLower(text)
		root.Insert(text)

		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	}
	inputFile.Close()
	// fmt.Printf("| %d domains imported to the freakin' huge trie\n", domains)

	/* Stage 2: Simplify */
	root := OptimizeSubtrie(root)
	regex := root.GenerateRegex()

	outputFile, err := os.Create("output/regex.txt")

	if err != nil {
		log.Fatal(err)
	}
	_, err2 := outputFile.WriteString(regex)

	if err2 != nil {
		log.Fatal(err2)
	}

	outputFile.Close()
	fmt.Println("done")
}
