package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Cache key patterns
	// user_order:role:column:{userId}:{projectId}
	RoleColumnOrderKeyPattern = "user_order:role:column:%s:%s"

	// user_order:stage:column:{userId}:{projectId}
	StageColumnOrderKeyPattern = "user_order:stage:column:%s:%s"

	// user_order:role:kanban:{userId}:{projectId}:{roleId}
	RoleKanbanOrderKeyPattern = "user_order:role:kanban:%s:%s:%s"

	// user_order:stage:kanban:{userId}:{projectId}:{stageId}
	StageKanbanOrderKeyPattern = "user_order:stage:kanban:%s:%s:%s"

	// TTL: 1 hour
	UserOrderTTL = 1 * time.Hour
)

type UserOrderCache struct {
	client *redis.Client
}

func NewUserOrderCache(client *redis.Client) *UserOrderCache {
	return &UserOrderCache{
		client: client,
	}
}

// ==================== Role Column Order Cache ====================

func (c *UserOrderCache) GetRoleColumnOrder(ctx context.Context, userID, projectID string) ([]byte, error) {
	key := fmt.Sprintf(RoleColumnOrderKeyPattern, userID, projectID)
	return c.client.Get(ctx, key).Bytes()
}

func (c *UserOrderCache) SetRoleColumnOrder(ctx context.Context, userID, projectID string, data interface{}) error {
	key := fmt.Sprintf(RoleColumnOrderKeyPattern, userID, projectID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal role column order: %w", err)
	}
	return c.client.Set(ctx, key, jsonData, UserOrderTTL).Err()
}

func (c *UserOrderCache) DeleteRoleColumnOrder(ctx context.Context, userID, projectID string) error {
	key := fmt.Sprintf(RoleColumnOrderKeyPattern, userID, projectID)
	return c.client.Del(ctx, key).Err()
}

// ==================== Stage Column Order Cache ====================

func (c *UserOrderCache) GetStageColumnOrder(ctx context.Context, userID, projectID string) ([]byte, error) {
	key := fmt.Sprintf(StageColumnOrderKeyPattern, userID, projectID)
	return c.client.Get(ctx, key).Bytes()
}

func (c *UserOrderCache) SetStageColumnOrder(ctx context.Context, userID, projectID string, data interface{}) error {
	key := fmt.Sprintf(StageColumnOrderKeyPattern, userID, projectID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal stage column order: %w", err)
	}
	return c.client.Set(ctx, key, jsonData, UserOrderTTL).Err()
}

func (c *UserOrderCache) DeleteStageColumnOrder(ctx context.Context, userID, projectID string) error {
	key := fmt.Sprintf(StageColumnOrderKeyPattern, userID, projectID)
	return c.client.Del(ctx, key).Err()
}

// ==================== Role Kanban Order Cache ====================

func (c *UserOrderCache) GetRoleKanbanOrder(ctx context.Context, userID, projectID, roleID string) ([]byte, error) {
	key := fmt.Sprintf(RoleKanbanOrderKeyPattern, userID, projectID, roleID)
	return c.client.Get(ctx, key).Bytes()
}

func (c *UserOrderCache) SetRoleKanbanOrder(ctx context.Context, userID, projectID, roleID string, data interface{}) error {
	key := fmt.Sprintf(RoleKanbanOrderKeyPattern, userID, projectID, roleID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal role kanban order: %w", err)
	}
	return c.client.Set(ctx, key, jsonData, UserOrderTTL).Err()
}

func (c *UserOrderCache) DeleteRoleKanbanOrder(ctx context.Context, userID, projectID, roleID string) error {
	key := fmt.Sprintf(RoleKanbanOrderKeyPattern, userID, projectID, roleID)
	return c.client.Del(ctx, key).Err()
}

// ==================== Stage Kanban Order Cache ====================

func (c *UserOrderCache) GetStageKanbanOrder(ctx context.Context, userID, projectID, stageID string) ([]byte, error) {
	key := fmt.Sprintf(StageKanbanOrderKeyPattern, userID, projectID, stageID)
	return c.client.Get(ctx, key).Bytes()
}

func (c *UserOrderCache) SetStageKanbanOrder(ctx context.Context, userID, projectID, stageID string, data interface{}) error {
	key := fmt.Sprintf(StageKanbanOrderKeyPattern, userID, projectID, stageID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal stage kanban order: %w", err)
	}
	return c.client.Set(ctx, key, jsonData, UserOrderTTL).Err()
}

func (c *UserOrderCache) DeleteStageKanbanOrder(ctx context.Context, userID, projectID, stageID string) error {
	key := fmt.Sprintf(StageKanbanOrderKeyPattern, userID, projectID, stageID)
	return c.client.Del(ctx, key).Err()
}

// ==================== Invalidate All User Order Caches ====================

// InvalidateUserOrderCache invalidates all order-related caches for a user in a project
func (c *UserOrderCache) InvalidateUserOrderCache(ctx context.Context, userID, projectID string) error {
	// Delete role column order
	if err := c.DeleteRoleColumnOrder(ctx, userID, projectID); err != nil && err != redis.Nil {
		return fmt.Errorf("failed to delete role column order cache: %w", err)
	}

	// Delete stage column order
	if err := c.DeleteStageColumnOrder(ctx, userID, projectID); err != nil && err != redis.Nil {
		return fmt.Errorf("failed to delete stage column order cache: %w", err)
	}

	// Note: Individual kanban order caches (per role/stage) are not deleted here
	// They will be invalidated when specific orders are updated
	// This is an optimization to avoid deleting all possible combinations

	return nil
}

// InvalidateAllRoleKanbanCaches invalidates all role-based kanban order caches for a project
func (c *UserOrderCache) InvalidateAllRoleKanbanCaches(ctx context.Context, userID, projectID string, roleIDs []string) error {
	for _, roleID := range roleIDs {
		if err := c.DeleteRoleKanbanOrder(ctx, userID, projectID, roleID); err != nil && err != redis.Nil {
			return fmt.Errorf("failed to delete role kanban order cache for role %s: %w", roleID, err)
		}
	}
	return nil
}

// InvalidateAllStageKanbanCaches invalidates all stage-based kanban order caches for a project
func (c *UserOrderCache) InvalidateAllStageKanbanCaches(ctx context.Context, userID, projectID string, stageIDs []string) error {
	for _, stageID := range stageIDs {
		if err := c.DeleteStageKanbanOrder(ctx, userID, projectID, stageID); err != nil && err != redis.Nil {
			return fmt.Errorf("failed to delete stage kanban order cache for stage %s: %w", stageID, err)
		}
	}
	return nil
}
