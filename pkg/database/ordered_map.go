package database

import (
	"bytes"
	"encoding/json"
)

// OrderedMap represents a map that preserves insertion order.
// It is implemented as a slice of KeyVal pairs to keep it simple and lightweight for this use case.
type KeyVal struct {
	Key string
	Val interface{}
}

type OrderedMap []KeyVal

// MarshalJSON implements the json.Marshaler interface.
func (om OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, kv := range om {
		if i > 0 {
			buf.WriteByte(',')
		}
		// Marshal key
		keyBytes, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		// Marshal value
		valBytes, err := json.Marshal(kv.Val)
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// Get returns the value for a key (O(N) lookup, but explicit for small projections)
func (om OrderedMap) Get(key string) (interface{}, bool) {
	for _, kv := range om {
		if kv.Key == key {
			return kv.Val, true
		}
	}
	return nil, false
}

// ToMap converts to a standard map (losing order)
func (om OrderedMap) ToMap() map[string]interface{} {
	m := make(map[string]interface{}, len(om))
	for _, kv := range om {
		m[kv.Key] = kv.Val
	}
	return m
}

// FromMap creates an OrderedMap from a standard map (arbitrary order)
// This is not usually what we want if we care about order, but useful for compatibility.
func FromMap(m map[string]interface{}) OrderedMap {
	om := make(OrderedMap, 0, len(m))
	for k, v := range m {
		om = append(om, KeyVal{Key: k, Val: v})
	}
	return om
}

// String implements fmt.Stringer
func (om OrderedMap) String() string {
	b, _ := om.MarshalJSON()
	return string(b)
}
