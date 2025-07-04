// Code generated by HashPost Generated Code. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/jaswdr/faker/v2"
	models "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
)

type PollVoteMod interface {
	Apply(context.Context, *PollVoteTemplate)
}

type PollVoteModFunc func(context.Context, *PollVoteTemplate)

func (f PollVoteModFunc) Apply(ctx context.Context, n *PollVoteTemplate) {
	f(ctx, n)
}

type PollVoteModSlice []PollVoteMod

func (mods PollVoteModSlice) Apply(ctx context.Context, n *PollVoteTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// PollVoteTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type PollVoteTemplate struct {
	VoteID          func() int64
	PollID          func() int64
	PseudonymID     func() string
	SelectedOptions func() types.JSON[json.RawMessage]
	CreatedAt       func() sql.Null[time.Time]

	r pollVoteR
	f *Factory
}

type pollVoteR struct {
	Poll      *pollVoteRPollR
	Pseudonym *pollVoteRPseudonymR
}

type pollVoteRPollR struct {
	o *PollTemplate
}
type pollVoteRPseudonymR struct {
	o *PseudonymTemplate
}

// Apply mods to the PollVoteTemplate
func (o *PollVoteTemplate) Apply(ctx context.Context, mods ...PollVoteMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.PollVote
// according to the relationships in the template. Nothing is inserted into the db
func (t PollVoteTemplate) setModelRels(o *models.PollVote) {
	if t.r.Poll != nil {
		rel := t.r.Poll.o.Build()
		rel.R.PollVotes = append(rel.R.PollVotes, o)
		o.PollID = rel.PollID // h2
		o.R.Poll = rel
	}

	if t.r.Pseudonym != nil {
		rel := t.r.Pseudonym.o.Build()
		rel.R.PollVotes = append(rel.R.PollVotes, o)
		o.PseudonymID = rel.PseudonymID // h2
		o.R.Pseudonym = rel
	}
}

// BuildSetter returns an *models.PollVoteSetter
// this does nothing with the relationship templates
func (o PollVoteTemplate) BuildSetter() *models.PollVoteSetter {
	m := &models.PollVoteSetter{}

	if o.VoteID != nil {
		val := o.VoteID()
		m.VoteID = &val
	}
	if o.PollID != nil {
		val := o.PollID()
		m.PollID = &val
	}
	if o.PseudonymID != nil {
		val := o.PseudonymID()
		m.PseudonymID = &val
	}
	if o.SelectedOptions != nil {
		val := o.SelectedOptions()
		m.SelectedOptions = &val
	}
	if o.CreatedAt != nil {
		val := o.CreatedAt()
		m.CreatedAt = &val
	}

	return m
}

// BuildManySetter returns an []*models.PollVoteSetter
// this does nothing with the relationship templates
func (o PollVoteTemplate) BuildManySetter(number int) []*models.PollVoteSetter {
	m := make([]*models.PollVoteSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.PollVote
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PollVoteTemplate.Create
func (o PollVoteTemplate) Build() *models.PollVote {
	m := &models.PollVote{}

	if o.VoteID != nil {
		m.VoteID = o.VoteID()
	}
	if o.PollID != nil {
		m.PollID = o.PollID()
	}
	if o.PseudonymID != nil {
		m.PseudonymID = o.PseudonymID()
	}
	if o.SelectedOptions != nil {
		m.SelectedOptions = o.SelectedOptions()
	}
	if o.CreatedAt != nil {
		m.CreatedAt = o.CreatedAt()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.PollVoteSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PollVoteTemplate.CreateMany
func (o PollVoteTemplate) BuildMany(number int) models.PollVoteSlice {
	m := make(models.PollVoteSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatablePollVote(m *models.PollVoteSetter) {
	if m.PollID == nil {
		val := random_int64(nil)
		m.PollID = &val
	}
	if m.PseudonymID == nil {
		val := random_string(nil, "64")
		m.PseudonymID = &val
	}
	if m.SelectedOptions == nil {
		val := random_types_JSON_json_RawMessage_(nil)
		m.SelectedOptions = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.PollVote
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *PollVoteTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.PollVote) (context.Context, error) {
	var err error

	return ctx, err
}

// Create builds a pollVote and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *PollVoteTemplate) Create(ctx context.Context, exec bob.Executor) (*models.PollVote, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a pollVote and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *PollVoteTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.PollVote {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a pollVote and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *PollVoteTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.PollVote {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a pollVote and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *PollVoteTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.PollVote, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatablePollVote(opt)

	if o.r.Poll == nil {
		PollVoteMods.WithNewPoll().Apply(ctx, o)
	}

	rel0, ok := pollCtx.Value(ctx)
	if !ok {
		ctx, rel0, err = o.r.Poll.o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	opt.PollID = &rel0.PollID

	if o.r.Pseudonym == nil {
		PollVoteMods.WithNewPseudonym().Apply(ctx, o)
	}

	rel1, ok := pseudonymCtx.Value(ctx)
	if !ok {
		ctx, rel1, err = o.r.Pseudonym.o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	opt.PseudonymID = &rel1.PseudonymID

	m, err := models.PollVotes.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = pollVoteCtx.WithValue(ctx, m)

	m.R.Poll = rel0
	m.R.Pseudonym = rel1

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple pollVotes and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o PollVoteTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.PollVoteSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple pollVotes and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o PollVoteTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.PollVoteSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple pollVotes and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o PollVoteTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.PollVoteSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple pollVotes and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o PollVoteTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.PollVoteSlice, error) {
	var err error
	m := make(models.PollVoteSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// PollVote has methods that act as mods for the PollVoteTemplate
var PollVoteMods pollVoteMods

type pollVoteMods struct{}

func (m pollVoteMods) RandomizeAllColumns(f *faker.Faker) PollVoteMod {
	return PollVoteModSlice{
		PollVoteMods.RandomVoteID(f),
		PollVoteMods.RandomPollID(f),
		PollVoteMods.RandomPseudonymID(f),
		PollVoteMods.RandomSelectedOptions(f),
		PollVoteMods.RandomCreatedAt(f),
	}
}

// Set the model columns to this value
func (m pollVoteMods) VoteID(val int64) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.VoteID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m pollVoteMods) VoteIDFunc(f func() int64) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.VoteID = f
	})
}

// Clear any values for the column
func (m pollVoteMods) UnsetVoteID() PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.VoteID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollVoteMods) RandomVoteID(f *faker.Faker) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.VoteID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m pollVoteMods) PollID(val int64) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PollID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m pollVoteMods) PollIDFunc(f func() int64) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PollID = f
	})
}

// Clear any values for the column
func (m pollVoteMods) UnsetPollID() PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PollID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollVoteMods) RandomPollID(f *faker.Faker) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PollID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m pollVoteMods) PseudonymID(val string) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PseudonymID = func() string { return val }
	})
}

// Set the Column from the function
func (m pollVoteMods) PseudonymIDFunc(f func() string) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PseudonymID = f
	})
}

// Clear any values for the column
func (m pollVoteMods) UnsetPseudonymID() PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PseudonymID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollVoteMods) RandomPseudonymID(f *faker.Faker) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.PseudonymID = func() string {
			return random_string(f, "64")
		}
	})
}

// Set the model columns to this value
func (m pollVoteMods) SelectedOptions(val types.JSON[json.RawMessage]) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.SelectedOptions = func() types.JSON[json.RawMessage] { return val }
	})
}

// Set the Column from the function
func (m pollVoteMods) SelectedOptionsFunc(f func() types.JSON[json.RawMessage]) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.SelectedOptions = f
	})
}

// Clear any values for the column
func (m pollVoteMods) UnsetSelectedOptions() PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.SelectedOptions = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollVoteMods) RandomSelectedOptions(f *faker.Faker) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.SelectedOptions = func() types.JSON[json.RawMessage] {
			return random_types_JSON_json_RawMessage_(f)
		}
	})
}

// Set the model columns to this value
func (m pollVoteMods) CreatedAt(val sql.Null[time.Time]) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] { return val }
	})
}

// Set the Column from the function
func (m pollVoteMods) CreatedAtFunc(f func() sql.Null[time.Time]) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m pollVoteMods) UnsetCreatedAt() PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m pollVoteMods) RandomCreatedAt(f *faker.Faker) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_time_Time(f)
			return sql.Null[time.Time]{V: val, Valid: f.Bool()}
		}
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is never null
func (m pollVoteMods) RandomCreatedAtNotNull(f *faker.Faker) PollVoteMod {
	return PollVoteModFunc(func(_ context.Context, o *PollVoteTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_time_Time(f)
			return sql.Null[time.Time]{V: val, Valid: true}
		}
	})
}

func (m pollVoteMods) WithParentsCascading() PollVoteMod {
	return PollVoteModFunc(func(ctx context.Context, o *PollVoteTemplate) {
		if isDone, _ := pollVoteWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = pollVoteWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewPoll(ctx, PollMods.WithParentsCascading())
			m.WithPoll(related).Apply(ctx, o)
		}
		{

			related := o.f.NewPseudonym(ctx, PseudonymMods.WithParentsCascading())
			m.WithPseudonym(related).Apply(ctx, o)
		}
	})
}

func (m pollVoteMods) WithPoll(rel *PollTemplate) PollVoteMod {
	return PollVoteModFunc(func(ctx context.Context, o *PollVoteTemplate) {
		o.r.Poll = &pollVoteRPollR{
			o: rel,
		}
	})
}

func (m pollVoteMods) WithNewPoll(mods ...PollMod) PollVoteMod {
	return PollVoteModFunc(func(ctx context.Context, o *PollVoteTemplate) {
		related := o.f.NewPoll(ctx, mods...)

		m.WithPoll(related).Apply(ctx, o)
	})
}

func (m pollVoteMods) WithoutPoll() PollVoteMod {
	return PollVoteModFunc(func(ctx context.Context, o *PollVoteTemplate) {
		o.r.Poll = nil
	})
}

func (m pollVoteMods) WithPseudonym(rel *PseudonymTemplate) PollVoteMod {
	return PollVoteModFunc(func(ctx context.Context, o *PollVoteTemplate) {
		o.r.Pseudonym = &pollVoteRPseudonymR{
			o: rel,
		}
	})
}

func (m pollVoteMods) WithNewPseudonym(mods ...PseudonymMod) PollVoteMod {
	return PollVoteModFunc(func(ctx context.Context, o *PollVoteTemplate) {
		related := o.f.NewPseudonym(ctx, mods...)

		m.WithPseudonym(related).Apply(ctx, o)
	})
}

func (m pollVoteMods) WithoutPseudonym() PollVoteMod {
	return PollVoteModFunc(func(ctx context.Context, o *PollVoteTemplate) {
		o.r.Pseudonym = nil
	})
}
