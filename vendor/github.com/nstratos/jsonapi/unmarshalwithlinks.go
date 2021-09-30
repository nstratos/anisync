package jsonapi

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// UnmarshalManyPayloadWithLinks is a copy of UnmarshalManyPayload with the
// only difference that it also returns a map of the pagination links which are
// included in the JSON API document. The map is used to parse the offset from
// the links and return it to the user in a convenient way.
func UnmarshalManyPayloadWithLinks(in io.Reader, t reflect.Type) ([]interface{}, *Links, error) {
	payload := new(ManyPayload)

	if err := json.NewDecoder(in).Decode(payload); err != nil {
		return nil, nil, err
	}

	models := []interface{}{}         // will be populated from the "data"
	includedMap := map[string]*Node{} // will be populate from the "included"

	if payload.Included != nil {
		for _, included := range payload.Included {
			key := fmt.Sprintf("%s,%s", included.Type, included.ID)
			includedMap[key] = included
		}
	}

	for _, data := range payload.Data {
		model := reflect.New(t.Elem())
		err := unmarshalNode(data, model, &includedMap)
		if err != nil {
			return nil, nil, err
		}
		models = append(models, model.Interface())
	}

	return models, payload.Links, nil
}
