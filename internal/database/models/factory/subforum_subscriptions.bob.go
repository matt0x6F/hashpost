// Code generated by HashPost Generated Code. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jaswdr/faker/v2"
	models "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

type SubforumSubscriptionMod interface {
	Apply(context.Context, *SubforumSubscriptionTemplate)
}

type SubforumSubscriptionModFunc func(context.Context, *SubforumSubscriptionTemplate)

func (f SubforumSubscriptionModFunc) Apply(ctx context.Context, n *SubforumSubscriptionTemplate) {
	f(ctx, n)
}

type SubforumSubscriptionModSlice []SubforumSubscriptionMod

func (mods SubforumSubscriptionModSlice) Apply(ctx context.Context, n *SubforumSubscriptionTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// SubforumSubscriptionTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type SubforumSubscriptionTemplate struct {
	SubscriptionID func() int64
	PseudonymID    func() string
	SubforumID     func() int32
	SubscribedAt   func() sql.Null[time.Time]
	IsFavorite     func() sql.Null[bool]

	r subforumSubscriptionR
	f *Factory
}

type subforumSubscriptionR struct {
	Pseudonym *subforumSubscriptionRPseudonymR
	Subforum  *subforumSubscriptionRSubforumR
}

type subforumSubscriptionRPseudonymR struct {
	o *PseudonymTemplate
}
type subforumSubscriptionRSubforumR struct {
	o *SubforumTemplate
}

// Apply mods to the SubforumSubscriptionTemplate
func (o *SubforumSubscriptionTemplate) Apply(ctx context.Context, mods ...SubforumSubscriptionMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.SubforumSubscription
// according to the relationships in the template. Nothing is inserted into the db
func (t SubforumSubscriptionTemplate) setModelRels(o *models.SubforumSubscription) {
	if t.r.Pseudonym != nil {
		rel := t.r.Pseudonym.o.Build()
		rel.R.SubforumSubscriptions = append(rel.R.SubforumSubscriptions, o)
		o.PseudonymID = rel.PseudonymID // h2
		o.R.Pseudonym = rel
	}

	if t.r.Subforum != nil {
		rel := t.r.Subforum.o.Build()
		rel.R.SubforumSubscriptions = append(rel.R.SubforumSubscriptions, o)
		o.SubforumID = rel.SubforumID // h2
		o.R.Subforum = rel
	}
}

// BuildSetter returns an *models.SubforumSubscriptionSetter
// this does nothing with the relationship templates
func (o SubforumSubscriptionTemplate) BuildSetter() *models.SubforumSubscriptionSetter {
	m := &models.SubforumSubscriptionSetter{}

	if o.SubscriptionID != nil {
		val := o.SubscriptionID()
		m.SubscriptionID = &val
	}
	if o.PseudonymID != nil {
		val := o.PseudonymID()
		m.PseudonymID = &val
	}
	if o.SubforumID != nil {
		val := o.SubforumID()
		m.SubforumID = &val
	}
	if o.SubscribedAt != nil {
		val := o.SubscribedAt()
		m.SubscribedAt = &val
	}
	if o.IsFavorite != nil {
		val := o.IsFavorite()
		m.IsFavorite = &val
	}

	return m
}

// BuildManySetter returns an []*models.SubforumSubscriptionSetter
// this does nothing with the relationship templates
func (o SubforumSubscriptionTemplate) BuildManySetter(number int) []*models.SubforumSubscriptionSetter {
	m := make([]*models.SubforumSubscriptionSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.SubforumSubscription
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use SubforumSubscriptionTemplate.Create
func (o SubforumSubscriptionTemplate) Build() *models.SubforumSubscription {
	m := &models.SubforumSubscription{}

	if o.SubscriptionID != nil {
		m.SubscriptionID = o.SubscriptionID()
	}
	if o.PseudonymID != nil {
		m.PseudonymID = o.PseudonymID()
	}
	if o.SubforumID != nil {
		m.SubforumID = o.SubforumID()
	}
	if o.SubscribedAt != nil {
		m.SubscribedAt = o.SubscribedAt()
	}
	if o.IsFavorite != nil {
		m.IsFavorite = o.IsFavorite()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.SubforumSubscriptionSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use SubforumSubscriptionTemplate.CreateMany
func (o SubforumSubscriptionTemplate) BuildMany(number int) models.SubforumSubscriptionSlice {
	m := make(models.SubforumSubscriptionSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableSubforumSubscription(m *models.SubforumSubscriptionSetter) {
	if m.PseudonymID == nil {
		val := random_string(nil, "64")
		m.PseudonymID = &val
	}
	if m.SubforumID == nil {
		val := random_int32(nil)
		m.SubforumID = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.SubforumSubscription
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *SubforumSubscriptionTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.SubforumSubscription) (context.Context, error) {
	var err error

	return ctx, err
}

// Create builds a subforumSubscription and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *SubforumSubscriptionTemplate) Create(ctx context.Context, exec bob.Executor) (*models.SubforumSubscription, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a subforumSubscription and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *SubforumSubscriptionTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.SubforumSubscription {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a subforumSubscription and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *SubforumSubscriptionTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.SubforumSubscription {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a subforumSubscription and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *SubforumSubscriptionTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.SubforumSubscription, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableSubforumSubscription(opt)

	if o.r.Pseudonym == nil {
		SubforumSubscriptionMods.WithNewPseudonym().Apply(ctx, o)
	}

	rel0, ok := pseudonymCtx.Value(ctx)
	if !ok {
		ctx, rel0, err = o.r.Pseudonym.o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	opt.PseudonymID = &rel0.PseudonymID

	if o.r.Subforum == nil {
		SubforumSubscriptionMods.WithNewSubforum().Apply(ctx, o)
	}

	rel1, ok := subforumCtx.Value(ctx)
	if !ok {
		ctx, rel1, err = o.r.Subforum.o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	opt.SubforumID = &rel1.SubforumID

	m, err := models.SubforumSubscriptions.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = subforumSubscriptionCtx.WithValue(ctx, m)

	m.R.Pseudonym = rel0
	m.R.Subforum = rel1

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple subforumSubscriptions and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o SubforumSubscriptionTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.SubforumSubscriptionSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple subforumSubscriptions and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o SubforumSubscriptionTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.SubforumSubscriptionSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple subforumSubscriptions and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o SubforumSubscriptionTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.SubforumSubscriptionSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple subforumSubscriptions and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o SubforumSubscriptionTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.SubforumSubscriptionSlice, error) {
	var err error
	m := make(models.SubforumSubscriptionSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// SubforumSubscription has methods that act as mods for the SubforumSubscriptionTemplate
var SubforumSubscriptionMods subforumSubscriptionMods

type subforumSubscriptionMods struct{}

func (m subforumSubscriptionMods) RandomizeAllColumns(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModSlice{
		SubforumSubscriptionMods.RandomSubscriptionID(f),
		SubforumSubscriptionMods.RandomPseudonymID(f),
		SubforumSubscriptionMods.RandomSubforumID(f),
		SubforumSubscriptionMods.RandomSubscribedAt(f),
		SubforumSubscriptionMods.RandomIsFavorite(f),
	}
}

// Set the model columns to this value
func (m subforumSubscriptionMods) SubscriptionID(val int64) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscriptionID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m subforumSubscriptionMods) SubscriptionIDFunc(f func() int64) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscriptionID = f
	})
}

// Clear any values for the column
func (m subforumSubscriptionMods) UnsetSubscriptionID() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscriptionID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m subforumSubscriptionMods) RandomSubscriptionID(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscriptionID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m subforumSubscriptionMods) PseudonymID(val string) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.PseudonymID = func() string { return val }
	})
}

// Set the Column from the function
func (m subforumSubscriptionMods) PseudonymIDFunc(f func() string) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.PseudonymID = f
	})
}

// Clear any values for the column
func (m subforumSubscriptionMods) UnsetPseudonymID() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.PseudonymID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m subforumSubscriptionMods) RandomPseudonymID(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.PseudonymID = func() string {
			return random_string(f, "64")
		}
	})
}

// Set the model columns to this value
func (m subforumSubscriptionMods) SubforumID(val int32) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubforumID = func() int32 { return val }
	})
}

// Set the Column from the function
func (m subforumSubscriptionMods) SubforumIDFunc(f func() int32) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubforumID = f
	})
}

// Clear any values for the column
func (m subforumSubscriptionMods) UnsetSubforumID() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubforumID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m subforumSubscriptionMods) RandomSubforumID(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubforumID = func() int32 {
			return random_int32(f)
		}
	})
}

// Set the model columns to this value
func (m subforumSubscriptionMods) SubscribedAt(val sql.Null[time.Time]) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscribedAt = func() sql.Null[time.Time] { return val }
	})
}

// Set the Column from the function
func (m subforumSubscriptionMods) SubscribedAtFunc(f func() sql.Null[time.Time]) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscribedAt = f
	})
}

// Clear any values for the column
func (m subforumSubscriptionMods) UnsetSubscribedAt() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscribedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m subforumSubscriptionMods) RandomSubscribedAt(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscribedAt = func() sql.Null[time.Time] {
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
func (m subforumSubscriptionMods) RandomSubscribedAtNotNull(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.SubscribedAt = func() sql.Null[time.Time] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_time_Time(f)
			return sql.Null[time.Time]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m subforumSubscriptionMods) IsFavorite(val sql.Null[bool]) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.IsFavorite = func() sql.Null[bool] { return val }
	})
}

// Set the Column from the function
func (m subforumSubscriptionMods) IsFavoriteFunc(f func() sql.Null[bool]) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.IsFavorite = f
	})
}

// Clear any values for the column
func (m subforumSubscriptionMods) UnsetIsFavorite() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.IsFavorite = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m subforumSubscriptionMods) RandomIsFavorite(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.IsFavorite = func() sql.Null[bool] {
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
func (m subforumSubscriptionMods) RandomIsFavoriteNotNull(f *faker.Faker) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(_ context.Context, o *SubforumSubscriptionTemplate) {
		o.IsFavorite = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: true}
		}
	})
}

func (m subforumSubscriptionMods) WithParentsCascading() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(ctx context.Context, o *SubforumSubscriptionTemplate) {
		if isDone, _ := subforumSubscriptionWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = subforumSubscriptionWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewPseudonym(ctx, PseudonymMods.WithParentsCascading())
			m.WithPseudonym(related).Apply(ctx, o)
		}
		{

			related := o.f.NewSubforum(ctx, SubforumMods.WithParentsCascading())
			m.WithSubforum(related).Apply(ctx, o)
		}
	})
}

func (m subforumSubscriptionMods) WithPseudonym(rel *PseudonymTemplate) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(ctx context.Context, o *SubforumSubscriptionTemplate) {
		o.r.Pseudonym = &subforumSubscriptionRPseudonymR{
			o: rel,
		}
	})
}

func (m subforumSubscriptionMods) WithNewPseudonym(mods ...PseudonymMod) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(ctx context.Context, o *SubforumSubscriptionTemplate) {
		related := o.f.NewPseudonym(ctx, mods...)

		m.WithPseudonym(related).Apply(ctx, o)
	})
}

func (m subforumSubscriptionMods) WithoutPseudonym() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(ctx context.Context, o *SubforumSubscriptionTemplate) {
		o.r.Pseudonym = nil
	})
}

func (m subforumSubscriptionMods) WithSubforum(rel *SubforumTemplate) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(ctx context.Context, o *SubforumSubscriptionTemplate) {
		o.r.Subforum = &subforumSubscriptionRSubforumR{
			o: rel,
		}
	})
}

func (m subforumSubscriptionMods) WithNewSubforum(mods ...SubforumMod) SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(ctx context.Context, o *SubforumSubscriptionTemplate) {
		related := o.f.NewSubforum(ctx, mods...)

		m.WithSubforum(related).Apply(ctx, o)
	})
}

func (m subforumSubscriptionMods) WithoutSubforum() SubforumSubscriptionMod {
	return SubforumSubscriptionModFunc(func(ctx context.Context, o *SubforumSubscriptionTemplate) {
		o.r.Subforum = nil
	})
}
