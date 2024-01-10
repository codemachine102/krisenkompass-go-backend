package authorize

import (
	"context"
	"errors"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
)

// Checks if user is one of specified types. If user types are not specified then user
// will be tested for having "superadmin" type.
func Authorize(c context.Context, types ...string) (*domain.Claims, error) {
	claims, ok := c.Value(domain.UserClaimsKey).(*domain.Claims)
	if !ok {
		return claims, errors.New("access denied")
	}

	if claims.Type == "superadmin" {
		return claims, nil
	}

	for _, typ := range types {
		if claims.Type == typ {
			return claims, nil
		}
	}

	return claims, errors.New("access denied")
}

func AuthorizeOrganization(c context.Context, organizationID int64, role string) (*domain.Claims, error) {
	claims, ok := c.Value(domain.UserClaimsKey).(*domain.Claims)
	if !ok {
		return nil, errors.New("access denied")
	}

	if claims.Roles[organizationID] == "owner" {
		return claims, nil
	} else if claims.Roles[organizationID] == "admin" && role != "owner" {
		return claims, nil
	} else if claims.Roles[organizationID] == "editor" && role != "owner" && role != "admin" {
		return claims, nil
	} else if claims.Roles[organizationID] == "user" && role == "user" {
		return claims, nil
	} else if claims.Type == "superadmin" {
		return claims, nil
	}

	return nil, errors.New("access denied")
}
