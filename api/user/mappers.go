package user

import "github.com/slipe-fun/bloom-kit/domain"

func NewEditUserRequest(username, displayName, description *string) *domain.EditUserRequest {
	return &domain.EditUserRequest{
		Username:    username,
		DisplayName: displayName,
		Description: description,
	}
}
