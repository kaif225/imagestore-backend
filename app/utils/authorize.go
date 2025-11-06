package utils

import "errors"

type ContextKey string

func Authroizeuser(userRole string, allowedRoles ...string) (bool, error) {

	for _, allowedRole := range allowedRoles {
		if allowedRole == userRole {
			return true, nil
		}
	}

	return false, errors.New("User is not authorized")
}
