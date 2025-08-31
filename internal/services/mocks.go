//go:build generate

package services

//go:generate mockgen -source=interfaces.go -destination=gomocks.go -package=services
