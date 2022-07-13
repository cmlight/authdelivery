package authdelivery

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/buger/jsonparser"
)

func ParseBidRequestOld(bidRequestJSON []byte) error {
	type SchainNode struct {
		asi, sid, params, replace []byte

		paramsDepth         int
		replacementsSmeared map[string]string
		replacementsToEmit  []string
	}

	schainNodes := make([]SchainNode, 0, 10)
	protectedParamsIndex := make([]string, 0, 10)
	_, err := jsonparser.ArrayEach(bidRequestJSON, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var asi, sid, params, replace []byte
		jsonparser.EachKey(value, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
			if err != nil {
				log.Printf("error parsing field: %v", err)
			}
			switch idx {
			case 0:
				asi = value
			case 1:
				sid = value
			case 2:
				params = value
			case 3:
				replace = value
			}
		}, []string{"asi"}, []string{"sid"}, []string{"params"}, []string{"replace"})
		if len(params) > 0 {
			paths := strings.Split(string(params), "&")
			protectedParamsIndex = append(protectedParamsIndex, paths...)
		}
		schainNodes = append(schainNodes,
			SchainNode{
				asi:     asi,
				sid:     sid,
				params:  params,
				replace: replace,

				paramsDepth:         len(protectedParamsIndex),
				replacementsSmeared: make(map[string]string, len(protectedParamsIndex)),
			})
	}, "source", "ext", "schain", "nodes")
	if err != nil {
		log.Printf("error parsing array: %v", err)
		return err
	}
	// Process replacements
	for i := len(schainNodes) - 1; i > 0; i-- {
		// Copy over relevant replacements
		for k, v := range schainNodes[i].replacementsSmeared {
			for j := 0; j < schainNodes[i-1].paramsDepth; j++ {
				if k == protectedParamsIndex[j] {
					schainNodes[i-1].replacementsSmeared[k] = v
				}
			}
		}
		if len(schainNodes[i].replace) > 0 {
			values, err := url.ParseQuery(string(schainNodes[i].replace))
			if err != nil {
				return err
			}
			for k, v := range values {
				schainNodes[i].replacementsToEmit = append(schainNodes[i].replacementsToEmit, k)
				if len(v) == 0 {
					schainNodes[i-1].replacementsSmeared[k] = ""
				} else {
					schainNodes[i-1].replacementsSmeared[k] = v[0] // TODO: handle multiple values
				}
			}
		}
	}

	fmt.Println("schain nodes:")

	for i, v := range schainNodes {
		fmt.Printf("\t%d: %s\n", i, v)
	}

	// protectedPaths := make(map[string]bool, 10)
	// for _, v := range schainNodes {
	// 	if len(v.params) > 0 {
	// 		paths := strings.Split(string(v.params), "&")
	// 		for _, path := range paths {
	// 			protectedPaths[path] = true
	// 		}
	// 	}
	// }
	pathsToFetch := make([][]string, 0, len(protectedParamsIndex))
	for _, path := range protectedParamsIndex {
		elements := strings.Split(path, ".")
		pathsToFetch = append(pathsToFetch, elements)
	}

	fmt.Printf("Fetching paths: %s\n", pathsToFetch)
	fmt.Printf("Path Index: %s\n", protectedParamsIndex)

	pathsValues := make([][]byte, len(protectedParamsIndex))
	jsonparser.EachKey(bidRequestJSON, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		pathsValues[idx] = value
	}, pathsToFetch...)
	fmt.Printf("Path values: %s\n", pathsValues)

	type keyValuePair struct {
		key   string
		value string
	}

	for i, schainNode := range schainNodes {
		fmt.Printf("\t%d: %s\n", i, schainNode)
		pairs := []keyValuePair{}

		// first output the signed schain node info
		pairs = append(pairs, keyValuePair{key: fmt.Sprintf("schain.[%d].asi", i), value: string(schainNode.asi)})
		pairs = append(pairs, keyValuePair{key: fmt.Sprintf("schain.[%d].sid", i), value: string(schainNode.sid)})

		// next iterate over the requested protected fields
		// If there is a smeared replacement, use it; otherwise use the value found in the final bid request
		// for k := schainNode.replacementsToEmit {

		// }

		// Next iterate over the requested replacement fields.  If a smeared value exists, use it; otherwise use the latest info since this is the latest replacement
		fragment := ""
		for _, pair := range pairs {
			fragment += "&" + pair.key + "=" + url.QueryEscape(pair.value)
		}

		fmt.Println("Fragment: " + fragment)
	}

	return nil
}

func ParseBidRequest(bidRequestJSON []byte) error {
	type keyValuePair struct {
		key   string
		value string
	}

	type SchainNode struct {
		asi, sid, params, replace []byte

		parsedParamNames   []string
		parsedReplacements []keyValuePair
	}

	schainNodes := make([]SchainNode, 0, 10)
	_, err := jsonparser.ArrayEach(bidRequestJSON, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		schainNode := SchainNode{
			parsedParamNames:   make([]string, 0, 10),
			parsedReplacements: make([]keyValuePair, 0, 10),			
		}
		jsonparser.EachKey(value, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
			if err != nil {
				log.Printf("error parsing field: %v", err)
			}
			switch idx {
			case 0:
				schainNode.asi = value
			case 1:
				schainNode.sid = value
			case 2:
				schainNode.params = value
			case 3:
				schainNode.replace = value
			}
		}, []string{"asi"}, []string{"sid"}, []string{"params"}, []string{"replace"})
		if len(schainNode.params) > 0 {
			schainNode.parsedParamNames = strings.Split(string(schainNode.params), "&")
		}
		if len(schainNode.replace) > 0 {

		}
		schainNodes = append(schainNodes, schainNode)
	}, "source", "ext", "schain", "nodes")
	if err != nil {
		log.Printf("error parsing array: %v", err)
		return err
	}
	// Process replacements
	for i := len(schainNodes) - 1; i > 0; i-- {
		// Copy over relevant replacements
		for k, v := range schainNodes[i].replacementsSmeared {
			for j := 0; j < schainNodes[i-1].paramsDepth; j++ {
				if k == protectedParamsIndex[j] {
					schainNodes[i-1].replacementsSmeared[k] = v
				}
			}
		}
		if len(schainNodes[i].replace) > 0 {
			values, err := url.ParseQuery(string(schainNodes[i].replace))
			if err != nil {
				return err
			}
			for k, v := range values {
				schainNodes[i].replacementsToEmit = append(schainNodes[i].replacementsToEmit, k)
				if len(v) == 0 {
					schainNodes[i-1].replacementsSmeared[k] = ""
				} else {
					schainNodes[i-1].replacementsSmeared[k] = v[0] // TODO: handle multiple values
				}
			}
		}
	}

	fmt.Println("schain nodes:")

	for i, v := range schainNodes {
		fmt.Printf("\t%d: %s\n", i, v)
	}

	// protectedPaths := make(map[string]bool, 10)
	// for _, v := range schainNodes {
	// 	if len(v.params) > 0 {
	// 		paths := strings.Split(string(v.params), "&")
	// 		for _, path := range paths {
	// 			protectedPaths[path] = true
	// 		}
	// 	}
	// }
	pathsToFetch := make([][]string, 0, len(protectedParamsIndex))
	for _, path := range protectedParamsIndex {
		elements := strings.Split(path, ".")
		pathsToFetch = append(pathsToFetch, elements)
	}

	fmt.Printf("Fetching paths: %s\n", pathsToFetch)
	fmt.Printf("Path Index: %s\n", protectedParamsIndex)

	pathsValues := make([][]byte, len(protectedParamsIndex))
	jsonparser.EachKey(bidRequestJSON, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		pathsValues[idx] = value
	}, pathsToFetch...)
	fmt.Printf("Path values: %s\n", pathsValues)

	for i, schainNode := range schainNodes {
		fmt.Printf("\t%d: %s\n", i, schainNode)
		pairs := []keyValuePair{}

		// first output the signed schain node info
		pairs = append(pairs, keyValuePair{key: fmt.Sprintf("schain.[%d].asi", i), value: string(schainNode.asi)})
		pairs = append(pairs, keyValuePair{key: fmt.Sprintf("schain.[%d].sid", i), value: string(schainNode.sid)})

		// next iterate over the requested protected fields
		// If there is a smeared replacement, use it; otherwise use the value found in the final bid request
		// for k := schainNode.replacementsToEmit {

		// }

		// Next iterate over the requested replacement fields.  If a smeared value exists, use it; otherwise use the latest info since this is the latest replacement
		fragment := ""
		for _, pair := range pairs {
			fragment += "&" + pair.key + "=" + url.QueryEscape(pair.value)
		}

		fmt.Println("Fragment: " + fragment)
	}

	return nil
}
