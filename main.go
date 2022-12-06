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

type Node struct {
	char             string
	children         NodeChildren
	lastLetter       bool
	optionalChildren NodeChildren
	parent           *Node
	weight           int
}

type NodeChildren [40]*Node

var insertWaitGroup sync.WaitGroup

func reverse(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		rns[i], rns[j] = rns[j], rns[i]
	}
	return string(rns)
}

func NewNode(char string, parent *Node) *Node {
	node := &Node{char: char, lastLetter: false, parent: parent, weight: 1}
	for i := 0; i < 40; i++ {
		node.children[i] = nil
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
		node.children[lastIndex].Insert(insert[1:], false)
	} else {
		node.children[lastIndex].lastLetter = true
	}
	if withWaitGroup {
		insertWaitGroup.Done()
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
	sort.Sort(node.children)
	for i := 0; i < 40; i++ {
		node.children[i] = OptimizeSubtrie(node.children[i])
	}
	if node.lastLetter == true {
		// Move everything to optionals
		// fmt.Printf("| Letter %s has optionals\n", node.char)
		for i := 0; i < 40; i++ {
			node.optionalChildren[i] = node.children[i]
			node.children[i] = nil
		}
	}
	return node
}

func PrintChar(char string, file *os.File) {
	if char == "." {
		_, err2 := file.WriteString("\\.")
		if err2 != nil {
			log.Fatal(err2)
		}
	} else if char == "?" {
		_, err2 := file.WriteString("\\?")
		if err2 != nil {
			log.Fatal(err2)
		}
	} else if char == "*" {
		_, err2 := file.WriteString(".")
		if err2 != nil {
			log.Fatal(err2)
		}
	} else {
		_, err2 := file.WriteString(char)
		if err2 != nil {
			log.Fatal(err2)
		}
	}
}

func (node *Node) GenerateRegex(file *os.File) {
	if node == nil {
		return
	}

	optionalParentChildrenSize := 0
	parentChildrenSize := 0
	optionalChildrenSize := 0
	childrenSize := 0

	for i := 0; i < 40; i++ {
		if node.parent != nil {
			if node.parent.optionalChildren[i] != nil && node.parent.optionalChildren[i].char != "\000" {
				optionalParentChildrenSize++
			}
			if node.parent.children[i] != nil && node.parent.children[i].char != "\000" {
				parentChildrenSize++
			}
		}
		if node.optionalChildren[i] != nil && node.optionalChildren[i].char != "\000" {
			optionalChildrenSize++
		}
		if node.children[i] != nil && node.children[i].char != "\000" {
			childrenSize++
		}
	}

	shouldPrintContainerBracket := node.parent != nil && optionalParentChildrenSize > 1

	if shouldPrintContainerBracket {
		_, err2 := file.WriteString("(")
		if err2 != nil {
			log.Fatal(err2)
		}
	}

	if node.char != "\000" {
		PrintChar(node.char, file)
		if node.weight > 1 {
			if node.char == "." || node.char == "?" || node.weight > 5 {
				_, err2 := file.WriteString(fmt.Sprintf("{%d}", node.weight))
				if err2 != nil {
					log.Fatal(err2)
				}
			} else {
				for i := 1; i < node.weight; i++ {
					PrintChar(node.char, file)
				}
			}
		}
	}

	if optionalChildrenSize > 0 {
		if optionalChildrenSize > 1 {
			_, err2 := file.WriteString("(")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		for i, child := range &node.optionalChildren {
			if child == nil {
				break
			}
			if i > 0 {
				_, err2 := file.WriteString("|")
				if err2 != nil {
					log.Fatal(err2)
				}
			}
			child.GenerateRegex(file)
		}
		if optionalChildrenSize > 1 {
			_, err2 := file.WriteString(")")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		_, err2 := file.WriteString("{0,1}")
		if err2 != nil {
			log.Fatal(err2)
		}
	}

	if childrenSize > 1 {
		_, err2 := file.WriteString("(")
		if err2 != nil {
			log.Fatal(err2)
		}
	}
	for i, child := range &node.children {
		if child == nil {
			break
		}
		if i > 0 {
			_, err2 := file.WriteString("|")
			if err2 != nil {
				log.Fatal(err2)
			}
		}
		child.GenerateRegex(file)
	}
	if childrenSize > 1 {
		_, err2 := file.WriteString(")")
		if err2 != nil {
			log.Fatal(err2)
		}
	}

	if shouldPrintContainerBracket {
		_, err2 := file.WriteString(")")
		if err2 != nil {
			log.Fatal(err2)
		}
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
	reversed := len(os.Args) > 1 && os.Args[1] == "--reverse"
	if reversed {
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
		if reversed {
			text = reverse(text)
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
	if reversed {
		fileName = "output/regex-reversed.txt"
	} else {
		fileName = "output/regex.txt"
	}
	outputFile, err := os.Create(fileName)
	root.GenerateRegex(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	println("| Finished writing file")

	outputFile.Close()
	fmt.Println("done")
}
