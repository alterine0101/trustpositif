package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
)

// Will be parsed as -[char]-[requiredChildren]-[optionalChildren]-[nextChildren]-
type Node struct {
	char             string
	nextChildren     NodeChildren
	lastLetter       bool
	optionalChildren *NodeChildren
	requiredChildren *NodeChildren
	parent           *Node
	weight           int
}

// 26: a-z
// 10: 0-9
// 04: .:-_
// With additional 4
// Total: 44
type NodeChildren [44]*Node

var insertWaitGroup sync.WaitGroup

func __(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ___[T interface{}](res T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func Reverse(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		rns[i], rns[j] = rns[j], rns[i]
	}
	return string(rns)
}

func NewNode(char string, parent *Node) *Node {
	node := &Node{char: char, lastLetter: false, parent: parent, weight: 1}
	for i := 0; i < 44; i++ {
		node.nextChildren[i] = nil
	}
	return node
}

func (node *Node) Insert(insert string, withWaitGroup bool) {
	if withWaitGroup {
		insertWaitGroup.Add(1)
	}
	first := insert[0:1]
	if len(insert) == 0 {
		return
	}
	if node == nil {
		node = NewNode("\000", nil)
	}
	lastIndex := -1
	found := false
	for i, child := range node.nextChildren {
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
		node.nextChildren[lastIndex] = NewNode(first, node)
	}
	if len(insert) > 1 {
		node.nextChildren[lastIndex].Insert(insert[1:], false)
	} else {
		node.nextChildren[lastIndex].lastLetter = true
	}
	if withWaitGroup {
		insertWaitGroup.Done()
	}
}

// Override a map of characters
func GetEndings(node *Node, data *map[string]int) {
	for i := 0; i < 44; i++ {
		if node.nextChildren[i] == nil {
			continue
		} else if node.lastLetter {
			(*data)[node.char]++
		}
		GetEndings(node.nextChildren[i], data)
	}
}

// Remove endings
func RemoveEndings(node *Node) {
	if node == nil || (node.lastLetter && node.nextChildren.Len() == 0) {
		node.parent.lastLetter = true
	}
	for i := 0; i < 44; i++ {
		RemoveEndings(node.nextChildren[i])
	}
}

func OptimizeSubtrie(node *Node) *Node {
	if node == nil {
		return nil
	}
	cursor := *node
DO:
	childrenSize := cursor.nextChildren.Len()
	currentWeight := cursor.weight
	// Simplify -[a(1)]-[a(1)]- to become -[a(2)]-
	if childrenSize == 1 && cursor.nextChildren[0].char == node.char {
		cursor = *cursor.nextChildren[0]
		currentWeight += cursor.weight
		cursor.weight = currentWeight
		goto DO
	} else {
		goto DONE
	}
DONE:
	node = &cursor
	sort.Sort(node.nextChildren)
	for i := 0; i < 44; i++ {
		node.nextChildren[i] = OptimizeSubtrie(node.nextChildren[i])
	}
	if node.lastLetter == true {
		// Move everything to optionals
		// -[a(1)]-[END]
		//     +---[b(1)]-
		// fmt.Printf("| Letter %s has optionals\n", node.char)
		node.optionalChildren = &NodeChildren{}
		for i := 0; i < 44; i++ {
			node.optionalChildren[i] = node.nextChildren[i]
			node.nextChildren[i] = nil
		}
	} else {
		// Attempt to simplify ending
		// -[a(1)]-[b(1)]-[d(2)]
		//     +---[c(1)]-[d(2)]
		// becoming:
		// -[a(1)]-[b(1)]-[d(2)]
		//     +---[c(1)]---+
		// by using requiredChildren
		ending := make(map[string]int)
		GetEndings(node, &ending)
		keys := make([]string, 0, len(ending))
		for k := range ending {
			keys = append(keys, k)
		}
		if len(keys) == 1 {
			// Move all nextChildren into requiredChildren
			node.requiredChildren = &NodeChildren{}
			node.requiredChildren = &node.nextChildren
			for l := 0; l < 44; l++ {
				RemoveEndings(node.requiredChildren[l])
			}
			node.nextChildren = NodeChildren{NewNode(keys[0], node)}
		}
	}
	return node
}

func PrintChar(char string, file *os.File) {
	if char == "." {
		___(file.WriteString("\\."))
	} else if char == "?" {
		___(file.WriteString("\\?"))
	} else if char == "*" {
		___(file.WriteString("."))
	} else {
		___(file.WriteString(char))
	}
}

func (node *Node) GenerateRegex(file *os.File) {
	if node == nil {
		return
	}

	optionalParentChildrenSize := 0
	requiredParentChildrenSize := 0
	parentChildrenSize := 0
	optionalChildrenSize := 0
	requiredChildrenSize := 0
	childrenSize := 0

	for i := 0; i < 44; i++ {
		if node.parent != nil {
			if node.parent.optionalChildren != nil && node.parent.optionalChildren[i] != nil && node.parent.optionalChildren[i].char != "\000" {
				optionalParentChildrenSize++
			}
			if node.parent.requiredChildren != nil && node.parent.requiredChildren[i] != nil && node.parent.requiredChildren[i].char != "\000" {
				requiredParentChildrenSize++
			}
			if node.parent.nextChildren[i] != nil && node.parent.nextChildren[i].char != "\000" {
				parentChildrenSize++
			}
		}
		if node.optionalChildren != nil && node.optionalChildren[i] != nil && node.optionalChildren[i].char != "\000" {
			optionalChildrenSize++
		}
		if node.requiredChildren != nil && node.requiredChildren[i] != nil && node.requiredChildren[i].char != "\000" {
			requiredChildrenSize++
		}
		if node.nextChildren[i] != nil && node.nextChildren[i].char != "\000" {
			childrenSize++
		}
	}

	shouldPrintContainerBracket := node.parent != nil && (requiredParentChildrenSize > 1 || optionalParentChildrenSize > 1)

	if shouldPrintContainerBracket {
		___(file.WriteString("("))
	}

	if node.char != "\000" {
		PrintChar(node.char, file)
		if node.weight > 1 {
			if node.char == "." || node.char == "?" || node.weight > 5 {
				___(file.WriteString(fmt.Sprintf("{%d}", node.weight)))
			} else {
				for i := 1; i < node.weight; i++ {
					PrintChar(node.char, file)
				}
			}
		}
	}

	// Print required children
	if requiredChildrenSize > 0 {
		if requiredChildrenSize > 1 {
			___(file.WriteString("("))
		}
		for i, child := range node.optionalChildren {
			if child == nil {
				break
			}
			if i > 0 {
				___(file.WriteString("|"))
			}
			child.GenerateRegex(file)
		}
		if requiredChildrenSize > 1 {
			___(file.WriteString(")"))
		}
	}

	// Print optional children
	if optionalChildrenSize > 0 {
		if optionalChildrenSize > 1 {
			___(file.WriteString("("))
		}
		for i, child := range node.optionalChildren {
			if child == nil {
				break
			}
			if i > 0 {
				___(file.WriteString("|"))
			}
			child.GenerateRegex(file)
		}
		if optionalChildrenSize > 1 {
			___(file.WriteString(")"))
		}
		___(file.WriteString("{0,1}"))
	}

	if childrenSize > 1 {
		___(file.WriteString("("))
	}
	for i, child := range node.nextChildren {
		if child == nil {
			break
		}
		if i > 0 {
			___(file.WriteString("|"))
		}
		child.GenerateRegex(file)
	}
	if childrenSize > 1 {
		___(file.WriteString(")"))
	}

	if shouldPrintContainerBracket {
		___(file.WriteString(")"))
	}
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
	Reversed := len(os.Args) > 1 && os.Args[1] == "--Reverse"
	if Reversed {
		println("* Reverse trie used")
	}

	/* Stage 1: Import all domains list */
	println("* Stage 1 start")
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
		if Reversed {
			text = Reverse(text)
		}
		go root.Insert(text, true)
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
		if domains%250000 == 0 {
			fmt.Printf("| %d domains imported to the freakin' huge trie\n", domains)
		}
	}
	fmt.Printf("| %d domains imported to the freakin' huge trie\n", domains)
	inputFile.Close()
	insertWaitGroup.Wait()
	fmt.Println("| All goroutines have been cleared")

	/* Stage 2: Simplify */
	println("* Stage 2 start")
	root := OptimizeSubtrie(root)

	/* Stage 3: Generate Regex */
	println("* Stage 3 start")
	var fileName string
	if Reversed {
		fileName = "output/regex-Reversed.txt"
	} else {
		fileName = "output/regex.txt"
	}
	outputFile, err := os.Create(fileName)
	root.GenerateRegex(outputFile)
	println("| Finished writing file")

	outputFile.Close()
	fmt.Println("done")
}
