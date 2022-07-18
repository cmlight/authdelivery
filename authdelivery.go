package authdelivery

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/buger/jsonparser"
)

type ParsedBidRequest struct {
	SignatureMessageFragments []string
}

type schainNode struct {
	asi, sid, params, replace []byte

	parsedParams       keyOffsetPairList
	parsedReplacements keyValuePairList
}

type schainNodeList []schainNode

func (l schainNodeList) findReplacement(keyToFind string) (bool, string) {
	for _, laterNode := range l {
		for _, replacement := range laterNode.parsedReplacements {
			if replacement.key == keyToFind {
				return true, replacement.value
			}
		}
	}
	return false, ""
}

type keyValuePair struct {
	key   string
	value string
}

type keyValuePairList []keyValuePair

type keyOffsetPair struct {
	key    string
	offset int
}

type keyOffsetPairList []keyOffsetPair

func (l keyOffsetPairList) findOffsetForKey(key string) int {
	for _, pair := range l {
		if pair.key == key {
			return pair.offset
		}
	}
	return -1
}

func ParseBidRequest(bidRequestJSON []byte) (ParsedBidRequest, error) {
	result := ParsedBidRequest{}

	var globalOffset int
	globalPathIndex := keyOffsetPairList{}

	var innerErr error
	schainNodes := make(schainNodeList, 0, 10)
	pathsToFetch := [][]string{}
	_, err := jsonparser.ArrayEach(bidRequestJSON, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		schainNode := schainNode{
			parsedParams: make(keyOffsetPairList, 0, 10),
		}
		jsonparser.EachKey(value, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
			if err != nil {
				innerErr = err
				return
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
			for _, path := range strings.Split(string(schainNode.params), "&") {
				pair := keyOffsetPair{key: path, offset: globalOffset}
				globalOffset++

				schainNode.parsedParams = append(schainNode.parsedParams, pair)
				globalPathIndex = append(globalPathIndex, pair)
				pathsToFetch = append(pathsToFetch, strings.Split(path, "."))
			}
		}
		if len(schainNode.replace) > 0 {
			for _, replacement := range strings.Split(string(schainNode.replace), "&") {
				replacementParsed, err := url.ParseQuery(replacement)
				if err != nil {
					innerErr = err
					return
				}
				for k, v := range replacementParsed {
					schainNode.parsedReplacements = append(schainNode.parsedReplacements, keyValuePair{key: k, value: v[0]})
				}
			}
		}
		schainNodes = append(schainNodes, schainNode)
	}, "source", "ext", "schain", "nodes")
	if err != nil {
		return result, err
	}
	if innerErr != nil {
		return result, innerErr
	}

	pathsValues := make([][]byte, len(pathsToFetch))
	jsonparser.EachKey(bidRequestJSON, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		if err != nil {
			innerErr = err
			return
		}
		pathsValues[idx] = value
	}, pathsToFetch...)
	if innerErr != nil {
		return result, innerErr
	}

	pathIndex := 0
	for i, schainNode := range schainNodes {
		pairs := []keyValuePair{}

		// first output the signed schain node info
		pairs = append(pairs, keyValuePair{key: fmt.Sprintf("schain.[%d].asi", i), value: string(schainNode.asi)})
		pairs = append(pairs, keyValuePair{key: fmt.Sprintf("schain.[%d].sid", i), value: string(schainNode.sid)})

		// next iterate over the requested protected fields
		for _, protectedParam := range schainNode.parsedParams {
			foundReplacement, valueToUse := schainNodes[i+1:].findReplacement(protectedParam.key)
			if !foundReplacement {
				valueToUse = string(pathsValues[pathIndex])
			}
			pairs = append(pairs, keyValuePair{key: protectedParam.key, value: valueToUse})
			pathIndex++
		}

		// Then iterate over replacements and put in the applicable value.
		for _, replacement := range schainNode.parsedReplacements {
			foundReplacement, valueToUse := schainNodes[i+1:].findReplacement(replacement.key)
			if !foundReplacement {
				valueToUse = string(pathsValues[globalPathIndex.findOffsetForKey(replacement.key)])
			}
			pairs = append(pairs, keyValuePair{key: replacement.key, value: valueToUse})
		}

		// Finally, generate the fragments.
		fragment := ""
		for _, pair := range pairs {
			fragment += "&" + pair.key + "=" + url.QueryEscape(pair.value)
		}

		result.SignatureMessageFragments = append(result.SignatureMessageFragments, fragment)
	}

	return result, nil
}
