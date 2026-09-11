package dto

type UserRetrieveRequest struct {
	HashString string `json:"hash_string" validate:"required"`
}
