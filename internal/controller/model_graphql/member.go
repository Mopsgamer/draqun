package model_graphql

import (
	"github.com/graphql-go/graphql"
)

var Member = graphql.NewObject(graphql.ObjectConfig{
	Name: "Member",
	Fields: graphql.Fields{
		"id":   &graphql.Field{Type: graphql.String},
		"name": &graphql.Field{Type: graphql.String},
	},
})
