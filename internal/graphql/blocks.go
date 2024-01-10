package graphql

import "github.com/graphql-go/graphql"

var fileBlockType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "FileBlock",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"path": &graphql.Field{
				Type: graphql.String,
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"page": &graphql.Field{
				Type: pageType,
			},
		},
	},
)

var blockType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Block",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"pageID": &graphql.Field{
				Type: graphql.Int,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"content": &graphql.Field{
				Type: NullableString,
			},
			"readmore": &graphql.Field{
				Type: NullableString,
			},
			"image": &graphql.Field{
				Type: NullableString,
			},
			"imageHover": &graphql.Field{
				Type: NullableString,
			},
			"type": &graphql.Field{
				Type: blockTypeEnum,
			},
			"page": &graphql.Field{
				Type: pageType,
			},
		},
	},
)
