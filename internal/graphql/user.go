package graphql

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/graphql-go/graphql"
)

var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"firstname": &graphql.Field{
				Type: NullableString,
			},
			"lastname": &graphql.Field{
				Type: NullableString,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"image": &graphql.Field{
				Type: NullableString,
			},
			"type": &graphql.Field{
				Type: userTypeEnum,
			},
			"roles": &graphql.Field{
				Type: graphql.NewList(roleType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					rows, err := service.OrganizationUser.GetByUserID(p.Source.(domain.User).ID)
					if err != nil {
						return nil, err
					}
					return rows, err
				},
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"updatedAt": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	},
)
