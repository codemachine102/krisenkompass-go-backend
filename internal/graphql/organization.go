package graphql

import (
	"errors"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/authorize"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/graphql-go/graphql"
)

var organizationType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Organization",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"image": &graphql.Field{
				Type: NullableString,
			},
			"city": &graphql.Field{
				Type: graphql.String,
			},
			"street": &graphql.Field{
				Type: graphql.String,
			},
			"population": &graphql.Field{
				Type: graphql.Int,
			},
			"address": &graphql.Field{
				Type: graphql.String,
			},
			"invoiceAddress": &graphql.Field{
				Type: graphql.String,
			},
			"plan": &graphql.Field{
				Type: planEnum,
			},
			"status": &graphql.Field{
				Type: organizationStatusEnum,
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"updatedAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"users": &graphql.Field{
				Type: graphql.NewList(userType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					organizationID := p.Source.(domain.Organization).ID

					_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "user")
					if err != nil {
						return nil, err
					}

					organizations, err := service.User.GetByOrganizationID(organizationID)

					return organizations, err
				},
			},
			"pages": &graphql.Field{
				Type: graphql.NewList(pageType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					organizationID := p.Source.(domain.Organization).ID
					status := p.Source.(domain.Organization).Status
					if status == "blocked" {
						return nil, errors.New("access denied")
					}

					_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "user")
					if err != nil {
						return nil, err
					}

					pages, err := service.Page.GetPages(&organizationID)

					return pages, err
				},
			},
			"addresses": &graphql.Field{
				Type: graphql.NewList(addressType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					organizationID := p.Source.(domain.Organization).ID
					status := p.Source.(domain.Organization).Status
					if status == "blocked" {
						return nil, errors.New("access denied")
					}

					_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "user")
					if err != nil {
						return nil, err
					}

					addresses, err := service.Address.GetByOrganizationID(organizationID)

					return addresses, err
				},
			},
		},
	},
)
