package handlers

import (
	"database/sql"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// ContentHandlerConfig holds configuration for creating a ContentHandler
type ContentHandlerConfig struct {
	DB                 bob.Executor
	RawDB              *sql.DB
	IBESystem          *ibe.IBESystem
	IdentityMappingDAO dao.IdentityMappingDAOInterface
	UserDAO            dao.UserDAOInterface
	PostDAO            dao.PostDAOInterface
	CommentDAO         dao.CommentDAOInterface
	SubforumDAO        dao.SubforumDAOInterface
	PseudonymDAO       dao.PseudonymDAOInterface
	VoteDAO            dao.VoteDAOInterface
	UserBlocksDAO      dao.UserBlocksDAOInterface
	RoleKeyDAO         dao.RoleKeyDAOInterface
	PermissionChecker  middleware.PermissionCheckerInterface
	PermissionDAO      dao.PermissionDAOInterface
	ReportDAO          dao.ReportDAOInterface
}

// NewContentHandlerConfig creates a new configuration for ContentHandler
func NewContentHandlerConfig(db bob.Executor, rawDB *sql.DB, ibeSystem *ibe.IBESystem) *ContentHandlerConfig {
	return &ContentHandlerConfig{
		DB:        db,
		RawDB:     rawDB,
		IBESystem: ibeSystem,
	}
}

// ContentHandler handles content-related requests
type ContentHandler struct {
	db                 bob.Executor
	rawDB              *sql.DB
	ibeSystem          *ibe.IBESystem
	identityMappingDAO dao.IdentityMappingDAOInterface
	userDAO            dao.UserDAOInterface
	postDAO            dao.PostDAOInterface
	commentDAO         dao.CommentDAOInterface
	subforumDAO        dao.SubforumDAOInterface
	pseudonymDAO       dao.PseudonymDAOInterface
	voteDAO            dao.VoteDAOInterface
	permissionChecker  middleware.PermissionCheckerInterface
	permissionDAO      dao.PermissionDAOInterface
	reportDAO          dao.ReportDAOInterface
}

// NewContentHandler creates a new content handler with optional dependencies
// If db is provided, real DAOs will be created. If nil, mock DAOs should be provided.
func NewContentHandler(
	db bob.Executor,
	rawDB *sql.DB,
	ibeSystem *ibe.IBESystem,
	identityMappingDAO dao.IdentityMappingDAOInterface,
	userDAO dao.UserDAOInterface,
	postDAO dao.PostDAOInterface,
	commentDAO dao.CommentDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	voteDAO dao.VoteDAOInterface,
	userBlocksDAO dao.UserBlocksDAOInterface,
	roleKeyDAO dao.RoleKeyDAOInterface,
	permissionChecker middleware.PermissionCheckerInterface,
	permissionDAO dao.PermissionDAOInterface,
	reportDAO dao.ReportDAOInterface,
) *ContentHandler {
	// If db is provided, create real DAOs (production mode)
	if db != nil {
		roleKeyDAO = dao.NewRoleKeyDAO(db, nil)
		userBlocksDAO = dao.NewUserBlocksDAO(db)

		// Safe type assertions with error handling
		identityMappingDAOImpl, ok := identityMappingDAO.(*dao.IdentityMappingDAO)
		if !ok {
			log.Error().Msg("identityMappingDAO is not of type *dao.IdentityMappingDAO")
			return nil
		}
		userDAOImpl, ok := userDAO.(*dao.UserDAO)
		if !ok {
			log.Error().Msg("userDAO is not of type *dao.UserDAO")
			return nil
		}
		roleKeyDAOImpl, ok := roleKeyDAO.(*dao.RoleKeyDAO)
		if !ok {
			log.Error().Msg("roleKeyDAO is not of type *dao.RoleKeyDAO")
			return nil
		}
		userBlocksDAOImpl, ok := userBlocksDAO.(*dao.UserBlocksDAO)
		if !ok {
			log.Error().Msg("userBlocksDAO is not of type *dao.UserBlocksDAO")
			return nil
		}

		pseudonymDAO = dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAOImpl, userDAOImpl, roleKeyDAOImpl, userBlocksDAOImpl)
		permissionDAO = dao.NewPermissionDAO(db)
		postDAO = dao.NewPostDAO(db)
		commentDAO = dao.NewCommentDAO(db)
		subforumDAO = dao.NewSubforumDAO(db)
		voteDAO = dao.NewVoteDAO(db)
		permissionChecker = middleware.NewPermissionChecker(db)
		reportDAO = dao.NewReportDAO(db)
	}

	return &ContentHandler{
		db:                 db,
		rawDB:              rawDB,
		ibeSystem:          ibeSystem,
		identityMappingDAO: identityMappingDAO,
		userDAO:            userDAO,
		postDAO:            postDAO,
		commentDAO:         commentDAO,
		subforumDAO:        subforumDAO,
		pseudonymDAO:       pseudonymDAO,
		voteDAO:            voteDAO,
		permissionChecker:  permissionChecker,
		permissionDAO:      permissionDAO,
		reportDAO:          reportDAO,
	}
}

// NewContentHandlerFromConfig creates a new content handler from a configuration struct
func NewContentHandlerFromConfig(cfg *ContentHandlerConfig) *ContentHandler {
	return NewContentHandler(
		cfg.DB,
		cfg.RawDB,
		cfg.IBESystem,
		cfg.IdentityMappingDAO,
		cfg.UserDAO,
		cfg.PostDAO,
		cfg.CommentDAO,
		cfg.SubforumDAO,
		cfg.PseudonymDAO,
		cfg.VoteDAO,
		cfg.UserBlocksDAO,
		cfg.RoleKeyDAO,
		cfg.PermissionChecker,
		cfg.PermissionDAO,
		cfg.ReportDAO,
	)
}
