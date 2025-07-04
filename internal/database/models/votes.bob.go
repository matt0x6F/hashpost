// Code generated by HashPost Generated Code. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/bob/orm"
	"github.com/stephenafamo/bob/types/pgtypes"
)

// Vote is an object representing the database table.
type Vote struct {
	VoteID      int64               `db:"vote_id,pk" scan:"vote_id" json:"vote_id"`
	PseudonymID string              `db:"pseudonym_id" scan:"pseudonym_id" json:"pseudonym_id"`
	ContentType string              `db:"content_type" scan:"content_type" json:"content_type"`
	ContentID   int64               `db:"content_id" scan:"content_id" json:"content_id"`
	VoteValue   int32               `db:"vote_value" scan:"vote_value" json:"vote_value"`
	CreatedAt   sql.Null[time.Time] `db:"created_at" scan:"created_at" json:"created_at"`
	UpdatedAt   sql.Null[time.Time] `db:"updated_at" scan:"updated_at" json:"updated_at"`

	R voteR `db:"-" scan:"rel" json:"rel"`
}

// VoteSlice is an alias for a slice of pointers to Vote.
// This should almost always be used instead of []*Vote.
type VoteSlice []*Vote

// Votes contains methods to work with the votes table
var Votes = psql.NewTablex[*Vote, VoteSlice, *VoteSetter]("", "votes")

// VotesQuery is a query on the votes table
type VotesQuery = *psql.ViewQuery[*Vote, VoteSlice]

// voteR is where relationships are stored.
type voteR struct {
	Pseudonym *Pseudonym `scan:"Pseudonym" json:"Pseudonym"` // votes.votes_pseudonym_id_fkey
}

type voteColumnNames struct {
	VoteID      string
	PseudonymID string
	ContentType string
	ContentID   string
	VoteValue   string
	CreatedAt   string
	UpdatedAt   string
}

var VoteColumns = buildVoteColumns("votes")

type voteColumns struct {
	tableAlias  string
	VoteID      psql.Expression
	PseudonymID psql.Expression
	ContentType psql.Expression
	ContentID   psql.Expression
	VoteValue   psql.Expression
	CreatedAt   psql.Expression
	UpdatedAt   psql.Expression
}

func (c voteColumns) Alias() string {
	return c.tableAlias
}

func (voteColumns) AliasedAs(alias string) voteColumns {
	return buildVoteColumns(alias)
}

func buildVoteColumns(alias string) voteColumns {
	return voteColumns{
		tableAlias:  alias,
		VoteID:      psql.Quote(alias, "vote_id"),
		PseudonymID: psql.Quote(alias, "pseudonym_id"),
		ContentType: psql.Quote(alias, "content_type"),
		ContentID:   psql.Quote(alias, "content_id"),
		VoteValue:   psql.Quote(alias, "vote_value"),
		CreatedAt:   psql.Quote(alias, "created_at"),
		UpdatedAt:   psql.Quote(alias, "updated_at"),
	}
}

type voteWhere[Q psql.Filterable] struct {
	VoteID      psql.WhereMod[Q, int64]
	PseudonymID psql.WhereMod[Q, string]
	ContentType psql.WhereMod[Q, string]
	ContentID   psql.WhereMod[Q, int64]
	VoteValue   psql.WhereMod[Q, int32]
	CreatedAt   psql.WhereNullMod[Q, time.Time]
	UpdatedAt   psql.WhereNullMod[Q, time.Time]
}

func (voteWhere[Q]) AliasedAs(alias string) voteWhere[Q] {
	return buildVoteWhere[Q](buildVoteColumns(alias))
}

func buildVoteWhere[Q psql.Filterable](cols voteColumns) voteWhere[Q] {
	return voteWhere[Q]{
		VoteID:      psql.Where[Q, int64](cols.VoteID),
		PseudonymID: psql.Where[Q, string](cols.PseudonymID),
		ContentType: psql.Where[Q, string](cols.ContentType),
		ContentID:   psql.Where[Q, int64](cols.ContentID),
		VoteValue:   psql.Where[Q, int32](cols.VoteValue),
		CreatedAt:   psql.WhereNull[Q, time.Time](cols.CreatedAt),
		UpdatedAt:   psql.WhereNull[Q, time.Time](cols.UpdatedAt),
	}
}

var VoteErrors = &voteErrors{
	ErrUniqueVotesPkey: &UniqueConstraintError{
		schema:  "",
		table:   "votes",
		columns: []string{"vote_id"},
		s:       "votes_pkey",
	},

	ErrUniqueVotesPseudonymIdContentTypeContentIdKey: &UniqueConstraintError{
		schema:  "",
		table:   "votes",
		columns: []string{"pseudonym_id", "content_type", "content_id"},
		s:       "votes_pseudonym_id_content_type_content_id_key",
	},
}

type voteErrors struct {
	ErrUniqueVotesPkey *UniqueConstraintError

	ErrUniqueVotesPseudonymIdContentTypeContentIdKey *UniqueConstraintError
}

// VoteSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type VoteSetter struct {
	VoteID      *int64               `db:"vote_id,pk" scan:"vote_id" json:"vote_id"`
	PseudonymID *string              `db:"pseudonym_id" scan:"pseudonym_id" json:"pseudonym_id"`
	ContentType *string              `db:"content_type" scan:"content_type" json:"content_type"`
	ContentID   *int64               `db:"content_id" scan:"content_id" json:"content_id"`
	VoteValue   *int32               `db:"vote_value" scan:"vote_value" json:"vote_value"`
	CreatedAt   *sql.Null[time.Time] `db:"created_at" scan:"created_at" json:"created_at"`
	UpdatedAt   *sql.Null[time.Time] `db:"updated_at" scan:"updated_at" json:"updated_at"`
}

func (s VoteSetter) SetColumns() []string {
	vals := make([]string, 0, 7)
	if s.VoteID != nil {
		vals = append(vals, "vote_id")
	}

	if s.PseudonymID != nil {
		vals = append(vals, "pseudonym_id")
	}

	if s.ContentType != nil {
		vals = append(vals, "content_type")
	}

	if s.ContentID != nil {
		vals = append(vals, "content_id")
	}

	if s.VoteValue != nil {
		vals = append(vals, "vote_value")
	}

	if s.CreatedAt != nil {
		vals = append(vals, "created_at")
	}

	if s.UpdatedAt != nil {
		vals = append(vals, "updated_at")
	}

	return vals
}

func (s VoteSetter) Overwrite(t *Vote) {
	if s.VoteID != nil {
		t.VoteID = *s.VoteID
	}
	if s.PseudonymID != nil {
		t.PseudonymID = *s.PseudonymID
	}
	if s.ContentType != nil {
		t.ContentType = *s.ContentType
	}
	if s.ContentID != nil {
		t.ContentID = *s.ContentID
	}
	if s.VoteValue != nil {
		t.VoteValue = *s.VoteValue
	}
	if s.CreatedAt != nil {
		t.CreatedAt = *s.CreatedAt
	}
	if s.UpdatedAt != nil {
		t.UpdatedAt = *s.UpdatedAt
	}
}

func (s *VoteSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return Votes.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 7)
		if s.VoteID != nil {
			vals[0] = psql.Arg(*s.VoteID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.PseudonymID != nil {
			vals[1] = psql.Arg(*s.PseudonymID)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.ContentType != nil {
			vals[2] = psql.Arg(*s.ContentType)
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.ContentID != nil {
			vals[3] = psql.Arg(*s.ContentID)
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		if s.VoteValue != nil {
			vals[4] = psql.Arg(*s.VoteValue)
		} else {
			vals[4] = psql.Raw("DEFAULT")
		}

		if s.CreatedAt != nil {
			vals[5] = psql.Arg(*s.CreatedAt)
		} else {
			vals[5] = psql.Raw("DEFAULT")
		}

		if s.UpdatedAt != nil {
			vals[6] = psql.Arg(*s.UpdatedAt)
		} else {
			vals[6] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s VoteSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s VoteSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 7)

	if s.VoteID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "vote_id")...),
			psql.Arg(s.VoteID),
		}})
	}

	if s.PseudonymID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "pseudonym_id")...),
			psql.Arg(s.PseudonymID),
		}})
	}

	if s.ContentType != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "content_type")...),
			psql.Arg(s.ContentType),
		}})
	}

	if s.ContentID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "content_id")...),
			psql.Arg(s.ContentID),
		}})
	}

	if s.VoteValue != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "vote_value")...),
			psql.Arg(s.VoteValue),
		}})
	}

	if s.CreatedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "created_at")...),
			psql.Arg(s.CreatedAt),
		}})
	}

	if s.UpdatedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "updated_at")...),
			psql.Arg(s.UpdatedAt),
		}})
	}

	return exprs
}

// FindVote retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindVote(ctx context.Context, exec bob.Executor, VoteIDPK int64, cols ...string) (*Vote, error) {
	if len(cols) == 0 {
		return Votes.Query(
			SelectWhere.Votes.VoteID.EQ(VoteIDPK),
		).One(ctx, exec)
	}

	return Votes.Query(
		SelectWhere.Votes.VoteID.EQ(VoteIDPK),
		sm.Columns(Votes.Columns().Only(cols...)),
	).One(ctx, exec)
}

// VoteExists checks the presence of a single record by primary key
func VoteExists(ctx context.Context, exec bob.Executor, VoteIDPK int64) (bool, error) {
	return Votes.Query(
		SelectWhere.Votes.VoteID.EQ(VoteIDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after Vote is retrieved from the database
func (o *Vote) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Votes.AfterSelectHooks.RunHooks(ctx, exec, VoteSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = Votes.AfterInsertHooks.RunHooks(ctx, exec, VoteSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = Votes.AfterUpdateHooks.RunHooks(ctx, exec, VoteSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = Votes.AfterDeleteHooks.RunHooks(ctx, exec, VoteSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the Vote
func (o *Vote) primaryKeyVals() bob.Expression {
	return psql.Arg(o.VoteID)
}

func (o *Vote) pkEQ() dialect.Expression {
	return psql.Quote("votes", "vote_id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the Vote
func (o *Vote) Update(ctx context.Context, exec bob.Executor, s *VoteSetter) error {
	v, err := Votes.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single Vote record with an executor
func (o *Vote) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := Votes.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the Vote using the executor
func (o *Vote) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := Votes.Query(
		SelectWhere.Votes.VoteID.EQ(o.VoteID),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after VoteSlice is retrieved from the database
func (o VoteSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Votes.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = Votes.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = Votes.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = Votes.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o VoteSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("votes", "vote_id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
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
func (o VoteSlice) copyMatchingRows(from ...*Vote) {
	for i, old := range o {
		for _, new := range from {
			if new.VoteID != old.VoteID {
				continue
			}
			new.R = old.R
			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o VoteSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Votes.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Vote:
				o.copyMatchingRows(retrieved)
			case []*Vote:
				o.copyMatchingRows(retrieved...)
			case VoteSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Vote or a slice of Vote
				// then run the AfterUpdateHooks on the slice
				_, err = Votes.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o VoteSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Votes.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Vote:
				o.copyMatchingRows(retrieved)
			case []*Vote:
				o.copyMatchingRows(retrieved...)
			case VoteSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Vote or a slice of Vote
				// then run the AfterDeleteHooks on the slice
				_, err = Votes.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o VoteSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals VoteSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Votes.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o VoteSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Votes.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o VoteSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := Votes.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

type voteJoins[Q dialect.Joinable] struct {
	typ       string
	Pseudonym modAs[Q, pseudonymColumns]
}

func (j voteJoins[Q]) aliasedAs(alias string) voteJoins[Q] {
	return buildVoteJoins[Q](buildVoteColumns(alias), j.typ)
}

func buildVoteJoins[Q dialect.Joinable](cols voteColumns, typ string) voteJoins[Q] {
	return voteJoins[Q]{
		typ: typ,
		Pseudonym: modAs[Q, pseudonymColumns]{
			c: PseudonymColumns,
			f: func(to pseudonymColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Pseudonyms.Name().As(to.Alias())).On(
						to.PseudonymID.EQ(cols.PseudonymID),
					))
				}

				return mods
			},
		},
	}
}

// Pseudonym starts a query for related objects on pseudonyms
func (o *Vote) Pseudonym(mods ...bob.Mod[*dialect.SelectQuery]) PseudonymsQuery {
	return Pseudonyms.Query(append(mods,
		sm.Where(PseudonymColumns.PseudonymID.EQ(psql.Arg(o.PseudonymID))),
	)...)
}

func (os VoteSlice) Pseudonym(mods ...bob.Mod[*dialect.SelectQuery]) PseudonymsQuery {
	pkPseudonymID := make(pgtypes.Array[string], len(os))
	for i, o := range os {
		pkPseudonymID[i] = o.PseudonymID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkPseudonymID), "character varying[]")),
	))

	return Pseudonyms.Query(append(mods,
		sm.Where(psql.Group(PseudonymColumns.PseudonymID).OP("IN", PKArgExpr)),
	)...)
}

func (o *Vote) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Pseudonym":
		rel, ok := retrieved.(*Pseudonym)
		if !ok {
			return fmt.Errorf("vote cannot load %T as %q", retrieved, name)
		}

		o.R.Pseudonym = rel

		if rel != nil {
			rel.R.Votes = VoteSlice{o}
		}
		return nil
	default:
		return fmt.Errorf("vote has no relationship %q", name)
	}
}

type votePreloader struct {
	Pseudonym func(...psql.PreloadOption) psql.Preloader
}

func buildVotePreloader() votePreloader {
	return votePreloader{
		Pseudonym: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*Pseudonym, PseudonymSlice](orm.Relationship{
				Name: "Pseudonym",
				Sides: []orm.RelSide{
					{
						From: TableNames.Votes,
						To:   TableNames.Pseudonyms,
						FromColumns: []string{
							ColumnNames.Votes.PseudonymID,
						},
						ToColumns: []string{
							ColumnNames.Pseudonyms.PseudonymID,
						},
					},
				},
			}, Pseudonyms.Columns().Names(), opts...)
		},
	}
}

type voteThenLoader[Q orm.Loadable] struct {
	Pseudonym func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildVoteThenLoader[Q orm.Loadable]() voteThenLoader[Q] {
	type PseudonymLoadInterface interface {
		LoadPseudonym(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return voteThenLoader[Q]{
		Pseudonym: thenLoadBuilder[Q](
			"Pseudonym",
			func(ctx context.Context, exec bob.Executor, retrieved PseudonymLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadPseudonym(ctx, exec, mods...)
			},
		),
	}
}

// LoadPseudonym loads the vote's Pseudonym into the .R struct
func (o *Vote) LoadPseudonym(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Pseudonym = nil

	related, err := o.Pseudonym(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	related.R.Votes = VoteSlice{o}

	o.R.Pseudonym = related
	return nil
}

// LoadPseudonym loads the vote's Pseudonym into the .R struct
func (os VoteSlice) LoadPseudonym(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	pseudonyms, err := os.Pseudonym(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		for _, rel := range pseudonyms {
			if o.PseudonymID != rel.PseudonymID {
				continue
			}

			rel.R.Votes = append(rel.R.Votes, o)

			o.R.Pseudonym = rel
			break
		}
	}

	return nil
}

func attachVotePseudonym0(ctx context.Context, exec bob.Executor, count int, vote0 *Vote, pseudonym1 *Pseudonym) (*Vote, error) {
	setter := &VoteSetter{
		PseudonymID: &pseudonym1.PseudonymID,
	}

	err := vote0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachVotePseudonym0: %w", err)
	}

	return vote0, nil
}

func (vote0 *Vote) InsertPseudonym(ctx context.Context, exec bob.Executor, related *PseudonymSetter) error {
	pseudonym1, err := Pseudonyms.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachVotePseudonym0(ctx, exec, 1, vote0, pseudonym1)
	if err != nil {
		return err
	}

	vote0.R.Pseudonym = pseudonym1

	pseudonym1.R.Votes = append(pseudonym1.R.Votes, vote0)

	return nil
}

func (vote0 *Vote) AttachPseudonym(ctx context.Context, exec bob.Executor, pseudonym1 *Pseudonym) error {
	var err error

	_, err = attachVotePseudonym0(ctx, exec, 1, vote0, pseudonym1)
	if err != nil {
		return err
	}

	vote0.R.Pseudonym = pseudonym1

	pseudonym1.R.Votes = append(pseudonym1.R.Votes, vote0)

	return nil
}
