// internal/repository/role_repo.go
package repository

import (
	"WMSS/user/api/internal/model"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	sysRoleRows = "`role_id`, `role_name`, `role_desc`, `create_time`, `update_time`, `deleted_at`"
)

type RoleRepository struct {
	model.SysRoleModel
	conn sqlx.SqlConn
}

func NewRoleRepository(conn sqlx.SqlConn, roleModel model.SysRoleModel) *RoleRepository {
	return &RoleRepository{
		SysRoleModel: roleModel,
		conn:         conn,
	}
}

// 分页查询角色
// internal/repository/role_repo.go
func (r *RoleRepository) FindByPage(ctx context.Context, page, pageSize int, roleName string) ([]*model.SysRole, int64, error) {
	// 详细的安全检查
	if r == nil {
		return nil, 0, fmt.Errorf("RoleRepository is nil")
	}
	if r.conn == nil {
		return nil, 0, fmt.Errorf("database connection is nil")
	}

	fmt.Printf("DEBUG: Starting FindByPage, page=%d, pageSize=%d, roleName=%s\n", page, pageSize, roleName)
	fmt.Printf("DEBUG: RoleRepository=%+v\n", r)
	fmt.Printf("DEBUG: conn=%+v\n", r.conn)

	var roles []*model.SysRole
	var total int64

	// 构建查询条件
	where := "WHERE deleted_at IS NULL"
	args := []interface{}{}

	if roleName != "" {
		where += " AND role_name LIKE ?"
		args = append(args, "%"+roleName+"%")
	}

	fmt.Printf("DEBUG: where clause: %s\n", where)
	fmt.Printf("DEBUG: args: %+v\n", args)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sys_role %s", where)
	fmt.Printf("DEBUG: countQuery: %s\n", countQuery)

	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		fmt.Printf("DEBUG: count query error: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("DEBUG: total count: %d\n", total)

	// 查询数据
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
        SELECT %s FROM sys_role 
        %s 
        ORDER BY create_time DESC 
        LIMIT ? OFFSET ?
    `, sysRoleRows, where)

	// 添加分页参数
	args = append(args, pageSize, offset)

	fmt.Printf("DEBUG: dataQuery: %s\n", dataQuery)
	fmt.Printf("DEBUG: final args: %+v\n", args)

	err = r.conn.QueryRowsCtx(ctx, &roles, dataQuery, args...)
	if err != nil {
		fmt.Printf("DEBUG: data query error: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("DEBUG: found %d roles\n", len(roles))
	return roles, total, nil
}

// 根据多个条件查询角色
func (r *RoleRepository) FindByConditions(ctx context.Context, conditions map[string]interface{}) ([]*model.SysRole, error) {
	var roles []*model.SysRole

	query := fmt.Sprintf("SELECT %s FROM sys_role WHERE deleted_at IS NULL", sysRoleRows)
	var args []interface{}

	// 动态构建查询条件
	if roleName, ok := conditions["role_name"]; ok && roleName != "" {
		query += " AND role_name LIKE ?"
		args = append(args, "%"+roleName.(string)+"%")
	}

	if startTime, ok := conditions["start_time"]; ok {
		query += " AND create_time >= ?"
		args = append(args, startTime)
	}

	if endTime, ok := conditions["end_time"]; ok {
		query += " AND create_time <= ?"
		args = append(args, endTime)
	}

	query += " ORDER BY create_time DESC"

	err := r.conn.QueryRowsCtx(ctx, &roles, query, args...)
	return roles, err
}

// 批量查询角色
func (r *RoleRepository) FindByIds(ctx context.Context, roleIds []string) ([]*model.SysRole, error) {
	if len(roleIds) == 0 {
		return []*model.SysRole{}, nil
	}

	var roles []*model.SysRole

	// 构建IN查询
	placeholders := make([]string, len(roleIds))
	args := make([]interface{}, len(roleIds))
	for i, id := range roleIds {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		"SELECT %s FROM sys_role WHERE role_id IN (%s) AND deleted_at IS NULL",
		sysRoleRows,
		strings.Join(placeholders, ","),
	)

	err := r.conn.QueryRowsCtx(ctx, &roles, query, args...)
	return roles, err
}

// 统计角色数量
func (r *RoleRepository) CountByTimeRange(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM sys_role WHERE create_time BETWEEN ? AND ? AND deleted_at IS NULL"
	err := r.conn.QueryRowCtx(ctx, &count, query, startTime, endTime)
	return count, err
}
