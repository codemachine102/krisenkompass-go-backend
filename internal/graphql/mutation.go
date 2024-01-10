package graphql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/authorize"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/contains"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/password"
	"github.com/graphql-go/graphql"
	"gopkg.in/guregu/null.v4"
)

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"login": &graphql.Field{
			Type:        tokenType,
			Description: "Check user's email and password. Return JWT tokens if correct.",
			Args: graphql.FieldConfigArgument{
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				email := p.Args["email"].(string)
				password := p.Args["password"].(string)
				tokens, err := service.Auth.Login(email, password)
				return tokens, err
			},
		},
		"signup": &graphql.Field{
			Type:        userType,
			Description: "Create organization",
			Args: graphql.FieldConfigArgument{
				"firstname": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"lastname": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"name": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Name of the organization.",
				},
				"city": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"population": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"role": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "User's role in the organization.",
				},
				"website": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"phone": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"address": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"invoiceAddress": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"plan": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(planEnum),
				},
				"notes": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "This info will be sent to the admin. Can be used for price breakdown.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				firstname := p.Args["firstname"].(string)
				lastname := p.Args["lastname"].(string)
				email := p.Args["email"].(string)
				name := p.Args["name"].(string)
				city := p.Args["city"].(string)
				population := p.Args["population"].(int)
				role := p.Args["role"].(string)
				website := p.Args["website"].(string)
				phone := p.Args["phone"].(string)
				address := p.Args["address"].(string)
				invoiceAddress := p.Args["invoiceAddress"].(string)
				plan := p.Args["plan"].(string)
				notes := p.Args["notes"].(string)

				userID, organizationID, err := service.User.CreateWithOrganization(&firstname, &lastname, email, name, city, address, invoiceAddress, plan, population, role)
				if err != nil {
					fmt.Println(err)
					return nil, errors.New("failed to create user and organization")
				}

				user, err := service.User.GetByID(userID)
				if err != nil {
					fmt.Println(err)
					return nil, errors.New("user was created but failed to get")
				}

				go service.Email.SendAdminNewOrganization(organizationID, plan, user.Email, fmt.Sprintf("%s %s %s", user.Salutation.String, user.Firstname.String, user.Lastname.String), role, name, website, city, population, phone, address, invoiceAddress, notes)
				go service.Email.SendUserNewOrganization(organizationID, plan, user.Email, fmt.Sprintf("%s %s %s", user.Salutation.String, user.Firstname.String, user.Lastname.String), role, name, website, city, population, phone, address, invoiceAddress, notes)

				return user, err
			},
		},
		"verify": &graphql.Field{
			Type:        tokenType,
			Description: "Set up password for the user if verification tokes is correct.",
			Args: graphql.FieldConfigArgument{
				"token": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				token := p.Args["token"].(string)
				password := p.Args["password"].(string)
				tokens, err := service.Auth.Verify(token, password)
				return tokens, err
			},
		},
		"sendPasswordResetLink": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Send link to reset password on email.",
			Args: graphql.FieldConfigArgument{
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				email := p.Args["email"].(string)

				err := service.User.CreatePasswordReset(email)
				if err != nil {
					return false, errors.New("failed to send reset link")
				}

				return true, nil
			},
		},
		"updateUserPassword": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Update user password using password reset token or old password.",
			Args: graphql.FieldConfigArgument{
				"token": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"oldPassword": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var token, oldPassword null.String

				if temp, ok := p.Args["token"].(string); ok {
					token = null.StringFrom(temp)
				}
				if temp, ok := p.Args["oldPassword"].(string); ok {
					oldPassword = null.StringFrom(temp)

				}
				newPassword := p.Args["password"].(string)

				var userID int64
				if token.Valid {
					user, err := service.User.GetByPasswordResetToken(token.String)
					if err != nil {
						return false, errors.New("wrong password reset token")
					}
					userID = user.ID
				} else if oldPassword.Valid {
					claims, err := authorize.Authorize(p.Context, "user")
					if err != nil {
						return nil, err
					}

					userID = claims.UserID

					user, err := service.User.GetByID(userID)
					if err != nil {
						return false, errors.New("user not found")
					}

					err = password.CheckPassword(oldPassword.String, user.Password.String)
					if err != nil {
						return false, errors.New("wrong old password")
					}
				}

				err := service.User.UpdatePassword(userID, newPassword)
				if err != nil {
					log.Println("Failed to update password: ", err)
					return false, errors.New("failed to update password")
				}

				return true, err
			},
		},
		"refreshToken": &graphql.Field{
			Type:        tokenType,
			Description: "Get new pair of tokens using refreshToken.",
			Args: graphql.FieldConfigArgument{
				"refreshToken": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				refreshToken := p.Args["refreshToken"].(string)
				tokens, err := service.Auth.RefreshToken(refreshToken)
				return tokens, err
			},
		},
		"createOrganization": &graphql.Field{
			Type:        organizationType,
			Description: "Create organization",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Name of the organization.",
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"city": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"population": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"role": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "User's role in the organization.",
				},
				"website": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"phone": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"address": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"invoiceAddress": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"plan": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(planEnum),
				},
				"notes": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "This info will be sent to the admin. Can be used for price breakdown.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				claims, err := authorize.Authorize(p.Context, "user")
				if err != nil {
					return nil, err
				}
				name := p.Args["name"].(string)
				city := p.Args["city"].(string)
				population := p.Args["population"].(int)
				role := p.Args["role"].(string)
				website := p.Args["website"].(string)
				phone := p.Args["phone"].(string)
				address := p.Args["address"].(string)
				invoiceAddress := p.Args["invoiceAddress"].(string)
				plan := p.Args["plan"].(string)
				notes := p.Args["notes"].(string)
				var image *string
				imageTemp, ok := p.Args["image"].(string)
				if ok {
					image = &imageTemp
				}

				user, err := service.User.GetByID(claims.UserID)
				if err != nil {
					fmt.Println(err)
					return nil, errors.New("failed to create organization")
				}

				id, err := service.Organization.Create(image, name, city, address, invoiceAddress, plan, population, claims.UserID)
				if err != nil {
					fmt.Println(err)
					return nil, errors.New("failed to create organization")
				}

				organization, err := service.Organization.GetByID(id)
				if err != nil {
					return nil, errors.New("organization was created but failed to get")
				}

				go service.Email.SendAdminNewOrganization(organization.ID, organization.Plan, user.Email, fmt.Sprintf("%s %s %s", user.Salutation.String, user.Firstname.String, user.Lastname.String), role, name, website, city, population, phone, address, invoiceAddress, notes)
				go service.Email.SendUserNewOrganization(organization.ID, organization.Plan, user.Email, fmt.Sprintf("%s %s %s", user.Salutation.String, user.Firstname.String, user.Lastname.String), role, name, website, city, population, phone, address, invoiceAddress, notes)

				return organization, err
			},
		},
		"updateOrganization": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"city": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"address": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"invoiceAddress": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"plan": &graphql.ArgumentConfig{
					Type: planEnum,
				},
				"status": &graphql.ArgumentConfig{
					Type: organizationStatusEnum,
				},
				"population": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := int64(p.Args["id"].(int))
				var name, image, city, address, invoiceAddress, plan, status null.String
				var population null.Int
				if temp, ok := p.Args["image"].(string); ok {
					image = null.StringFrom(temp)
				}
				if temp, ok := p.Args["name"].(string); ok {
					name = null.StringFrom(temp)
				}
				if temp, ok := p.Args["city"].(string); ok {
					city = null.StringFrom(temp)
				}
				if temp, ok := p.Args["address"].(string); ok {
					address = null.StringFrom(temp)
				}
				if temp, ok := p.Args["invoiceAddress"].(string); ok {
					invoiceAddress = null.StringFrom(temp)
				}
				if temp, ok := p.Args["plan"].(string); ok {
					plan = null.StringFrom(temp)
				}
				if temp, ok := p.Args["status"].(string); ok {
					status = null.StringFrom(temp)
				}
				if temp, ok := p.Args["population"].(int); ok {
					population = null.IntFrom(int64(temp))
				}

				claims, err := authorize.AuthorizeOrganization(p.Context, id, "editor")
				if err != nil {
					return false, err
				}

				if claims.Type != "superadmin" {
					organization, err := service.Organization.GetByID(id)
					if err != nil {
						return nil, err
					}

					if organization.Status == "blocked" {
						return nil, errors.New("access denied")
					}
				}

				err = service.Organization.Update(id, image, name, city, address, invoiceAddress, plan, status, population)
				if err != nil {
					return false, errors.New("failed to update organization")
				}

				return true, nil
			},
		},
		"deleteOrganization": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := int64(p.Args["id"].(int))

				_, err := authorize.AuthorizeOrganization(p.Context, id, "owner")
				if err != nil {
					return false, err
				}

				err = service.Organization.Delete(id)
				if err != nil {
					return false, errors.New("failed to delete organization")
				}

				return true, nil
			},
		},
		"createUser": &graphql.Field{
			Type:        userType,
			Description: "Create a user and send a verification link on email.",
			Args: graphql.FieldConfigArgument{
				"firstname": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"lastname": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var firstname, lastname, image *string

				email := p.Args["email"].(string)

				firstnameTemp, ok := p.Args["firstname"].(string)
				if ok {
					firstname = &firstnameTemp
				}

				lastnameTemp, ok := p.Args["lastname"].(string)
				if ok {
					lastname = &lastnameTemp
				}

				imageTemp, ok := p.Args["image"].(string)
				if ok {
					image = &imageTemp
				}

				id, err := service.User.Create(firstname, lastname, email, image)
				if err != nil {
					return nil, err
				}

				user, err := service.User.GetByID(id)
				if err != nil {
					return nil, err
				}

				return user, nil
			},
		},
		"deleteUser": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := int64(p.Args["id"].(int))

				claims, err := authorize.Authorize(p.Context, "user")
				if err != nil || (claims.UserID != id && claims.Type != "superadmin") {
					return nil, errors.New("access denied")
				}

				err = service.User.Delete(id)
				if err != nil {
					return false, errors.New("failed to delete user")
				}

				return true, nil
			},
		},
		"createPage": &graphql.Field{
			Type:        pageType,
			Description: "Create page",
			Args: graphql.FieldConfigArgument{
				"organizationID": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"parentID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"type": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(pageTypeEnum),
				},
				"theme": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(pageThemeEnum),
				},
				"status": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(statusEnum),
				},
				"title": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"imageHover": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var image, imageHover *string
				var organizationID *int64
				tempOrganizationID, ok := p.Args["organizationID"].(int)
				if ok && tempOrganizationID != 0 {
					tempID := int64(tempOrganizationID)
					organizationID = &tempID
				}
				parentID := p.Args["parentID"].(int)
				pageType := p.Args["type"].(string)
				theme := p.Args["theme"].(string)
				status := p.Args["status"].(string)
				title := p.Args["title"].(string)
				tempImage, ok := p.Args["image"].(string)
				if ok && tempImage != "" {
					image = &tempImage
				}
				tempImageHover, ok := p.Args["imageHover"].(string)
				if ok && tempImageHover != "" {
					imageHover = &tempImageHover
				}

				if organizationID != nil {
					_, err := authorize.AuthorizeOrganization(p.Context, *organizationID, "editor")
					if err != nil {
						return nil, err
					}
				} else {
					_, err := authorize.Authorize(p.Context, "superadmin")
					if err != nil {
						return nil, err
					}
				}

				id, err := service.Page.Create(organizationID, int64(parentID), "de", pageType, theme, status, title, image, imageHover)
				if err != nil {
					return nil, err
				}

				page, err := service.Page.GetByID(id)
				return page, err
			},
		},
		"createBlock": &graphql.Field{
			Type:        blockType,
			Description: "Create block for a page",
			Args: graphql.FieldConfigArgument{
				"pageID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"title": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"content": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"readmore": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"imageHover": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"type": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(blockTypeEnum),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var content, readmore, image, imageHover *string
				pageID := p.Args["pageID"].(int)
				blockType := p.Args["type"].(string)
				title := p.Args["title"].(string)

				if temp, ok := p.Args["content"].(string); ok {
					content = &temp
				}
				if temp, ok := p.Args["readmore"].(string); ok {
					readmore = &temp
				}
				if temp, ok := p.Args["image"].(string); ok {
					image = &temp
				}
				if temp, ok := p.Args["imageHover"].(string); ok {
					imageHover = &temp
				}

				page, err := service.Page.GetByID(int64(pageID))
				if err != nil {
					return nil, errors.New("unable to find page with given id")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return nil, err
				}

				id, err := service.Block.Create(int64(pageID), title, blockType, content, readmore, image, imageHover)
				if err != nil {
					return nil, err
				}

				block, err := service.Block.GetByID(id)
				return block, err
			},
		},
		"createFileBlock": &graphql.Field{
			Type:        fileBlockType,
			Description: "Create file for a page",
			Args: graphql.FieldConfigArgument{
				"pageID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"path": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				pageID := p.Args["pageID"].(int)
				path := p.Args["path"].(string)

				page, err := service.Page.GetByID(int64(pageID))
				if err != nil {
					return nil, errors.New("unable to find page with given id")
				}
				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return nil, err
				}

				id, err := service.FileBlock.Create(int64(pageID), path)
				if err != nil {
					return nil, err
				}

				file, err := service.FileBlock.GetByID(id)
				return file, err
			},
		},
		"updatePage": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Update page",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"parentID": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"title": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"status": &graphql.ArgumentConfig{
					Type: statusEnum,
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"imageHover": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"plans": &graphql.ArgumentConfig{
					Type:        graphql.NewList(planEnum),
					Description: "Only for admins. Used to set default page plans",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var parentID *int64
				var title, status, image, imageHover *string
				id := p.Args["id"].(int)

				if temp, ok := p.Args["parentID"].(int); ok {
					temp := int64(temp)
					parentID = &temp
				}
				if temp, ok := p.Args["title"].(string); ok {
					title = &temp
				}
				if temp, ok := p.Args["status"].(string); ok {
					status = &temp
				}
				if temp, ok := p.Args["image"].(string); ok {
					image = &temp
				}
				if temp, ok := p.Args["imageHover"].(string); ok {
					imageHover = &temp
				}

				page, err := service.Page.GetByID(int64(id))
				if err != nil {
					return nil, errors.New("unable to find page with given id")
				}

				if page.OrganizationID.Valid {
					_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
					if err != nil {
						return false, ErrFailedToUpdate
					}
				} else {
					_, err := authorize.Authorize(p.Context, "superadmin")
					if err != nil {
						return false, ErrFailedToUpdate
					}

					plansTemp, ok := p.Args["plans"].([]interface{})
					if ok {
						plans := make([]string, len(plansTemp))
						for i := range plans {
							plans[i] = plansTemp[i].(string)
						}

						defaultPages, err := service.Page.GetDefaultPage(int64(id))
						if err != nil && err != sql.ErrNoRows {
							return false, ErrFailedToUpdate
						}

						var deletePlans = make([]string, 0)

						for _, p := range defaultPages {
							if !contains.ContainsStr(plans, p.Plan) {
								deletePlans = append(deletePlans, p.Plan)
							}
						}

						err = service.Page.DeleteDefaultPage(int64(id), deletePlans)
						if err != nil {
							return false, ErrFailedToUpdate
						}

						err = service.Page.CreateDefaultPage(int64(id), plans)
						if err != nil {
							return false, ErrFailedToUpdate
						}
					}
				}

				err = service.Page.Update(int64(id), parentID, status, title, image, imageHover)
				if err != nil {
					return false, ErrFailedToUpdate
				}

				return true, nil
			},
		},
		"updatePagesSort": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Update pages order",
			Args: graphql.FieldConfigArgument{
				"parentID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"sort": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				parentID := p.Args["parentID"].(int)
				sortTemp := p.Args["sort"].([]interface{})
				sort := make([]int, len(sortTemp))
				for i := range sort {
					sort[i] = sortTemp[i].(int)
				}

				page, err := service.Page.GetByID(int64(parentID))
				if err != nil {
					return nil, errors.New("unable to find page with given id")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return false, err
				}

				err = service.Page.UpdateSort(int64(parentID), sort)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"updateBlock": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Update block",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"title": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"content": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"readmore": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"imageHover": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var title, content, readmore, image, imageHover *string
				id := p.Args["id"].(int)

				if temp, ok := p.Args["title"].(string); ok {
					title = &temp
				}
				if temp, ok := p.Args["content"].(string); ok {
					content = &temp
				}
				if temp, ok := p.Args["readmore"].(string); ok {
					readmore = &temp
				}
				if temp, ok := p.Args["image"].(string); ok {
					image = &temp
				}
				if temp, ok := p.Args["imageHover"].(string); ok {
					imageHover = &temp
				}

				block, err := service.Block.GetByID(int64(id))
				if err != nil {
					return nil, errors.New("unable to find block with given id")
				}

				page, err := service.Page.GetByID(block.PageID)
				if err != nil {
					return nil, errors.New("unable to find page fora given block")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return false, err
				}

				err = service.Block.Update(int64(id), title, content, readmore, image, imageHover)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"updateBlocksSort": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Update pages order",
			Args: graphql.FieldConfigArgument{
				"pageID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"sort": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				pageID := p.Args["pageID"].(int)
				sortTemp := p.Args["sort"].([]interface{})
				sort := make([]int, len(sortTemp))
				for i := range sort {
					sort[i] = sortTemp[i].(int)
				}

				page, err := service.Page.GetByID(int64(pageID))
				if err != nil {
					return nil, errors.New("unable to find page with given id")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return false, err
				}

				err = service.Block.UpdateSort(int64(pageID), sort)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"updateFileBlock": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Update file block",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"path": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := p.Args["id"].(int)
				path := p.Args["path"].(string)

				file, err := service.FileBlock.GetByID(int64(id))
				if err != nil {
					return nil, errors.New("unable to find file with given id")
				}

				page, err := service.Page.GetByID(file.ID)
				if err != nil {
					return nil, errors.New("unable to find page for a given file")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return false, err
				}

				err = service.FileBlock.Update(int64(id), path)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"deletePage": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Delete page and all related data",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				pageID := p.Args["id"].(int)

				page, err := service.Page.GetByID(int64(pageID))
				if err != nil {
					return false, err
				}

				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return false, err
				}

				err = service.Page.Delete(int64(pageID))
				if err != nil {
					return false, err
				}

				return true, err
			},
		},
		"deleteBlock": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Delete page and all related data",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				blockID := p.Args["id"].(int)

				block, err := service.Block.GetByID(int64(blockID))
				if err != nil {
					return false, errors.New("unable to find block with given id")
				}

				page, err := service.Page.GetByID(block.PageID)
				if err != nil {
					return false, err
				}

				_, err = authorize.AuthorizeOrganization(p.Context, page.OrganizationID.Int64, "editor")
				if err != nil {
					return false, err
				}

				err = service.Block.Delete(block.ID)
				if err != nil {
					return false, err
				}

				return true, err
			},
		},
		"deleteAddress": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Delete address by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				addressID := int64(p.Args["id"].(int))

				block, err := service.Address.GetByID(addressID)
				if err != nil {
					return false, errors.New("unable to find address with given id")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, block.OrganizationID, "editor")
				if err != nil {
					return false, err
				}

				err = service.Address.Delete(addressID)
				if err != nil {
					return false, err
				}

				return true, err
			},
		},
		"deleteFile": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Delete file or folder",
			Args: graphql.FieldConfigArgument{
				"path": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Path to file or folder. Name of the file should be included",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				path := p.Args["path"].(string)
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

					_, err = authorize.AuthorizeOrganization(p.Context, id, "editor")
					if err != nil {
						return nil, err
					}
				} else if pathParts[0] == "user" {
					id, err := strconv.ParseInt(pathParts[1], 10, 64)
					if err != nil {
						return nil, errors.New("invalid user id")
					}

					claims, err := authorize.Authorize(p.Context, "user")
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

				err := service.File.DeleteFiles(path)
				if err != nil {
					return false, err
				}

				return true, err
			},
		},
		"createFolder": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Create folder in the filesystem",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Name of the folder",
				},
				"path": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Path to folder where new folder should be created",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				name := p.Args["name"].(string)
				path := p.Args["path"].(string)
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

					_, err = authorize.AuthorizeOrganization(p.Context, id, "editor")
					if err != nil {
						return nil, err
					}
				} else if pathParts[0] == "user" {
					id, err := strconv.ParseInt(pathParts[1], 10, 64)
					if err != nil {
						return nil, errors.New("invalid user id")
					}

					claims, err := authorize.Authorize(p.Context, "user")
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

				err := service.File.CreateFolder(name, path)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"renameFile": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Rename file or folder",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Name of the folder",
				},
				"path": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "Path to folder where new folder should be created",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				name := p.Args["name"].(string)
				path := p.Args["path"].(string)
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

					_, err = authorize.AuthorizeOrganization(p.Context, id, "editor")
					if err != nil {
						return nil, err
					}
				} else if pathParts[0] == "user" {
					id, err := strconv.ParseInt(pathParts[1], 10, 64)
					if err != nil {
						return nil, errors.New("invalid user id")
					}

					claims, err := authorize.Authorize(p.Context, "user")
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

				pathParts = pathParts[:len(pathParts)-1]
				newPath := "/" + strings.Join(pathParts, "/") + "/" + name

				err := service.File.RenameFile(path, newPath)
				if err != nil {
					return false, errors.New("failed to rename")
				}

				return true, nil
			},
		},
		"createAddress": &graphql.Field{
			Type:        addressType,
			Description: "Create address record for organization",
			Args: graphql.FieldConfigArgument{
				"organizationID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"firstname": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"lastname": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"phone": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"phoneExtra": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"role": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"info": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				organizationID := int64(p.Args["organizationID"].(int))
				firstname := p.Args["firstname"].(string)
				lastname := p.Args["lastname"].(string)
				email := p.Args["email"].(string)
				phone := p.Args["phone"].(string)
				var phoneExtra, role, info null.String
				if temp, ok := p.Args["phoneExtra"].(string); ok {
					phoneExtra = null.StringFrom(temp)
				}
				if temp, ok := p.Args["role"].(string); ok {
					role = null.StringFrom(temp)
				}
				if temp, ok := p.Args["info"].(string); ok {
					info = null.StringFrom(temp)
				}

				_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "editor")
				if err != nil {
					return nil, err
				}

				id, err := service.Address.Create(organizationID, firstname, lastname, email, phone, phoneExtra, role, info)
				if err != nil {
					return false, errors.New("failed to create address")
				}

				address, err := service.Address.GetByID(id)
				if err != nil {
					return false, errors.New("failed to get created address")
				}

				return address, nil
			},
		},
		"updateAddress": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Create address record for organization",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"firstname": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"lastname": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"phone": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"phoneExtra": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"role": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"info": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := int64(p.Args["id"].(int))
				var firstname, lastname, email, phone, phoneExtra, role, info null.String

				if temp, ok := p.Args["firstname"].(string); ok {
					firstname = null.StringFrom(temp)
				}
				if temp, ok := p.Args["lastname"].(string); ok {
					lastname = null.StringFrom(temp)
				}
				if temp, ok := p.Args["email"].(string); ok {
					email = null.StringFrom(temp)
				}
				if temp, ok := p.Args["phone"].(string); ok {
					phone = null.StringFrom(temp)
				}
				if temp, ok := p.Args["phoneExtra"].(string); ok {
					phoneExtra = null.StringFrom(temp)
				}
				if temp, ok := p.Args["role"].(string); ok {
					role = null.StringFrom(temp)
				}
				if temp, ok := p.Args["info"].(string); ok {
					info = null.StringFrom(temp)
				}

				address, err := service.Address.GetByID(id)
				if err != nil {
					return false, errors.New("address with given id not found")
				}

				_, err = authorize.AuthorizeOrganization(p.Context, address.OrganizationID, "editor")
				if err != nil {
					return false, err
				}

				err = service.Address.Update(id, firstname, lastname, email, phone, phoneExtra, role, info)
				if err != nil {
					return false, errors.New("failed to update address")
				}

				return true, nil
			},
		},
		"updateAddressesSort": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Update pages order",
			Args: graphql.FieldConfigArgument{
				"organizationID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"sort": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				organizationID := int64(p.Args["organizationID"].(int))
				sortTemp := p.Args["sort"].([]interface{})
				sort := make([]int64, len(sortTemp))
				for i := range sort {
					sort[i] = int64(sortTemp[i].(int))
				}

				_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "editor")
				if err != nil {
					return false, err
				}

				err = service.Address.UpdateSort(organizationID, sort)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"updateUser": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Create address record for organization",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"image": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"firstname": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"lastname": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"type": &graphql.ArgumentConfig{
					Type: userTypeEnum,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := int64(p.Args["id"].(int))
				var image, firstname, lastname, email, userType *string

				if temp, ok := p.Args["image"].(string); ok {
					image = &temp
				}
				if temp, ok := p.Args["firstname"].(string); ok {
					firstname = &temp
				}
				if temp, ok := p.Args["lastname"].(string); ok {
					lastname = &temp
				}
				if temp, ok := p.Args["email"].(string); ok {
					email = &temp
				}
				if temp, ok := p.Args["type"].(string); ok {
					userType = &temp
				}

				claims, err := authorize.Authorize(p.Context, "user")
				if err != nil || (claims.UserID != id && claims.Type != "superadmin") {
					return false, errors.New("access denied")
				}

				if (email != nil || userType != nil) && claims.Type != "superadmin" {
					return false, errors.New("access denied")
				}

				err = service.User.Update(id, image, firstname, lastname, email, userType)
				if err != nil {
					return false, errors.New("failed to update user")
				}

				return true, nil
			},
		},
		"copyDefaultContent": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Copy default content to organization.",
			Args: graphql.FieldConfigArgument{
				"organizationID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				_, err := authorize.Authorize(p.Context, "superadmin")
				if err != nil {
					return nil, err
				}

				organizationID := int64(p.Args["organizationID"].(int))
				organization, err := service.Organization.GetByID(organizationID)
				if err != nil {
					return false, errors.New("organization not found")
				}

				err = service.Organization.CopyDefaultContent(organizationID, organization.Plan)
				if err != nil {
					return false, errors.New("failed to copy content")
				}

				return true, nil
			},
		},
		"inviteUser": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Invite user to the organization.",
			Args: graphql.FieldConfigArgument{
				"organizationID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"email": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"role": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(roleEnum),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				organizationID := int64(p.Args["organizationID"].(int))
				email := p.Args["email"].(string)
				role := p.Args["role"].(string)

				_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "admin")
				if err != nil {
					return false, err
				}

				organization, err := service.Organization.GetByID(organizationID)
				if err != nil {
					return nil, err
				}

				if organization.Status == "blocked" {
					return nil, errors.New("access denied")
				}

				err = service.User.CreateOrganizationUser(organizationID, email, role)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"deleteOrganizationUser": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Invite user to the organization.",
			Args: graphql.FieldConfigArgument{
				"organizationID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"userID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				organizationID := int64(p.Args["organizationID"].(int))
				userID := int64(p.Args["userID"].(int))

				_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "admin")
				if err != nil {
					return false, err
				}

				organization, err := service.Organization.GetByID(organizationID)
				if err != nil {
					return nil, err
				}

				if organization.Status == "blocked" {
					return nil, errors.New("access denied")
				}

				err = service.User.DeleteOrganizationUser(organizationID, userID)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
		"updateOrganizationUser": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "Invite user to the organization.",
			Args: graphql.FieldConfigArgument{
				"organizationID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"userID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"role": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(roleEnum),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				organizationID := int64(p.Args["organizationID"].(int))
				userID := int64(p.Args["userID"].(int))
				role := p.Args["role"].(string)

				_, err := authorize.AuthorizeOrganization(p.Context, organizationID, "admin")
				if err != nil {
					return false, err
				}

				organization, err := service.Organization.GetByID(organizationID)
				if err != nil {
					return nil, err
				}

				if organization.Status == "blocked" {
					return nil, errors.New("access denied")
				}

				err = service.User.UpdateOrganizationUser(organizationID, userID, role)
				if err != nil {
					return false, err
				}

				return true, nil
			},
		},
	},
})
