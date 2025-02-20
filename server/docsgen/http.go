package docsgen

import "reflect"

type Docs struct {
	HTTP map[string][]DocsHTTPMethod
}

type DocsHTTPMethod struct {
	Path        string
	Method      string
	Description string
	Request     []reflect.StructField
	Response    string
}

func New() *Docs {
	return &Docs{
		HTTP: map[string][]DocsHTTPMethod{},
	}
}

func FieldsOf(o any) []reflect.StructField {
	return reflect.VisibleFields(reflect.TypeOf(o))
}
