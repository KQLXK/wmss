package repository

import (
	"WMSS/user/api/internal/model"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	sysPermissionRows = "`permission_id`, `permission_name`, `permission_code`,`permission_type`," +
		"`parent_permission_id`,`create_time`, `update_time`, `deleted_at`"
)

type PermissionRepository struct {
	model.SysPermissionModel
	conn sqlx.SqlConn
}

func NewPermissionRepository(conn sqlx.SqlConn, permissionModel model.SysPermissionModel) *PermissionRepository {
	return &PermissionRepository{
		SysPermissionModel: permissionModel,
		conn:               conn,
	}
}

// 分页查询权限
// 假设 sysPermissionRows 是预定义的查询字段列表（避免SELECT *，提升性能和安全性）

// FindByPage 分页查询权限（SysPermission）
// page: 页码（从1开始）
// pageSize: 每页条数
// permissionType: 权限类型筛选（可选，如"菜单权限""按钮权限"）
// roleName: 角色名称筛选（可选，查询该角色关联的权限；无需则删除该参数）
func (r *PermissionRepository) FindByPage(
	ctx context.Context,
	page, pageSize int,
	permissionType string, // 权限类型筛选（可选）
	roleName string, // 角色名称筛选（可选，关联中间表）
) ([]*model.SysPermission, int64, error) {
	// 原有安全检查逻辑不变（适配PermissionRepository）
	if r == nil {
		return nil, 0, fmt.Errorf("PermissionRepository is nil")
	}
	if r.conn == nil {
		return nil, 0, fmt.Errorf("database connection is nil")
	}

	// 调试日志（更新参数打印）
	fmt.Printf("DEBUG: Starting FindByPage (Permission), page=%d, pageSize=%d, permissionType=%s, roleName=%s\n",
		page, pageSize, permissionType, roleName)

	var permissions []*model.SysPermission
	var total int64

	// 构建查询基础：主表（sys_permission）+ 软删除过滤
	baseFrom := "FROM sys_permission p"
	joinClause := "" // 关联中间表（角色-权限）的子句（按需启用）
	where := "WHERE p.deleted_at IS NULL"
	args := []interface{}{}

	// 1. 权限类型筛选（直接查sys_permission表的permission_type字段）
	if permissionType != "" {
		where += " AND p.permission_type = ?" // 精确匹配；需模糊匹配则改为 LIKE ? 并拼接 %
		args = append(args, permissionType)
		// 模糊匹配示例：where += " AND p.permission_type LIKE ?"; args = append(args, "%"+permissionType+"%")
	}

	// 2. 角色名称筛选（关联角色-权限中间表 sys_role_permission + 角色表 sys_role）
	if roleName != "" {
		// 关联逻辑：sys_permission ←→ sys_role_permission（权限-角色关联）←→ sys_role（角色表）
		joinClause = " JOIN sys_role_permission rp ON p.permission_id = rp.permission_id " +
			" JOIN sys_role r ON rp.role_id = r.id "
		// 角色名称模糊匹配（如输入"管理员"匹配"超级管理员""普通管理员"）
		where += " AND r.role_name LIKE ?"
		args = append(args, "%"+roleName+"%")
	}

	// -------------------------- 第一步：查询总数（需去重，避免关联导致重复计数）--------------------------
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT p.permission_id) %s %s %s",
		baseFrom, joinClause, where)
	fmt.Printf("DEBUG: countQuery: %s\n", countQuery)
	fmt.Printf("DEBUG: count args: %+v\n", args)

	// 执行总数查询
	err := r.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		fmt.Printf("DEBUG: permission count query error: %v\n", err)
		return nil, 0, err
	}
	fmt.Printf("DEBUG: total permission count: %d\n", total)

	// -------------------------- 第二步：查询分页数据 --------------------------
	// 计算偏移量（页码从1开始，offset = (page-1)*pageSize）
	offset := (page - 1) * pageSize
	// 构建数据查询SQL（DISTINCT去重，避免关联中间表导致重复权限）
	dataQuery := fmt.Sprintf(`
        SELECT DISTINCT %s 
        %s %s %s
        ORDER BY p.create_time DESC 
        LIMIT ? OFFSET ?
    `, sysPermissionRows, baseFrom, joinClause, where)

	// 追加分页参数（pageSize和offset）
	args = append(args, pageSize, offset)

	// 调试日志
	fmt.Printf("DEBUG: dataQuery: %s\n", dataQuery)
	fmt.Printf("DEBUG: final args: %+v\n", args)

	// 执行分页查询（结果存入permissions切片）
	err = r.conn.QueryRowsCtx(ctx, &permissions, dataQuery, args...)
	if err != nil {
		fmt.Printf("DEBUG: permission data query error: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("DEBUG: found %d permissions\n", len(permissions))
	return permissions, total, nil
}

// 逻辑删除
func (r *PermissionRepository) DeleteLogical(ctx context.Context, PermissionId string) error {
	query := fmt.Sprintf("update sys_permission set deleted_at = NOW() WHERE permission_id = ?")
	_, err := r.conn.ExecCtx(ctx, query, PermissionId)
	return err
}

//// 根据多个条件查询角色
//func (r *RoleRepository) FindByConditions(ctx context.Context, conditions map[string]interface{}) ([]*model.SysRole, error) {
//	var roles []*model.SysRole
//
//	query := fmt.Sprintf("SELECT %s FROM sys_role WHERE deleted_at IS NULL", sysRoleRows)
//	var args []interface{}
//
//	// 动态构建查询条件
//	if roleName, ok := conditions["role_name"]; ok && roleName != "" {
//		query += " AND role_name LIKE ?"
//		args = append(args, "%"+roleName.(string)+"%")
//	}
//
//	if startTime, ok := conditions["start_time"]; ok {
//		query += " AND create_time >= ?"
//		args = append(args, startTime)
//	}
//
//	if endTime, ok := conditions["end_time"]; ok {
//		query += " AND create_time <= ?"
//		args = append(args, endTime)
//	}
//
//	query += " ORDER BY create_time DESC"
//
//	err := r.conn.QueryRowsCtx(ctx, &roles, query, args...)
//	return roles, err
//}
//
//// 批量查询角色
//func (r *RoleRepository) FindByIds(ctx context.Context, roleIds []string) ([]*model.SysRole, error) {
//	if len(roleIds) == 0 {
//		return []*model.SysRole{}, nil
//	}
//
//	var roles []*model.SysRole
//
//	// 构建IN查询
//	placeholders := make([]string, len(roleIds))
//	args := make([]interface{}, len(roleIds))
//	for i, id := range roleIds {
//		placeholders[i] = "?"
//		args[i] = id
//	}
//
//	query := fmt.Sprintf(
//		"SELECT %s FROM sys_role WHERE role_id IN (%s) AND deleted_at IS NULL",
//		sysRoleRows,
//		strings.Join(placeholders, ","),
//	)
//
//	err := r.conn.QueryRowsCtx(ctx, &roles, query, args...)
//	return roles, err
//}
//
//// 统计角色数量
//func (r *RoleRepository) CountByTimeRange(ctx context.Context, startTime, endTime time.Time) (int64, error) {
//	var count int64
//	query := "SELECT COUNT(*) FROM sys_role WHERE create_time BETWEEN ? AND ? AND deleted_at IS NULL"
//	err := r.conn.QueryRowCtx(ctx, &count, query, startTime, endTime)
//	return count, err
//}
