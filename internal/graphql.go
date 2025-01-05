package internal

import (
	"restapp/internal/controller/database"
	"restapp/internal/controller/model_graphql"
	"strconv"

	"github.com/graphql-go/graphql"
)

var graphqlTypes = []*graphql.Object{
	model_graphql.Member,
}

var graphqlFields = func(db database.Database) graphql.Fields {
	return graphql.Fields{
		// FIXME: This is unsafe. Should be authorized, be a member and have rights.
		"members": &graphql.Field{
			Type: graphql.NewList(model_graphql.Member),
			Args: graphql.FieldConfigArgument{
				"groupId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (any, error) {
				groupIdString := p.Args["groupId"].(string)
				groupId, err := strconv.ParseUint(groupIdString, 10, 1)
				if err != nil {
					return nil, err
				}

				members := db.MemberList(groupId)
				return members, nil
			},
		},
	}
}

type GraphqlInput struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

func initGraphql(db database.Database) (graphql.Schema, graphql.Fields, error) {
	fields := graphqlFields(db)
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	return schema, fields, err
}
