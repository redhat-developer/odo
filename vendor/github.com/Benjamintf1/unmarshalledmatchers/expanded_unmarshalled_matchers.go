package unmarshalledmatchers

import (
	"github.com/onsi/gomega/types"
	"fmt"
)

//This is for json with Unordered lists ie [1,2,3] is equal to [2,3,1].
//If you want to have some lists enforce order, add keys exclusions using
//Add WithOrderedListKeys( json keys that refer to unordered lists )
func MatchUnorderedJSON(json interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: false,
		Subset:  false,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		} else {
			fmt.Errorf("You are trying to set unordered list keys for unordered JSON")
		}
	}


	return &ExpandedJsonMatcher{
		JSONToMatch: json,
		DeepMatcher: deepMatcher,
	}
}

//This is for json with Ordered lists ie [1,2,3] is not equal to [2,3,1].
//This is just like the default match json.
//If you want to have some lists enforce order, add keys exclusions using
//Add WithUnorderedListKeys( json keys that refer to unordered lists )
func MatchOrderedJSON(json interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: true,
		Subset:  false,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			fmt.Errorf("You are trying to set ordered list keys for ordered JSON")
		} else {
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		}
	}


	return &ExpandedJsonMatcher{
		JSONToMatch: json,
		DeepMatcher: deepMatcher,
	}
}

//This is for json with Unordered lists ie [1,2,3] is equal to [2,3,1].
//This also is a subset match rather then a full match ie [1,2,3] contains [1,2]
//If you want to have some lists enforce order, add keys exclusions using
//Add WithOrderedListKeys( json keys that refer to unordered lists )
func ContainUnorderedJSON(json interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: false,
		Subset:  true,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		} else {
			fmt.Errorf("You are trying to set unordered list keys for unordered JSON")
		}
	}


	return &ExpandedJsonMatcher{
		JSONToMatch: json,
		DeepMatcher: deepMatcher,
	}
}

//This is for json with Ordered lists ie [1,2,3] is not equal to [2,3,1].
//This is just like the default match json.
//This also is a subset match rather then a full match ie [1,2,3] contains [1,2]
//If you want to have some lists enforce order, add keys exclusions using
//Add WithUnorderedListKeys( json keys that refer to unordered lists )
func ContainOrderedJSON(json interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: true,
		Subset:  true,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			fmt.Errorf("You are trying to set ordered list keys for ordered JSON")
		} else {
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		}
	}


	return &ExpandedJsonMatcher{
		JSONToMatch: json,
		DeepMatcher: deepMatcher,
	}
}

//This is for yaml with Unordered lists ie [1,2,3] is equal to [2,3,1].
//If you want to have some lists enforce order, add keys exclusions using
//Add WithOrderedListKeys( YAML keys that refer to unordered lists )
func MatchUnorderedYAML(YAML interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: false,
		Subset:  false,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		} else {
			fmt.Errorf("You are trying to set unordered list keys for unordered YAML")
		}
	}


	return &ExpandedYAMLMatcher{
		YAMLToMatch: YAML,
		DeepMatcher: deepMatcher,
	}
}

//This is for yaml with Ordered lists ie [1,2,3] is not equal to [2,3,1].
//This is just like the default match yaml.
//If you want to have some lists enforce order, add keys exclusions using
//Add WithUnorderedListKeys( YAML keys that refer to unordered lists )
func MatchOrderedYAML(YAML interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: true,
		Subset:  false,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			fmt.Errorf("You are trying to set ordered list keys for ordered YAML")
		} else {
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		}
	}


	return &ExpandedYAMLMatcher{
		YAMLToMatch: YAML,
		DeepMatcher: deepMatcher,
	}
}

//This is for yaml with Unordered lists ie [1,2,3] is equal to [2,3,1].
//This also is a subset match rather then a full match ie [1,2,3] contains [1,2]
//If you want to have some lists enforce order, add keys exclusions using
//Add WithOrderedListKeys( YAML keys that refer to unordered lists )
func ContainUnorderedYAML(YAML interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: false,
		Subset:  true,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		} else {
			fmt.Errorf("You are trying to set unordered list keys for unordered YAML")
		}
	}


	return &ExpandedYAMLMatcher{
		YAMLToMatch: YAML,
		DeepMatcher: deepMatcher,
	}
}

//This is for yaml with Ordered lists ie [1,2,3] is not equal to [2,3,1].
//This is just like the default match yaml.
//This also is a subset match rather then a full match ie [1,2,3] contains [1,2]
//If you want to have some lists enforce order, add keys exclusions using
//Add WithUnorderedListKeys( YAML keys that refer to unordered lists )
func ContainOrderedYAML(YAML interface{}, keys ...KeyExclusions) types.GomegaMatcher {
	deepMatcher := UnmarshalledDeepMatcher{
		Ordered: true,
		Subset:  true,
	}

	if len(keys) > 0{
		if len(keys) > 1 {
			fmt.Errorf("Only 1 key exclusion set is currently supported")
		} else if keys[0].IsOrdered(){
			fmt.Errorf("You are trying to set ordered list keys for ordered YAML")
		} else {
			deepMatcher.InvertOrderingKeys = keys[0].GetMap()
		}
	}


	return &ExpandedYAMLMatcher{
		YAMLToMatch: YAML,
		DeepMatcher: deepMatcher,
	}
}


type OrderedKeys struct {
	val map[interface{}]bool
}

func NewOrderedKeys() OrderedKeys {
	return OrderedKeys{
		val: make(map[interface{}]bool),
	}
}

func (k OrderedKeys) IsOrdered() bool {
	return true;
}

func (k OrderedKeys) GetMap() map[interface{}]bool {
	return k.val;
}

type UnorderedKeys struct {
	val map[interface{}]bool
}

func NewUnorderedKeys() UnorderedKeys {
	return UnorderedKeys{
		val: make(map[interface{}]bool),
	}
}

func (k UnorderedKeys) IsOrdered() bool {
	return false;
}

func (k UnorderedKeys) GetMap() map[interface{}]bool {
	return k.val;
}

type KeyExclusions interface {
	IsOrdered() bool
	GetMap() map[interface{}]bool
}

func WithOrderedListKeys(keys ...interface{}) OrderedKeys{
	ok := NewOrderedKeys()

	for _, v := range keys {
		ok.val[v] = true
	}

	return ok
}

func WithUnorderedListKeys(keys ...interface{}) UnorderedKeys{
	uk := NewUnorderedKeys()

	for _, v := range keys {
		uk.val[v] = true
	}

	return uk
}