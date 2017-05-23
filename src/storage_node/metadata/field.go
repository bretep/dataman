package metadata

import (
	"fmt"
	"strconv"
)

type FieldType string

// TODO: re-work to have multiple mappings
// The intention here is to have a mapping of client -> dataman -> datastore
// this should be our listing of dataman FieldTypes, which have limits and validation methods
// which we then leave up to the datasource to store.
const (
	Document FieldType = "document"
	String             = "string"
	Text               = "text"
	Int                = "int"
	Bool               = "bool"
	DateTime           = "datetime"
)

type Field struct {
	ID int64 `json:"_id,omitempty"`
	// TODO: remove? Need a method to link them
	CollectionID  int64     `json:"-"`
	ParentFieldID int64     `json:"-"`
	Name          string    `json:"name"`
	Type          FieldType `json:"type"`
	// Arguments (limits etc.) for a given FieldType (varies per field)
	TypeArgs map[string]interface{} `json:"type_args,omitempty"`

	// Various configuration options
	NotNull bool `json:"not_null,omitempty"` // Should we allow NULL fields

	// Optional subfields
	SubFields map[string]*Field `json:"subfields,omitempty"`

	// Optional relation
	Relation *FieldRelation `json:"relation,omitempty"`

	ProvisionState ProvisionState `json:"provision_state"`
}

func (f *Field) Equal(o *Field) bool {
	// TODO: better?
	return f.Name == o.Name && f.Type == o.Type && f.NotNull == o.NotNull
}

// Validate a field
func (f *Field) Validate(val interface{}) error {
	switch f.Type {
	case Document:
		valTyped, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Not a document")
		}

		// TODO: We need to check that we where given no more than the subFields we know about
		// TODO: add "strict" arg to typeArgs
		for k, subField := range f.SubFields {
			if v, ok := valTyped[k]; ok {
				if err := subField.Validate(v); err != nil {
					return err
				}
			} else {
				if subField.NotNull {
					return fmt.Errorf("Missing required subfield %s", k)
				}
			}
		}
		return nil
	case String:
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("Not a string")
		}
		if float64(len(s)) > f.TypeArgs["size"].(float64) {
			return fmt.Errorf("String too long")
		}
		return nil
	case Text:
		_, ok := val.(string)
		if !ok {
			return fmt.Errorf("Not a string")
		}
		return nil
	case Int:
		switch typedVal := val.(type) {
		case int:
			return nil
		case float64:
			return nil
		case string:
			_, err := strconv.ParseInt(typedVal, 10, 64)
			return err
		default:
			return fmt.Errorf("Unknown Int type")
		}
	case Bool:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("Not a bool")
		}
		return nil
	// TODO: implement
	case DateTime:
	}

	return fmt.Errorf("Unknown type \"%s\" defined", f.Type)
}

type FieldRelation struct {
	ID      int64 `json:"_id,omitempty"`
	FieldID int64 `json:"field_id,omitempty"`

	Collection string `json:"collection"`
	Field      string `json:"field"`

	// TODO: update and delete
	//CascadeDelete bool `json:"cascade_on_delete"`
}
