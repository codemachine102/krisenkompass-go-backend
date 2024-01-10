package graphql

import "github.com/graphql-go/graphql"

var addressType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Address",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"organizationID": &graphql.Field{
				Type: graphql.Int,
			},
			"firstname": &graphql.Field{
				Type: graphql.String,
			},
			"lastname": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"phone": &graphql.Field{
				Type: graphql.String,
			},
			"phoneExtra": &graphql.Field{
				Type: NullableString,
			},
			"role": &graphql.Field{
				Type: NullableString,
			},
			"info": &graphql.Field{
				Type: NullableString,
			},
			"sort": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)
