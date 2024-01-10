package graphql

import (
	"database/sql"
	"log"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/authorize"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	s "bitbucket.org/ibros_nsk/krisenkompass-backend/internal/service"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

var service *s.Service

func supplementTypes() {
	userType.AddFieldConfig("organizations", &graphql.Field{
		Type: graphql.NewList(organizationType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			claims, err := authorize.Authorize(p.Context, "user", "demo")
			if err != nil {
				return nil, err
			}

			organizations, err := service.Organization.GetByUserID(claims.UserID)

			return organizations, err
		},
	})

	organizationType.AddFieldConfig("rootPages", &graphql.Field{
		Type: graphql.NewList(pageType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			organizationID := p.Source.(domain.Organization).ID
			_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "user")
			if err != nil {
				return nil, err
			}

			pages, err := service.Page.GetRootPages(&organizationID)

			return pages, err
		},
	})

	pageType.AddFieldConfig("parent", &graphql.Field{
		Type: pageType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return nil, nil
		},
	})

	pageType.AddFieldConfig("childrens", &graphql.Field{
		Type:        graphql.NewList(pageType),
		Description: "Get child pages",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			pages, err := service.Page.GetChildrens(p.Source.(domain.Page).ID)
			if err != nil {
				return nil, err
			}
			return pages, err
		},
	})

	pageType.AddFieldConfig("blocks", &graphql.Field{
		Type:        graphql.NewList(blockType),
		Description: "Get blocks for a page",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			blocks, err := service.Block.GetByPageID(p.Source.(domain.Page).ID)
			if err != nil {
				return nil, err
			}
			return blocks, err
		},
	})

	pageType.AddFieldConfig("file", &graphql.Field{
		Type:        fileBlockType,
		Description: "Get file for a page",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			file, err := service.FileBlock.GetByPageID(p.Source.(domain.Page).ID)
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}

			if err == sql.ErrNoRows {
				return nil, nil
			}

			return file, nil
		},
	})
}

func schema() *graphql.Schema {
	supplementTypes()

	schemaConfig := graphql.SchemaConfig{Query: queryType, Mutation: mutationType}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("Failed to create GraphQL Schema, err %v", err)
	}

	return &schema
}

func Handler(s *s.Service) *handler.Handler {
	service = s

	return handler.New(&handler.Config{
		Schema:     schema(),
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})
}
