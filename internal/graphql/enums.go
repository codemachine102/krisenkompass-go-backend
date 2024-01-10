package graphql

import "github.com/graphql-go/graphql"

var statusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "status",
	Values: map[string]*graphql.EnumValueConfig{
		"hidden": {
			Value: "hidden",
		},
		"visible": {
			Value: "visible",
		},
		"deleted": {
			Value: "deleted",
		},
	},
	Description: "Status enum",
})

var organizationStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "organizationStatus",
	Values: map[string]*graphql.EnumValueConfig{
		"paid": {
			Value: "paid",
		},
		"not_paid": {
			Value: "not_paid",
		},
		"blocked": {
			Value: "blocked",
		},
	},
	Description: "Status enum",
})

var verificationStatusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "verificationStatus",
	Values: map[string]*graphql.EnumValueConfig{
		"verified": {
			Value: "verified",
		},
		"not_verified": {
			Value: "not_verified",
		},
	},
	Description: "Status enum",
})

var roleEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "role",
	Values: map[string]*graphql.EnumValueConfig{
		"owner": {
			Value: "owner",
		},
		"admin": {
			Value: "admin",
		},
		"editor": {
			Value: "editor",
		},
		"user": {
			Value: "user",
		},
	},
	Description: "Role enum",
})

var planEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "plan",
	Values: map[string]*graphql.EnumValueConfig{
		"basic": {
			Value: "basic",
		},
		"conference": {
			Value: "conference",
		},
		"school": {
			Value: "school",
		},
		"pro": {
			Value: "pro",
		},
	},
	Description: "Plan enum",
})

var pageThemeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "theme",
	Values: map[string]*graphql.EnumValueConfig{
		"deal_with": {
			Value: "deal_with",
		},
		"e_restore": {
			Value: "e_restore",
		},
		"precautions": {
			Value: "precautions",
		},
		"e_avoid": {
			Value: "e_avoid",
		},
		"e_gfs": {
			Value: "e_gfs",
		},
		"e_school": {
			Value: "e_school",
		},
	},
	Description: "Page theme enum",
})

var userTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "userType",
	Values: map[string]*graphql.EnumValueConfig{
		"superadmin": {
			Value: "superadmin",
		},
		"user": {
			Value: "user",
		},
		"demo": {
			Value: "demo",
		},
	},
	Description: "Page type enum",
})

var pageTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "pageType",
	Values: map[string]*graphql.EnumValueConfig{
		"section": {
			Value: "section",
		},
		"content": {
			Value: "content",
		},
		"file": {
			Value: "file",
		},
	},
	Description: "Page type enum",
})

var blockTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "blockType",
	Values: map[string]*graphql.EnumValueConfig{
		"default": {
			Value: "default",
		},
		"accordion": {
			Value: "accordion",
		},
		"link": {
			Value: "link",
		},
	},
	Description: "Block type enum",
})
