//go:build generate

package ibe

//go:generate mockgen -source=interfaces.go -destination=gomocks.go -package=ibe
