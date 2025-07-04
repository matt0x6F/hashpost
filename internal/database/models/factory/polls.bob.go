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

type PollMod interface {
	Apply(context.Context, *PollTemplate)
}

type PollModFunc func(context.Context, *PollTemplate)

func (f PollModFunc) Apply(ctx context.Context, n *PollTemplate) {
	f(ctx, n)
}

type PollModSlice []PollMod

func (mods PollModSlice) Apply(ctx context.Context, n *PollTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// PollTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type PollTemplate struct {
	PollID             func() int64
	PostID             func() int64
	Question           func() string
	Options            func() types.JSON[json.RawMessage]
	AllowMultipleVotes func() sql.Null[bool]
	ExpiresAt          func() sql.Null[time.Time]
	CreatedAt          func() sql.Null[time.Time]

	r pollR
	f *Factory
}

type pollR struct {
	PollVotes []*pollRPollVotesR
	Post      *pollRPostR
}

type pollRPollVotesR struct {
	number int
	o      *PollVoteTemplate
}
type pollRPostR struct {
	o *PostTemplate
}

// Apply mods to the PollTemplate
func (o *PollTemplate) Apply(ctx context.Context, mods ...PollMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.Poll
// according to the relationships in the template. Nothing is inserted into the db
func (t PollTemplate) setModelRels(o *models.Poll) {
	if t.r.PollVotes != nil {
		rel := models.PollVoteSlice{}
		for _, r := range t.r.PollVotes {
			related := r.o.BuildMany(r.number)
			for _, rel := range related {
				rel.PollID = o.PollID // h2
				rel.R.Poll = o
			}
			rel = append(rel, related...)
		}
		o.R.PollVotes = rel
	}

	if t.r.Post != nil {
		rel := t.r.Post.o.Build()
		rel.R.Poll = o
		o.PostID = rel.PostID // h2
		o.R.Post = rel
	}
}

// BuildSetter returns an *models.PollSetter
// this does nothing with the relationship templates
func (o PollTemplate) BuildSetter() *models.PollSetter {
	m := &models.PollSetter{}

	if o.PollID != nil {
		val := o.PollID()
		m.PollID = &val
	}
	if o.PostID != nil {
		val := o.PostID()
		m.PostID = &val
	}
	if o.Question != nil {
		val := o.Question()
		m.Question = &val
	}
	if o.Options != nil {
		val := o.Options()
		m.Options = &val
	}
	if o.AllowMultipleVotes != nil {
		val := o.AllowMultipleVotes()
		m.AllowMultipleVotes = &val
	}
	if o.ExpiresAt != nil {
		val := o.ExpiresAt()
		m.ExpiresAt = &val
	}
	if o.CreatedAt != nil {
		val := o.CreatedAt()
		m.CreatedAt = &val
	}

	return m
}

// BuildManySetter returns an []*models.PollSetter
// this does nothing with the relationship templates
func (o PollTemplate) BuildManySetter(number int) []*models.PollSetter {
	m := make([]*models.PollSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.Poll
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PollTemplate.Create
func (o PollTemplate) Build() *models.Poll {
	m := &models.Poll{}

	if o.PollID != nil {
		m.PollID = o.PollID()
	}
	if o.PostID != nil {
		m.PostID = o.PostID()
	}
	if o.Question != nil {
		m.Question = o.Question()
	}
	if o.Options != nil {
		m.Options = o.Options()
	}
	if o.AllowMultipleVotes != nil {
		m.AllowMultipleVotes = o.AllowMultipleVotes()
	}
	if o.ExpiresAt != nil {
		m.ExpiresAt = o.ExpiresAt()
	}
	if o.CreatedAt != nil {
		m.CreatedAt = o.CreatedAt()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.PollSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PollTemplate.CreateMany
func (o PollTemplate) BuildMany(number int) models.PollSlice {
	m := make(models.PollSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatablePoll(m *models.PollSetter) {
	if m.PostID == nil {
		val := random_int64(nil)
		m.PostID = &val
	}
	if m.Question == nil {
		val := random_string(nil)
		m.Question = &val
	}
	if m.Options == nil {
		val := random_types_JSON_json_RawMessage_(nil)
		m.Options = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.Poll
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *PollTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.Poll) (context.Context, error) {
	var err error

	isPollVotesDone, _ := pollRelPollVotesCtx.Value(ctx)
	if !isPollVotesDone && o.r.PollVotes != nil {
		ctx = pollRelPollVotesCtx.WithValue(ctx, true)
		for _, r := range o.r.PollVotes {
			var rel0 models.PollVoteSlice
			ctx, rel0, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachPollVotes(ctx, exec, rel0...)
			if err != nil {
				return ctx, err
			}
		}
	}

	return ctx, err
}

// Create builds a poll and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *PollTemplate) Create(ctx context.Context, exec bob.Executor) (*models.Poll, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a poll and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *PollTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.Poll {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a poll and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *PollTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.Poll {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a poll and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *PollTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.Poll, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatablePoll(opt)

	if o.r.Post == nil {
		PollMods.WithNewPost().Apply(ctx, o)
	}

	ctx, rel1, err := o.r.Post.o.create(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}

	opt.PostID = &rel1.PostID

	m, err := models.Polls.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = pollCtx.WithValue(ctx, m)

	m.R.Post = rel1

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple polls and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o PollTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.PollSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple polls and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o PollTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.PollSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple polls and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o PollTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.PollSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple polls and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o PollTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.PollSlice, error) {
	var err error
	m := make(models.PollSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// Poll has methods that act as mods for the PollTemplate
var PollMods pollMods

type pollMods struct{}

func (m pollMods) RandomizeAllColumns(f *faker.Faker) PollMod {
	return PollModSlice{
		PollMods.RandomPollID(f),
		PollMods.RandomPostID(f),
		PollMods.RandomQuestion(f),
		PollMods.RandomOptions(f),
		PollMods.RandomAllowMultipleVotes(f),
		PollMods.RandomExpiresAt(f),
		PollMods.RandomCreatedAt(f),
	}
}

// Set the model columns to this value
func (m pollMods) PollID(val int64) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PollID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m pollMods) PollIDFunc(f func() int64) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PollID = f
	})
}

// Clear any values for the column
func (m pollMods) UnsetPollID() PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PollID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollMods) RandomPollID(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PollID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m pollMods) PostID(val int64) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PostID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m pollMods) PostIDFunc(f func() int64) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PostID = f
	})
}

// Clear any values for the column
func (m pollMods) UnsetPostID() PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PostID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollMods) RandomPostID(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.PostID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m pollMods) Question(val string) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Question = func() string { return val }
	})
}

// Set the Column from the function
func (m pollMods) QuestionFunc(f func() string) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Question = f
	})
}

// Clear any values for the column
func (m pollMods) UnsetQuestion() PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Question = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollMods) RandomQuestion(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Question = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m pollMods) Options(val types.JSON[json.RawMessage]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Options = func() types.JSON[json.RawMessage] { return val }
	})
}

// Set the Column from the function
func (m pollMods) OptionsFunc(f func() types.JSON[json.RawMessage]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Options = f
	})
}

// Clear any values for the column
func (m pollMods) UnsetOptions() PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Options = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m pollMods) RandomOptions(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.Options = func() types.JSON[json.RawMessage] {
			return random_types_JSON_json_RawMessage_(f)
		}
	})
}

// Set the model columns to this value
func (m pollMods) AllowMultipleVotes(val sql.Null[bool]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.AllowMultipleVotes = func() sql.Null[bool] { return val }
	})
}

// Set the Column from the function
func (m pollMods) AllowMultipleVotesFunc(f func() sql.Null[bool]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.AllowMultipleVotes = f
	})
}

// Clear any values for the column
func (m pollMods) UnsetAllowMultipleVotes() PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.AllowMultipleVotes = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m pollMods) RandomAllowMultipleVotes(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.AllowMultipleVotes = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: f.Bool()}
		}
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is never null
func (m pollMods) RandomAllowMultipleVotesNotNull(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.AllowMultipleVotes = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m pollMods) ExpiresAt(val sql.Null[time.Time]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.ExpiresAt = func() sql.Null[time.Time] { return val }
	})
}

// Set the Column from the function
func (m pollMods) ExpiresAtFunc(f func() sql.Null[time.Time]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.ExpiresAt = f
	})
}

// Clear any values for the column
func (m pollMods) UnsetExpiresAt() PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.ExpiresAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m pollMods) RandomExpiresAt(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.ExpiresAt = func() sql.Null[time.Time] {
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
func (m pollMods) RandomExpiresAtNotNull(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.ExpiresAt = func() sql.Null[time.Time] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_time_Time(f)
			return sql.Null[time.Time]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m pollMods) CreatedAt(val sql.Null[time.Time]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] { return val }
	})
}

// Set the Column from the function
func (m pollMods) CreatedAtFunc(f func() sql.Null[time.Time]) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m pollMods) UnsetCreatedAt() PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m pollMods) RandomCreatedAt(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
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
func (m pollMods) RandomCreatedAtNotNull(f *faker.Faker) PollMod {
	return PollModFunc(func(_ context.Context, o *PollTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_time_Time(f)
			return sql.Null[time.Time]{V: val, Valid: true}
		}
	})
}

func (m pollMods) WithParentsCascading() PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		if isDone, _ := pollWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = pollWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewPost(ctx, PostMods.WithParentsCascading())
			m.WithPost(related).Apply(ctx, o)
		}
	})
}

func (m pollMods) WithPost(rel *PostTemplate) PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		o.r.Post = &pollRPostR{
			o: rel,
		}
	})
}

func (m pollMods) WithNewPost(mods ...PostMod) PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		related := o.f.NewPost(ctx, mods...)

		m.WithPost(related).Apply(ctx, o)
	})
}

func (m pollMods) WithoutPost() PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		o.r.Post = nil
	})
}

func (m pollMods) WithPollVotes(number int, related *PollVoteTemplate) PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		o.r.PollVotes = []*pollRPollVotesR{{
			number: number,
			o:      related,
		}}
	})
}

func (m pollMods) WithNewPollVotes(number int, mods ...PollVoteMod) PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		related := o.f.NewPollVote(ctx, mods...)
		m.WithPollVotes(number, related).Apply(ctx, o)
	})
}

func (m pollMods) AddPollVotes(number int, related *PollVoteTemplate) PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		o.r.PollVotes = append(o.r.PollVotes, &pollRPollVotesR{
			number: number,
			o:      related,
		})
	})
}

func (m pollMods) AddNewPollVotes(number int, mods ...PollVoteMod) PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		related := o.f.NewPollVote(ctx, mods...)
		m.AddPollVotes(number, related).Apply(ctx, o)
	})
}

func (m pollMods) WithoutPollVotes() PollMod {
	return PollModFunc(func(ctx context.Context, o *PollTemplate) {
		o.r.PollVotes = nil
	})
}
