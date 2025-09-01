//go:build generate

package dao

//go:generate mockgen -source=interfaces.go -destination=dao_gomocks.go -package=dao
