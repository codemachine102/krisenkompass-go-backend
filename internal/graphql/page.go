package graphql

import (
	"database/sql"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/graphql-go/graphql"
)

var pageType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Page",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"organizationID": &graphql.Field{
				Type: NullableInt,
			},
			"parentID": &graphql.Field{
				Type: NullableInt,
			},
			"languageTag": &graphql.Field{
				Type: graphql.String,
			},
			"type": &graphql.Field{
				Type: pageTypeEnum,
			},
			"theme": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: statusEnum,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"image": &graphql.Field{
				Type: NullableString,
			},
			"imageHover": &graphql.Field{
				Type: NullableString,
			},
			"sort": &graphql.Field{
				Type: graphql.Int,
			},
			"plans": &graphql.Field{
				Type: graphql.NewList(planEnum),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if p.Source.(domain.Page).OrganizationID.Valid {
						return []string{}, nil
					}

					defaultPages, err := service.Page.GetDefaultPage(p.Source.(domain.Page).ID)
					if err != nil {
						if err == sql.ErrNoRows {
							return []string{}, nil
						}
						return nil, err
					}

					plans := make([]string, 0)
					for _, p := range defaultPages {
						plans = append(plans, p.Plan)
					}

					return plans, err
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
