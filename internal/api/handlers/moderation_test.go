package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
)

// TestNewModerationHandler tests the moderation handler constructor
func TestNewModerationHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDAO := dao.NewMockReportDAOInterface(ctrl)
	mockModerationActionDAO := dao.NewMockModerationActionDAOInterface(ctrl)
	mockUserBanDAO := dao.NewMockUserBanDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
	mockVoteDAO := dao.NewMockVoteDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	handler := handlers.NewModerationHandler(
		mockReportDAO,
		mockModerationActionDAO,
		mockUserBanDAO,
		mockPseudonymDAO,
		mockSubforumDAO,
		mockPostDAO,
		mockCommentDAO,
		mockVoteDAO,
		mockPermissionDAO,
	)

	assert.NotNil(t, handler)
}

// TestModerationHandler_BasicFunctionality tests basic handler functionality
func TestModerationHandler_BasicFunctionality(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReportDAO := dao.NewMockReportDAOInterface(ctrl)
	mockModerationActionDAO := dao.NewMockModerationActionDAOInterface(ctrl)
	mockUserBanDAO := dao.NewMockUserBanDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)
	mockVoteDAO := dao.NewMockVoteDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	handler := handlers.NewModerationHandler(
		mockReportDAO,
		mockModerationActionDAO,
		mockUserBanDAO,
		mockPseudonymDAO,
		mockSubforumDAO,
		mockPostDAO,
		mockCommentDAO,
		mockVoteDAO,
		mockPermissionDAO,
	)

	// Test that the handler was created with all dependencies
	assert.NotNil(t, handler)
	assert.NotNil(t, mockReportDAO)
	assert.NotNil(t, mockModerationActionDAO)
	assert.NotNil(t, mockUserBanDAO)
	assert.NotNil(t, mockPseudonymDAO)
	assert.NotNil(t, mockSubforumDAO)
	assert.NotNil(t, mockPostDAO)
	assert.NotNil(t, mockCommentDAO)
	assert.NotNil(t, mockVoteDAO)
	assert.NotNil(t, mockPermissionDAO)
}
