package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dlclark/regexp2"
)

func TestRegexResult(t *testing.T) {
	/* Stage 1: Import all domains list */
	t.Log("Opening file...")
	inputFile, err := os.Open("input/domains")
	regexFile, err2 := ioutil.ReadFile("output/regex.txt")
	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}
	if err2 != nil {
		t.Log(err2.Error())
		t.Fail()
		return
	}

	scanner := bufio.NewScanner(inputFile)
	correct := 0
	domains := 0
	var undetected []string

	t.Log("Compiling regex...")
	regex := string(regexFile)
	matcher := regexp2.MustCompile(regex, 0)
	t.Log("Starting validation...")

	for scanner.Scan() {
		domain := scanner.Text()
		val, err := matcher.MatchString(domain)
		if val {
			correct++
		} else {
			undetected = append(undetected, domain)
		}
		err = scanner.Err()
		domains++
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if domains%50000 == 0 {
			t.Logf("%d out of %d correct (%.2f percent)", correct, domains, float32(correct)/float32(domains)*100)
		}
	}
	inputFile.Close()
	t.Log("done")
}
