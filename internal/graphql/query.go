package graphql

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/authorize"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/graphql-go/graphql"
)

var queryType = graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: graphql.Fields{
	"organization": &graphql.Field{
		Type:        organizationType,
		Description: "Get organization by id",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			organization, err := service.Organization.GetByID(int64(p.Args["id"].(int)))
			if err != nil {
				return nil, err
			}

			return organization, err
		},
	},
	"organizations": &graphql.Field{
		Type:        graphql.NewList(organizationType),
		Description: "Get all organizations",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			_, err := authorize.Authorize(p.Context, "superadmin")
			if err != nil {
				return nil, err
			}

			organizations, err := service.Organization.GetAll()
			if err != nil {
				return nil, err
			}

			return organizations, nil
		},
	},
	"user": &graphql.Field{
		Type:        userType,
		Description: "Get user by id",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			_, err := authorize.Authorize(p.Context, "user", "demo")
			if err != nil {
				return nil, err
			}

			user, err := service.User.GetByID(int64(p.Args["id"].(int)))
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, errors.New("user not found")
				}
				return nil, err
			}

			return user, nil
		},
	},
	"users": &graphql.Field{
		Type:        graphql.NewList(userType),
		Description: "Get all users",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			_, err := authorize.Authorize(p.Context, "superadmin")
			if err != nil {
				return nil, err
			}

			users, err := service.User.GetAll()
			if err != nil {
				return nil, err
			}

			return users, nil
		},
	},
	"userVerification": &graphql.Field{
		Type:        userVerificationType,
		Description: "Get user verification",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			user, err := service.User.GetVerification(int64(p.Args["id"].(int)))
			if err != nil {
				return nil, err
			}

			return user, nil
		},
	},
	"page": &graphql.Field{
		Type:        pageType,
		Description: "Get page by id",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			page, err := service.Page.GetByID(int64(p.Args["id"].(int)))
			if err != nil {
				return nil, err
			}

			if page.OrganizationID.Valid {
				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "user")
				if err != nil {
					return nil, err
				}

				organization, err := service.Organization.GetByID(page.OrganizationID.Int64)
				if err != nil {
					return nil, err
				}

				if organization.Status == "blocked" {
					return nil, errors.New("access denied")
				}
			} else {
				_, err := authorize.Authorize(p.Context, "superadmin")
				if err != nil {
					return nil, err
				}
			}

			return page, err
		},
	},
	"pages": &graphql.Field{
		Type:        graphql.NewList(pageType),
		Description: "Get list of all pages for organization.",
		Args: graphql.FieldConfigArgument{
			"organizationID": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var organizationID *int64
			tempOrganizationID, ok := p.Args["organizationID"].(int)
			if ok && tempOrganizationID != 0 {
				tempID := int64(tempOrganizationID)
				organizationID = &tempID
			}

			if organizationID != nil {
				_, err := authorize.AuthorizeOrganization(p.Context, *organizationID, "user")
				if err != nil {
					return nil, err
				}
				
				organization, err := service.Organization.GetByID(*organizationID)
				if err != nil {
					return nil, err
				}

				if organization.Status == "blocked" {
					return nil, errors.New("access denied")
				}

			} else {
				_, err := authorize.Authorize(p.Context, "superadmin")
				if err != nil {
					return nil, err
				}
			}

			pages, err := service.Page.GetPages(organizationID)

			return pages, err
		},
	},
	"fileBlock": &graphql.Field{
		Type:        fileBlockType,
		Description: "Get file block by id",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			file, err := service.FileBlock.GetByID(int64(p.Args["id"].(int)))
			if err != nil {
				return nil, err
			}
			return file, err
		},
	},
	"files": &graphql.Field{
		Type:        graphql.NewList(fileType),
		Description: "Get list of files for specified path and organization",
		Args: graphql.FieldConfigArgument{
			"path": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			path, _ := p.Args["path"].(string)
			path = strings.Trim(path, "/")
			pathParts := strings.Split(path, "/")
			if len(path) != 0 {
				path = "/" + path
			}

			// Forbid navigation to parent.
			if strings.Contains(path, "../") {
				return nil, errors.New("access denied")
			}

			// Check if user has rights to access requested path.
			if pathParts[0] == "organization" {
				id, err := strconv.ParseInt(pathParts[1], 10, 64)
				if err != nil {
					return nil, errors.New("invalid organization id")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, id, "user")
				if err != nil {
					return nil, err
				}
			} else if pathParts[0] == "user" {
				id, err := strconv.ParseInt(pathParts[1], 10, 64)
				if err != nil {
					return nil, errors.New("invalid user id")
				}

				claims, err := authorize.Authorize(p.Context, "user", "demo")
				if err != nil {
					return nil, err
				}

				if claims.UserID != id {
					return nil, errors.New("access denied")
				}
			} else if pathParts[0] == "common" {
				_, err := authorize.Authorize(p.Context, "superadmin")
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("wrong path")
			}

			fileList, err := service.File.GetFiles(path)
			if err != nil {
				return nil, err
			}

			var files []domain.File
			for _, f := range fileList {
				files = append(files, domain.File{
					Name:      f.Name(),
					Path:      fmt.Sprintf("%s/%s", path, f.Name()),
					IsDir:     f.IsDir(),
					UpdatedAt: f.ModTime(),
				})
			}

			return files, nil
		},
	},
	"adminPages": &graphql.Field{
		Type:        graphql.NewList(pageType),
		Description: "Get admin root pages",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			_, err := authorize.Authorize(p.Context, "superadmin")
			if err != nil {
				return nil, err
			}

			pages, err := service.Page.GetRootPages(nil)

			return pages, err
		},
	},
}})
