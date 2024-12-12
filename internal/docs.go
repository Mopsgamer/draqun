package internal

import "reflect"

type Docs struct {
	HTTP map[string][]DocsHTTPMethod
}

type DocsHTTPMethod struct {
	Path        string
	Method      string
	Description string
	Request     []reflect.StructField
}

func initDocs() *Docs {
	return &Docs{
		HTTP: map[string][]DocsHTTPMethod{},
	}
}

func fieldsOf(o any) []reflect.StructField {
	return reflect.VisibleFields(reflect.TypeOf(o))
}
