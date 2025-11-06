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
	RoleBoardOrderKeyPattern = "user_order:role:kanban:%s:%s:%s"

	// user_order:stage:kanban:{userId}:{projectId}:{stageId}
	StageBoardOrderKeyPattern = "user_order:stage:kanban:%s:%s:%s"

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

// ==================== Role Board Order Cache ====================

func (c *UserOrderCache) GetRoleBoardOrder(ctx context.Context, userID, projectID, roleID string) ([]byte, error) {
	key := fmt.Sprintf(RoleBoardOrderKeyPattern, userID, projectID, roleID)
	return c.client.Get(ctx, key).Bytes()
}

func (c *UserOrderCache) SetRoleBoardOrder(ctx context.Context, userID, projectID, roleID string, data interface{}) error {
	key := fmt.Sprintf(RoleBoardOrderKeyPattern, userID, projectID, roleID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal role board order: %w", err)
	}
	return c.client.Set(ctx, key, jsonData, UserOrderTTL).Err()
}

func (c *UserOrderCache) DeleteRoleBoardOrder(ctx context.Context, userID, projectID, roleID string) error {
	key := fmt.Sprintf(RoleBoardOrderKeyPattern, userID, projectID, roleID)
	return c.client.Del(ctx, key).Err()
}

// ==================== Stage Board Order Cache ====================

func (c *UserOrderCache) GetStageBoardOrder(ctx context.Context, userID, projectID, stageID string) ([]byte, error) {
	key := fmt.Sprintf(StageBoardOrderKeyPattern, userID, projectID, stageID)
	return c.client.Get(ctx, key).Bytes()
}

func (c *UserOrderCache) SetStageBoardOrder(ctx context.Context, userID, projectID, stageID string, data interface{}) error {
	key := fmt.Sprintf(StageBoardOrderKeyPattern, userID, projectID, stageID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal stage board order: %w", err)
	}
	return c.client.Set(ctx, key, jsonData, UserOrderTTL).Err()
}

func (c *UserOrderCache) DeleteStageBoardOrder(ctx context.Context, userID, projectID, stageID string) error {
	key := fmt.Sprintf(StageBoardOrderKeyPattern, userID, projectID, stageID)
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

	// Note: Individual board order caches (per role/stage) are not deleted here
	// They will be invalidated when specific orders are updated
	// This is an optimization to avoid deleting all possible combinations

	return nil
}

// InvalidateAllRoleBoardCaches invalidates all role-based board order caches for a project
func (c *UserOrderCache) InvalidateAllRoleBoardCaches(ctx context.Context, userID, projectID string, roleIDs []string) error {
	for _, roleID := range roleIDs {
		if err := c.DeleteRoleBoardOrder(ctx, userID, projectID, roleID); err != nil && err != redis.Nil {
			return fmt.Errorf("failed to delete role board order cache for role %s: %w", roleID, err)
		}
	}
	return nil
}

// InvalidateAllStageBoardCaches invalidates all stage-based board order caches for a project
func (c *UserOrderCache) InvalidateAllStageBoardCaches(ctx context.Context, userID, projectID string, stageIDs []string) error {
	for _, stageID := range stageIDs {
		if err := c.DeleteStageBoardOrder(ctx, userID, projectID, stageID); err != nil && err != redis.Nil {
			return fmt.Errorf("failed to delete stage board order cache for stage %s: %w", stageID, err)
		}
	}
	return nil
}
