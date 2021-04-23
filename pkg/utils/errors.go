package utils

type ErrorExplained interface {
	Error() string
	ExplainError() string
}
