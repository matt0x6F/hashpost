// Code generated by HashPost Generated Code. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/bob/orm"
	"github.com/stephenafamo/bob/types"
	"github.com/stephenafamo/bob/types/pgtypes"
)

// RoleKey is an object representing the database table.
type RoleKey struct {
	KeyID        uuid.UUID                   `db:"key_id,pk" scan:"key_id" json:"key_id"`
	RoleName     string                      `db:"role_name" scan:"role_name" json:"role_name"`
	Scope        string                      `db:"scope" scan:"scope" json:"scope"`
	KeyData      []byte                      `db:"key_data" scan:"key_data" json:"key_data"`
	KeyVersion   int32                       `db:"key_version" scan:"key_version" json:"key_version"`
	Capabilities types.JSON[json.RawMessage] `db:"capabilities" scan:"capabilities" json:"capabilities"`
	CreatedAt    sql.Null[time.Time]         `db:"created_at" scan:"created_at" json:"created_at"`
	ExpiresAt    time.Time                   `db:"expires_at" scan:"expires_at" json:"expires_at"`
	IsActive     sql.Null[bool]              `db:"is_active" scan:"is_active" json:"is_active"`
	CreatedBy    int64                       `db:"created_by" scan:"created_by" json:"created_by"`

	R roleKeyR `db:"-" scan:"rel" json:"rel"`
}

// RoleKeySlice is an alias for a slice of pointers to RoleKey.
// This should almost always be used instead of []*RoleKey.
type RoleKeySlice []*RoleKey

// RoleKeys contains methods to work with the role_keys table
var RoleKeys = psql.NewTablex[*RoleKey, RoleKeySlice, *RoleKeySetter]("", "role_keys")

// RoleKeysQuery is a query on the role_keys table
type RoleKeysQuery = *psql.ViewQuery[*RoleKey, RoleKeySlice]

// roleKeyR is where relationships are stored.
type roleKeyR struct {
	KeyKeyUsageAudits KeyUsageAuditSlice `scan:"KeyKeyUsageAudits" json:"KeyKeyUsageAudits"` // key_usage_audit.key_usage_audit_key_id_fkey
	CreatedByUser     *User              `scan:"CreatedByUser" json:"CreatedByUser"`         // role_keys.role_keys_created_by_fkey
}

type roleKeyColumnNames struct {
	KeyID        string
	RoleName     string
	Scope        string
	KeyData      string
	KeyVersion   string
	Capabilities string
	CreatedAt    string
	ExpiresAt    string
	IsActive     string
	CreatedBy    string
}

var RoleKeyColumns = buildRoleKeyColumns("role_keys")

type roleKeyColumns struct {
	tableAlias   string
	KeyID        psql.Expression
	RoleName     psql.Expression
	Scope        psql.Expression
	KeyData      psql.Expression
	KeyVersion   psql.Expression
	Capabilities psql.Expression
	CreatedAt    psql.Expression
	ExpiresAt    psql.Expression
	IsActive     psql.Expression
	CreatedBy    psql.Expression
}

func (c roleKeyColumns) Alias() string {
	return c.tableAlias
}

func (roleKeyColumns) AliasedAs(alias string) roleKeyColumns {
	return buildRoleKeyColumns(alias)
}

func buildRoleKeyColumns(alias string) roleKeyColumns {
	return roleKeyColumns{
		tableAlias:   alias,
		KeyID:        psql.Quote(alias, "key_id"),
		RoleName:     psql.Quote(alias, "role_name"),
		Scope:        psql.Quote(alias, "scope"),
		KeyData:      psql.Quote(alias, "key_data"),
		KeyVersion:   psql.Quote(alias, "key_version"),
		Capabilities: psql.Quote(alias, "capabilities"),
		CreatedAt:    psql.Quote(alias, "created_at"),
		ExpiresAt:    psql.Quote(alias, "expires_at"),
		IsActive:     psql.Quote(alias, "is_active"),
		CreatedBy:    psql.Quote(alias, "created_by"),
	}
}

type roleKeyWhere[Q psql.Filterable] struct {
	KeyID        psql.WhereMod[Q, uuid.UUID]
	RoleName     psql.WhereMod[Q, string]
	Scope        psql.WhereMod[Q, string]
	KeyData      psql.WhereMod[Q, []byte]
	KeyVersion   psql.WhereMod[Q, int32]
	Capabilities psql.WhereMod[Q, types.JSON[json.RawMessage]]
	CreatedAt    psql.WhereNullMod[Q, time.Time]
	ExpiresAt    psql.WhereMod[Q, time.Time]
	IsActive     psql.WhereNullMod[Q, bool]
	CreatedBy    psql.WhereMod[Q, int64]
}

func (roleKeyWhere[Q]) AliasedAs(alias string) roleKeyWhere[Q] {
	return buildRoleKeyWhere[Q](buildRoleKeyColumns(alias))
}

func buildRoleKeyWhere[Q psql.Filterable](cols roleKeyColumns) roleKeyWhere[Q] {
	return roleKeyWhere[Q]{
		KeyID:        psql.Where[Q, uuid.UUID](cols.KeyID),
		RoleName:     psql.Where[Q, string](cols.RoleName),
		Scope:        psql.Where[Q, string](cols.Scope),
		KeyData:      psql.Where[Q, []byte](cols.KeyData),
		KeyVersion:   psql.Where[Q, int32](cols.KeyVersion),
		Capabilities: psql.Where[Q, types.JSON[json.RawMessage]](cols.Capabilities),
		CreatedAt:    psql.WhereNull[Q, time.Time](cols.CreatedAt),
		ExpiresAt:    psql.Where[Q, time.Time](cols.ExpiresAt),
		IsActive:     psql.WhereNull[Q, bool](cols.IsActive),
		CreatedBy:    psql.Where[Q, int64](cols.CreatedBy),
	}
}

var RoleKeyErrors = &roleKeyErrors{
	ErrUniqueRoleKeysPkey: &UniqueConstraintError{
		schema:  "",
		table:   "role_keys",
		columns: []string{"key_id"},
		s:       "role_keys_pkey",
	},
}

type roleKeyErrors struct {
	ErrUniqueRoleKeysPkey *UniqueConstraintError
}

// RoleKeySetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type RoleKeySetter struct {
	KeyID        *uuid.UUID                   `db:"key_id,pk" scan:"key_id" json:"key_id"`
	RoleName     *string                      `db:"role_name" scan:"role_name" json:"role_name"`
	Scope        *string                      `db:"scope" scan:"scope" json:"scope"`
	KeyData      *[]byte                      `db:"key_data" scan:"key_data" json:"key_data"`
	KeyVersion   *int32                       `db:"key_version" scan:"key_version" json:"key_version"`
	Capabilities *types.JSON[json.RawMessage] `db:"capabilities" scan:"capabilities" json:"capabilities"`
	CreatedAt    *sql.Null[time.Time]         `db:"created_at" scan:"created_at" json:"created_at"`
	ExpiresAt    *time.Time                   `db:"expires_at" scan:"expires_at" json:"expires_at"`
	IsActive     *sql.Null[bool]              `db:"is_active" scan:"is_active" json:"is_active"`
	CreatedBy    *int64                       `db:"created_by" scan:"created_by" json:"created_by"`
}

func (s RoleKeySetter) SetColumns() []string {
	vals := make([]string, 0, 10)
	if s.KeyID != nil {
		vals = append(vals, "key_id")
	}

	if s.RoleName != nil {
		vals = append(vals, "role_name")
	}

	if s.Scope != nil {
		vals = append(vals, "scope")
	}

	if s.KeyData != nil {
		vals = append(vals, "key_data")
	}

	if s.KeyVersion != nil {
		vals = append(vals, "key_version")
	}

	if s.Capabilities != nil {
		vals = append(vals, "capabilities")
	}

	if s.CreatedAt != nil {
		vals = append(vals, "created_at")
	}

	if s.ExpiresAt != nil {
		vals = append(vals, "expires_at")
	}

	if s.IsActive != nil {
		vals = append(vals, "is_active")
	}

	if s.CreatedBy != nil {
		vals = append(vals, "created_by")
	}

	return vals
}

func (s RoleKeySetter) Overwrite(t *RoleKey) {
	if s.KeyID != nil {
		t.KeyID = *s.KeyID
	}
	if s.RoleName != nil {
		t.RoleName = *s.RoleName
	}
	if s.Scope != nil {
		t.Scope = *s.Scope
	}
	if s.KeyData != nil {
		t.KeyData = *s.KeyData
	}
	if s.KeyVersion != nil {
		t.KeyVersion = *s.KeyVersion
	}
	if s.Capabilities != nil {
		t.Capabilities = *s.Capabilities
	}
	if s.CreatedAt != nil {
		t.CreatedAt = *s.CreatedAt
	}
	if s.ExpiresAt != nil {
		t.ExpiresAt = *s.ExpiresAt
	}
	if s.IsActive != nil {
		t.IsActive = *s.IsActive
	}
	if s.CreatedBy != nil {
		t.CreatedBy = *s.CreatedBy
	}
}

func (s *RoleKeySetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return RoleKeys.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 10)
		if s.KeyID != nil {
			vals[0] = psql.Arg(*s.KeyID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.RoleName != nil {
			vals[1] = psql.Arg(*s.RoleName)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.Scope != nil {
			vals[2] = psql.Arg(*s.Scope)
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.KeyData != nil {
			vals[3] = psql.Arg(*s.KeyData)
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		if s.KeyVersion != nil {
			vals[4] = psql.Arg(*s.KeyVersion)
		} else {
			vals[4] = psql.Raw("DEFAULT")
		}

		if s.Capabilities != nil {
			vals[5] = psql.Arg(*s.Capabilities)
		} else {
			vals[5] = psql.Raw("DEFAULT")
		}

		if s.CreatedAt != nil {
			vals[6] = psql.Arg(*s.CreatedAt)
		} else {
			vals[6] = psql.Raw("DEFAULT")
		}

		if s.ExpiresAt != nil {
			vals[7] = psql.Arg(*s.ExpiresAt)
		} else {
			vals[7] = psql.Raw("DEFAULT")
		}

		if s.IsActive != nil {
			vals[8] = psql.Arg(*s.IsActive)
		} else {
			vals[8] = psql.Raw("DEFAULT")
		}

		if s.CreatedBy != nil {
			vals[9] = psql.Arg(*s.CreatedBy)
		} else {
			vals[9] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s RoleKeySetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s RoleKeySetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 10)

	if s.KeyID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "key_id")...),
			psql.Arg(s.KeyID),
		}})
	}

	if s.RoleName != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "role_name")...),
			psql.Arg(s.RoleName),
		}})
	}

	if s.Scope != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "scope")...),
			psql.Arg(s.Scope),
		}})
	}

	if s.KeyData != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "key_data")...),
			psql.Arg(s.KeyData),
		}})
	}

	if s.KeyVersion != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "key_version")...),
			psql.Arg(s.KeyVersion),
		}})
	}

	if s.Capabilities != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "capabilities")...),
			psql.Arg(s.Capabilities),
		}})
	}

	if s.CreatedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "created_at")...),
			psql.Arg(s.CreatedAt),
		}})
	}

	if s.ExpiresAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "expires_at")...),
			psql.Arg(s.ExpiresAt),
		}})
	}

	if s.IsActive != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "is_active")...),
			psql.Arg(s.IsActive),
		}})
	}

	if s.CreatedBy != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "created_by")...),
			psql.Arg(s.CreatedBy),
		}})
	}

	return exprs
}

// FindRoleKey retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindRoleKey(ctx context.Context, exec bob.Executor, KeyIDPK uuid.UUID, cols ...string) (*RoleKey, error) {
	if len(cols) == 0 {
		return RoleKeys.Query(
			SelectWhere.RoleKeys.KeyID.EQ(KeyIDPK),
		).One(ctx, exec)
	}

	return RoleKeys.Query(
		SelectWhere.RoleKeys.KeyID.EQ(KeyIDPK),
		sm.Columns(RoleKeys.Columns().Only(cols...)),
	).One(ctx, exec)
}

// RoleKeyExists checks the presence of a single record by primary key
func RoleKeyExists(ctx context.Context, exec bob.Executor, KeyIDPK uuid.UUID) (bool, error) {
	return RoleKeys.Query(
		SelectWhere.RoleKeys.KeyID.EQ(KeyIDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after RoleKey is retrieved from the database
func (o *RoleKey) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = RoleKeys.AfterSelectHooks.RunHooks(ctx, exec, RoleKeySlice{o})
	case bob.QueryTypeInsert:
		ctx, err = RoleKeys.AfterInsertHooks.RunHooks(ctx, exec, RoleKeySlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = RoleKeys.AfterUpdateHooks.RunHooks(ctx, exec, RoleKeySlice{o})
	case bob.QueryTypeDelete:
		ctx, err = RoleKeys.AfterDeleteHooks.RunHooks(ctx, exec, RoleKeySlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the RoleKey
func (o *RoleKey) primaryKeyVals() bob.Expression {
	return psql.Arg(o.KeyID)
}

func (o *RoleKey) pkEQ() dialect.Expression {
	return psql.Quote("role_keys", "key_id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the RoleKey
func (o *RoleKey) Update(ctx context.Context, exec bob.Executor, s *RoleKeySetter) error {
	v, err := RoleKeys.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single RoleKey record with an executor
func (o *RoleKey) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := RoleKeys.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the RoleKey using the executor
func (o *RoleKey) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := RoleKeys.Query(
		SelectWhere.RoleKeys.KeyID.EQ(o.KeyID),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after RoleKeySlice is retrieved from the database
func (o RoleKeySlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = RoleKeys.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = RoleKeys.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = RoleKeys.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = RoleKeys.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o RoleKeySlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("role_keys", "key_id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		pkPairs := make([]bob.Expression, len(o))
		for i, row := range o {
			pkPairs[i] = row.primaryKeyVals()
		}
		return bob.ExpressSlice(ctx, w, d, start, pkPairs, "", ", ", "")
	}))
}

// copyMatchingRows finds models in the given slice that have the same primary key
// then it first copies the existing relationships from the old model to the new model
// and then replaces the old model in the slice with the new model
func (o RoleKeySlice) copyMatchingRows(from ...*RoleKey) {
	for i, old := range o {
		for _, new := range from {
			if new.KeyID != old.KeyID {
				continue
			}
			new.R = old.R
			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o RoleKeySlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return RoleKeys.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *RoleKey:
				o.copyMatchingRows(retrieved)
			case []*RoleKey:
				o.copyMatchingRows(retrieved...)
			case RoleKeySlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a RoleKey or a slice of RoleKey
				// then run the AfterUpdateHooks on the slice
				_, err = RoleKeys.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o RoleKeySlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return RoleKeys.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *RoleKey:
				o.copyMatchingRows(retrieved)
			case []*RoleKey:
				o.copyMatchingRows(retrieved...)
			case RoleKeySlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a RoleKey or a slice of RoleKey
				// then run the AfterDeleteHooks on the slice
				_, err = RoleKeys.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o RoleKeySlice) UpdateAll(ctx context.Context, exec bob.Executor, vals RoleKeySetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := RoleKeys.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o RoleKeySlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := RoleKeys.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o RoleKeySlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := RoleKeys.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

type roleKeyJoins[Q dialect.Joinable] struct {
	typ               string
	KeyKeyUsageAudits modAs[Q, keyUsageAuditColumns]
	CreatedByUser     modAs[Q, userColumns]
}

func (j roleKeyJoins[Q]) aliasedAs(alias string) roleKeyJoins[Q] {
	return buildRoleKeyJoins[Q](buildRoleKeyColumns(alias), j.typ)
}

func buildRoleKeyJoins[Q dialect.Joinable](cols roleKeyColumns, typ string) roleKeyJoins[Q] {
	return roleKeyJoins[Q]{
		typ: typ,
		KeyKeyUsageAudits: modAs[Q, keyUsageAuditColumns]{
			c: KeyUsageAuditColumns,
			f: func(to keyUsageAuditColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, KeyUsageAudits.Name().As(to.Alias())).On(
						to.KeyID.EQ(cols.KeyID),
					))
				}

				return mods
			},
		},
		CreatedByUser: modAs[Q, userColumns]{
			c: UserColumns,
			f: func(to userColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Users.Name().As(to.Alias())).On(
						to.UserID.EQ(cols.CreatedBy),
					))
				}

				return mods
			},
		},
	}
}

// KeyKeyUsageAudits starts a query for related objects on key_usage_audit
func (o *RoleKey) KeyKeyUsageAudits(mods ...bob.Mod[*dialect.SelectQuery]) KeyUsageAuditsQuery {
	return KeyUsageAudits.Query(append(mods,
		sm.Where(KeyUsageAuditColumns.KeyID.EQ(psql.Arg(o.KeyID))),
	)...)
}

func (os RoleKeySlice) KeyKeyUsageAudits(mods ...bob.Mod[*dialect.SelectQuery]) KeyUsageAuditsQuery {
	pkKeyID := make(pgtypes.Array[uuid.UUID], len(os))
	for i, o := range os {
		pkKeyID[i] = o.KeyID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkKeyID), "uuid[]")),
	))

	return KeyUsageAudits.Query(append(mods,
		sm.Where(psql.Group(KeyUsageAuditColumns.KeyID).OP("IN", PKArgExpr)),
	)...)
}

// CreatedByUser starts a query for related objects on users
func (o *RoleKey) CreatedByUser(mods ...bob.Mod[*dialect.SelectQuery]) UsersQuery {
	return Users.Query(append(mods,
		sm.Where(UserColumns.UserID.EQ(psql.Arg(o.CreatedBy))),
	)...)
}

func (os RoleKeySlice) CreatedByUser(mods ...bob.Mod[*dialect.SelectQuery]) UsersQuery {
	pkCreatedBy := make(pgtypes.Array[int64], len(os))
	for i, o := range os {
		pkCreatedBy[i] = o.CreatedBy
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkCreatedBy), "bigint[]")),
	))

	return Users.Query(append(mods,
		sm.Where(psql.Group(UserColumns.UserID).OP("IN", PKArgExpr)),
	)...)
}

func (o *RoleKey) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "KeyKeyUsageAudits":
		rels, ok := retrieved.(KeyUsageAuditSlice)
		if !ok {
			return fmt.Errorf("roleKey cannot load %T as %q", retrieved, name)
		}

		o.R.KeyKeyUsageAudits = rels

		for _, rel := range rels {
			if rel != nil {
				rel.R.KeyRoleKey = o
			}
		}
		return nil
	case "CreatedByUser":
		rel, ok := retrieved.(*User)
		if !ok {
			return fmt.Errorf("roleKey cannot load %T as %q", retrieved, name)
		}

		o.R.CreatedByUser = rel

		if rel != nil {
			rel.R.CreatedByRoleKeys = RoleKeySlice{o}
		}
		return nil
	default:
		return fmt.Errorf("roleKey has no relationship %q", name)
	}
}

type roleKeyPreloader struct {
	CreatedByUser func(...psql.PreloadOption) psql.Preloader
}

func buildRoleKeyPreloader() roleKeyPreloader {
	return roleKeyPreloader{
		CreatedByUser: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*User, UserSlice](orm.Relationship{
				Name: "CreatedByUser",
				Sides: []orm.RelSide{
					{
						From: TableNames.RoleKeys,
						To:   TableNames.Users,
						FromColumns: []string{
							ColumnNames.RoleKeys.CreatedBy,
						},
						ToColumns: []string{
							ColumnNames.Users.UserID,
						},
					},
				},
			}, Users.Columns().Names(), opts...)
		},
	}
}

type roleKeyThenLoader[Q orm.Loadable] struct {
	KeyKeyUsageAudits func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
	CreatedByUser     func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildRoleKeyThenLoader[Q orm.Loadable]() roleKeyThenLoader[Q] {
	type KeyKeyUsageAuditsLoadInterface interface {
		LoadKeyKeyUsageAudits(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}
	type CreatedByUserLoadInterface interface {
		LoadCreatedByUser(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return roleKeyThenLoader[Q]{
		KeyKeyUsageAudits: thenLoadBuilder[Q](
			"KeyKeyUsageAudits",
			func(ctx context.Context, exec bob.Executor, retrieved KeyKeyUsageAuditsLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadKeyKeyUsageAudits(ctx, exec, mods...)
			},
		),
		CreatedByUser: thenLoadBuilder[Q](
			"CreatedByUser",
			func(ctx context.Context, exec bob.Executor, retrieved CreatedByUserLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadCreatedByUser(ctx, exec, mods...)
			},
		),
	}
}

// LoadKeyKeyUsageAudits loads the roleKey's KeyKeyUsageAudits into the .R struct
func (o *RoleKey) LoadKeyKeyUsageAudits(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.KeyKeyUsageAudits = nil

	related, err := o.KeyKeyUsageAudits(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, rel := range related {
		rel.R.KeyRoleKey = o
	}

	o.R.KeyKeyUsageAudits = related
	return nil
}

// LoadKeyKeyUsageAudits loads the roleKey's KeyKeyUsageAudits into the .R struct
func (os RoleKeySlice) LoadKeyKeyUsageAudits(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	keyUsageAudits, err := os.KeyKeyUsageAudits(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		o.R.KeyKeyUsageAudits = nil
	}

	for _, o := range os {
		for _, rel := range keyUsageAudits {
			if o.KeyID != rel.KeyID {
				continue
			}

			rel.R.KeyRoleKey = o

			o.R.KeyKeyUsageAudits = append(o.R.KeyKeyUsageAudits, rel)
		}
	}

	return nil
}

// LoadCreatedByUser loads the roleKey's CreatedByUser into the .R struct
func (o *RoleKey) LoadCreatedByUser(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.CreatedByUser = nil

	related, err := o.CreatedByUser(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	related.R.CreatedByRoleKeys = RoleKeySlice{o}

	o.R.CreatedByUser = related
	return nil
}

// LoadCreatedByUser loads the roleKey's CreatedByUser into the .R struct
func (os RoleKeySlice) LoadCreatedByUser(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	users, err := os.CreatedByUser(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		for _, rel := range users {
			if o.CreatedBy != rel.UserID {
				continue
			}

			rel.R.CreatedByRoleKeys = append(rel.R.CreatedByRoleKeys, o)

			o.R.CreatedByUser = rel
			break
		}
	}

	return nil
}

func insertRoleKeyKeyKeyUsageAudits0(ctx context.Context, exec bob.Executor, keyUsageAudits1 []*KeyUsageAuditSetter, roleKey0 *RoleKey) (KeyUsageAuditSlice, error) {
	for i := range keyUsageAudits1 {
		keyUsageAudits1[i].KeyID = &roleKey0.KeyID
	}

	ret, err := KeyUsageAudits.Insert(bob.ToMods(keyUsageAudits1...)).All(ctx, exec)
	if err != nil {
		return ret, fmt.Errorf("insertRoleKeyKeyKeyUsageAudits0: %w", err)
	}

	return ret, nil
}

func attachRoleKeyKeyKeyUsageAudits0(ctx context.Context, exec bob.Executor, count int, keyUsageAudits1 KeyUsageAuditSlice, roleKey0 *RoleKey) (KeyUsageAuditSlice, error) {
	setter := &KeyUsageAuditSetter{
		KeyID: &roleKey0.KeyID,
	}

	err := keyUsageAudits1.UpdateAll(ctx, exec, *setter)
	if err != nil {
		return nil, fmt.Errorf("attachRoleKeyKeyKeyUsageAudits0: %w", err)
	}

	return keyUsageAudits1, nil
}

func (roleKey0 *RoleKey) InsertKeyKeyUsageAudits(ctx context.Context, exec bob.Executor, related ...*KeyUsageAuditSetter) error {
	if len(related) == 0 {
		return nil
	}

	var err error

	keyUsageAudits1, err := insertRoleKeyKeyKeyUsageAudits0(ctx, exec, related, roleKey0)
	if err != nil {
		return err
	}

	roleKey0.R.KeyKeyUsageAudits = append(roleKey0.R.KeyKeyUsageAudits, keyUsageAudits1...)

	for _, rel := range keyUsageAudits1 {
		rel.R.KeyRoleKey = roleKey0
	}
	return nil
}

func (roleKey0 *RoleKey) AttachKeyKeyUsageAudits(ctx context.Context, exec bob.Executor, related ...*KeyUsageAudit) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	keyUsageAudits1 := KeyUsageAuditSlice(related)

	_, err = attachRoleKeyKeyKeyUsageAudits0(ctx, exec, len(related), keyUsageAudits1, roleKey0)
	if err != nil {
		return err
	}

	roleKey0.R.KeyKeyUsageAudits = append(roleKey0.R.KeyKeyUsageAudits, keyUsageAudits1...)

	for _, rel := range related {
		rel.R.KeyRoleKey = roleKey0
	}

	return nil
}

func attachRoleKeyCreatedByUser0(ctx context.Context, exec bob.Executor, count int, roleKey0 *RoleKey, user1 *User) (*RoleKey, error) {
	setter := &RoleKeySetter{
		CreatedBy: &user1.UserID,
	}

	err := roleKey0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachRoleKeyCreatedByUser0: %w", err)
	}

	return roleKey0, nil
}

func (roleKey0 *RoleKey) InsertCreatedByUser(ctx context.Context, exec bob.Executor, related *UserSetter) error {
	user1, err := Users.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachRoleKeyCreatedByUser0(ctx, exec, 1, roleKey0, user1)
	if err != nil {
		return err
	}

	roleKey0.R.CreatedByUser = user1

	user1.R.CreatedByRoleKeys = append(user1.R.CreatedByRoleKeys, roleKey0)

	return nil
}

func (roleKey0 *RoleKey) AttachCreatedByUser(ctx context.Context, exec bob.Executor, user1 *User) error {
	var err error

	_, err = attachRoleKeyCreatedByUser0(ctx, exec, 1, roleKey0, user1)
	if err != nil {
		return err
	}

	roleKey0.R.CreatedByUser = user1

	user1.R.CreatedByRoleKeys = append(user1.R.CreatedByRoleKeys, roleKey0)

	return nil
}
