package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/beevik/etree"
	"github.com/scylladb/go-set/strset"
)

var wg sync.WaitGroup
var xmlnsDeclarations = make(map[string]*strset.Set)

func walkXML(e *etree.Element, af func(*etree.Element, etree.Attr)) {
	for _, a := range e.Attr {
		af(e, a)
	}
	for _, c := range e.ChildElements() {
		walkXML(c, af)
	}
}

var lock = sync.Mutex{}

func saveDeclaration(prefix string, namespaceURI string) {
	lock.Lock()
	defer lock.Unlock()

	if xmlnsDeclarations[namespaceURI] == nil {
		xmlnsDeclarations[namespaceURI] = strset.New(prefix)
	} else {
		xmlnsDeclarations[namespaceURI].Add(prefix)
	}
}

func extractXmlnsDeclarations(path string) {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		panic(err)
	}

	root := doc.Root()
	walkXML(root, func(e *etree.Element, a etree.Attr) {
		if a.Space == "xmlns" {
			saveDeclaration(a.Key, a.Value)
		}
	})
}

func walkDir(dir string) {
	defer wg.Done()

	visit := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() && path != dir {
			wg.Add(1)
			go walkDir(path)
			return filepath.SkipDir
		}
		if f.Mode().IsRegular() && filepath.Ext(f.Name()) == ".xml" {
			extractXmlnsDeclarations(path)
		}
		return nil
	}

	filepath.Walk(dir, visit)
}

func main() {

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Analyzing: %s\n", workingDir)

	wg.Add(1)
	walkDir(workingDir)
	wg.Wait()

	for namespace, prefixes := range xmlnsDeclarations {
		log.Printf("[%v] \"%s\" declared as %s\n",
			map[bool]string{true: "-OK!-", false: "-ERR-"}[(prefixes).Size() == 1],
			namespace,
			prefixes)
	}
}
