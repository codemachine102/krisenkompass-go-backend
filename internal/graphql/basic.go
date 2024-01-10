package graphql

import "github.com/graphql-go/graphql"

var tokenType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Token",
		Fields: graphql.Fields{
			"accessToken": &graphql.Field{
				Type: graphql.String,
			},
			"refreshToken": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var roleType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Role",
		Fields: graphql.Fields{
			"organizationID": &graphql.Field{
				Type: graphql.Int,
			},
			"role": &graphql.Field{
				Type: roleEnum,
			},
		},
	},
)

var fileType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "File",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"path": &graphql.Field{
				Type: graphql.String,
			},
			"isDir": &graphql.Field{
				Type: graphql.Boolean,
			},
			"updatedAt": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	},
)

var userVerificationType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "UserVerification",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"token": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: verificationStatusEnum,
			},
		},
	},
)
