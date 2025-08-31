//go:build generate

package dao

//go:generate mockgen -source=interfaces.go -destination=daos.go -package=dao
