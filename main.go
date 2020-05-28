package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/beevik/etree"
	"github.com/scylladb/go-set/strset"
)

func walkXML(e *etree.Element, af func(*etree.Element, etree.Attr)) {
	for _, a := range e.Attr {
		af(e, a)
	}
	for _, c := range e.ChildElements() {
		walkXML(c, af)
	}
}

func main() {

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Analyzing: %s\n", wd)

	xmlnsDeclarations := make(map[string]*strset.Set)

	saveDeclaration := func(prefix string, namespaceURI string) {
		if xmlnsDeclarations[namespaceURI] == nil {
			xmlnsDeclarations[namespaceURI] = strset.New(prefix)
		} else {
			xmlnsDeclarations[namespaceURI].Add(prefix)
		}
	}

	extractXmlnsDeclarations := func(path string, err error) error {
		doc := etree.NewDocument()
		err = doc.ReadFromFile(path)
		if err != nil {
			return err
		}

		root := doc.Root()
		walkXML(root, func(e *etree.Element, a etree.Attr) {
			if a.Space == "xmlns" {
				saveDeclaration(a.Key, a.Value)
			}
		})
		return nil
	}

	err = filepath.Walk(wd,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".xml" {
				return extractXmlnsDeclarations(path, err)
			}
			return nil
		})
	if err != nil {
		panic(err)
	}

	for namespace, prefixes := range xmlnsDeclarations {
		log.Printf("[%v] \"%s\" declared as %s\n",
			map[bool]string{true: "-OK!-", false: "-ERR-"}[prefixes.Size() == 1],
			namespace,
			prefixes)
	}
}
