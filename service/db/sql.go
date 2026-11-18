package db

import (
	"context"
	"fmt"
	"time"

	"nextopsx/domain/connector"
	"nextopsx/domain/dbexec"
)

// ExecuteSQL 在指定数据库资产上执行一条 SQL，返回结构化结果。
// 查询类（SELECT/SHOW 等）返回列与行；写类返回受影响行数。
// 非 dbexec 的 DB 驱动（如 redis）走 connector 注册表统一按查询语义执行。
func (s *Service) ExecuteSQL(hostID, sqlText string) (*dbexec.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if ex := s.resolver.DbFor(hostID); ex != nil {
		if dbexec.IsQuery(sqlText) {
			return ex.Query(ctx, sqlText)
		}
		return ex.Exec(ctx, sqlText)
	}
	// Redis 等注册表 DB 驱动：Redis 命令无查询/写语法之分，统一 Query。
	res, err := s.queryFromRegistry(ctx, hostID, sqlText)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// queryFromRegistry 经 connector 注册表构造 DB 资产能力并执行查询。
func (s *Service) queryFromRegistry(ctx context.Context, hostID, cmd string) (*dbexec.Result, error) {
	cap := s.resolver.ConnFor(hostID)
	qr, ok := cap.(connector.QueryRunner)
	if !ok {
		return nil, fmt.Errorf("无法解析数据库资产，请确认资产为数据库协议且凭据已录入")
	}
	res, err := qr.Query(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return &dbexec.Result{Columns: res.Columns, Rows: res.Rows}, nil
}

// ListSchema 返回指定数据库资产的表/列结构（供前端左侧 schema 树）。
// 非 dbexec 的 DB 驱动（如 redis，无表结构概念）返回空 schema，前端树为空而非报错。
func (s *Service) ListSchema(hostID string) (*dbexec.Schema, error) {
	if ex := s.resolver.DbFor(hostID); ex != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return ex.Schema(ctx)
	}
	if cap := s.resolver.ConnFor(hostID); cap != nil {
		return &dbexec.Schema{}, nil
	}
	return nil, fmt.Errorf("无法解析数据库资产，请确认资产为数据库协议且凭据已录入")
}
