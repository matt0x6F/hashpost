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

type UserPreferenceMod interface {
	Apply(context.Context, *UserPreferenceTemplate)
}

type UserPreferenceModFunc func(context.Context, *UserPreferenceTemplate)

func (f UserPreferenceModFunc) Apply(ctx context.Context, n *UserPreferenceTemplate) {
	f(ctx, n)
}

type UserPreferenceModSlice []UserPreferenceMod

func (mods UserPreferenceModSlice) Apply(ctx context.Context, n *UserPreferenceTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// UserPreferenceTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type UserPreferenceTemplate struct {
	UserID             func() int64
	Timezone           func() sql.Null[string]
	Language           func() sql.Null[string]
	Theme              func() sql.Null[string]
	EmailNotifications func() sql.Null[bool]
	PushNotifications  func() sql.Null[bool]
	AutoHideNSFW       func() sql.Null[bool]
	AutoHideSpoilers   func() sql.Null[bool]
	CreatedAt          func() sql.Null[time.Time]
	UpdatedAt          func() sql.Null[time.Time]

	r userPreferenceR
	f *Factory
}

type userPreferenceR struct {
	User *userPreferenceRUserR
}

type userPreferenceRUserR struct {
	o *UserTemplate
}

// Apply mods to the UserPreferenceTemplate
func (o *UserPreferenceTemplate) Apply(ctx context.Context, mods ...UserPreferenceMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.UserPreference
// according to the relationships in the template. Nothing is inserted into the db
func (t UserPreferenceTemplate) setModelRels(o *models.UserPreference) {
	if t.r.User != nil {
		rel := t.r.User.o.Build()
		rel.R.UserPreference = o
		o.UserID = rel.UserID // h2
		o.R.User = rel
	}
}

// BuildSetter returns an *models.UserPreferenceSetter
// this does nothing with the relationship templates
func (o UserPreferenceTemplate) BuildSetter() *models.UserPreferenceSetter {
	m := &models.UserPreferenceSetter{}

	if o.UserID != nil {
		val := o.UserID()
		m.UserID = &val
	}
	if o.Timezone != nil {
		val := o.Timezone()
		m.Timezone = &val
	}
	if o.Language != nil {
		val := o.Language()
		m.Language = &val
	}
	if o.Theme != nil {
		val := o.Theme()
		m.Theme = &val
	}
	if o.EmailNotifications != nil {
		val := o.EmailNotifications()
		m.EmailNotifications = &val
	}
	if o.PushNotifications != nil {
		val := o.PushNotifications()
		m.PushNotifications = &val
	}
	if o.AutoHideNSFW != nil {
		val := o.AutoHideNSFW()
		m.AutoHideNSFW = &val
	}
	if o.AutoHideSpoilers != nil {
		val := o.AutoHideSpoilers()
		m.AutoHideSpoilers = &val
	}
	if o.CreatedAt != nil {
		val := o.CreatedAt()
		m.CreatedAt = &val
	}
	if o.UpdatedAt != nil {
		val := o.UpdatedAt()
		m.UpdatedAt = &val
	}

	return m
}

// BuildManySetter returns an []*models.UserPreferenceSetter
// this does nothing with the relationship templates
func (o UserPreferenceTemplate) BuildManySetter(number int) []*models.UserPreferenceSetter {
	m := make([]*models.UserPreferenceSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.UserPreference
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use UserPreferenceTemplate.Create
func (o UserPreferenceTemplate) Build() *models.UserPreference {
	m := &models.UserPreference{}

	if o.UserID != nil {
		m.UserID = o.UserID()
	}
	if o.Timezone != nil {
		m.Timezone = o.Timezone()
	}
	if o.Language != nil {
		m.Language = o.Language()
	}
	if o.Theme != nil {
		m.Theme = o.Theme()
	}
	if o.EmailNotifications != nil {
		m.EmailNotifications = o.EmailNotifications()
	}
	if o.PushNotifications != nil {
		m.PushNotifications = o.PushNotifications()
	}
	if o.AutoHideNSFW != nil {
		m.AutoHideNSFW = o.AutoHideNSFW()
	}
	if o.AutoHideSpoilers != nil {
		m.AutoHideSpoilers = o.AutoHideSpoilers()
	}
	if o.CreatedAt != nil {
		m.CreatedAt = o.CreatedAt()
	}
	if o.UpdatedAt != nil {
		m.UpdatedAt = o.UpdatedAt()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.UserPreferenceSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use UserPreferenceTemplate.CreateMany
func (o UserPreferenceTemplate) BuildMany(number int) models.UserPreferenceSlice {
	m := make(models.UserPreferenceSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableUserPreference(m *models.UserPreferenceSetter) {
	if m.UserID == nil {
		val := random_int64(nil)
		m.UserID = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.UserPreference
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *UserPreferenceTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.UserPreference) (context.Context, error) {
	var err error

	return ctx, err
}

// Create builds a userPreference and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *UserPreferenceTemplate) Create(ctx context.Context, exec bob.Executor) (*models.UserPreference, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a userPreference and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *UserPreferenceTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.UserPreference {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a userPreference and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *UserPreferenceTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.UserPreference {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a userPreference and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *UserPreferenceTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.UserPreference, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableUserPreference(opt)

	if o.r.User == nil {
		UserPreferenceMods.WithNewUser().Apply(ctx, o)
	}

	ctx, rel0, err := o.r.User.o.create(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}

	opt.UserID = &rel0.UserID

	m, err := models.UserPreferences.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = userPreferenceCtx.WithValue(ctx, m)

	m.R.User = rel0

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple userPreferences and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o UserPreferenceTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.UserPreferenceSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple userPreferences and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o UserPreferenceTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.UserPreferenceSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple userPreferences and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o UserPreferenceTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.UserPreferenceSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple userPreferences and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o UserPreferenceTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.UserPreferenceSlice, error) {
	var err error
	m := make(models.UserPreferenceSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// UserPreference has methods that act as mods for the UserPreferenceTemplate
var UserPreferenceMods userPreferenceMods

type userPreferenceMods struct{}

func (m userPreferenceMods) RandomizeAllColumns(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModSlice{
		UserPreferenceMods.RandomUserID(f),
		UserPreferenceMods.RandomTimezone(f),
		UserPreferenceMods.RandomLanguage(f),
		UserPreferenceMods.RandomTheme(f),
		UserPreferenceMods.RandomEmailNotifications(f),
		UserPreferenceMods.RandomPushNotifications(f),
		UserPreferenceMods.RandomAutoHideNSFW(f),
		UserPreferenceMods.RandomAutoHideSpoilers(f),
		UserPreferenceMods.RandomCreatedAt(f),
		UserPreferenceMods.RandomUpdatedAt(f),
	}
}

// Set the model columns to this value
func (m userPreferenceMods) UserID(val int64) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UserID = func() int64 { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) UserIDFunc(f func() int64) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UserID = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetUserID() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UserID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m userPreferenceMods) RandomUserID(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UserID = func() int64 {
			return random_int64(f)
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) Timezone(val sql.Null[string]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Timezone = func() sql.Null[string] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) TimezoneFunc(f func() sql.Null[string]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Timezone = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetTimezone() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Timezone = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomTimezone(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Timezone = func() sql.Null[string] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_string(f, "50")
			return sql.Null[string]{V: val, Valid: f.Bool()}
		}
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is never null
func (m userPreferenceMods) RandomTimezoneNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Timezone = func() sql.Null[string] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_string(f, "50")
			return sql.Null[string]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) Language(val sql.Null[string]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Language = func() sql.Null[string] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) LanguageFunc(f func() sql.Null[string]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Language = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetLanguage() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Language = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomLanguage(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Language = func() sql.Null[string] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_string(f, "10")
			return sql.Null[string]{V: val, Valid: f.Bool()}
		}
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is never null
func (m userPreferenceMods) RandomLanguageNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Language = func() sql.Null[string] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_string(f, "10")
			return sql.Null[string]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) Theme(val sql.Null[string]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Theme = func() sql.Null[string] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) ThemeFunc(f func() sql.Null[string]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Theme = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetTheme() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Theme = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomTheme(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Theme = func() sql.Null[string] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_string(f, "20")
			return sql.Null[string]{V: val, Valid: f.Bool()}
		}
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is never null
func (m userPreferenceMods) RandomThemeNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.Theme = func() sql.Null[string] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_string(f, "20")
			return sql.Null[string]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) EmailNotifications(val sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.EmailNotifications = func() sql.Null[bool] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) EmailNotificationsFunc(f func() sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.EmailNotifications = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetEmailNotifications() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.EmailNotifications = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomEmailNotifications(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.EmailNotifications = func() sql.Null[bool] {
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
func (m userPreferenceMods) RandomEmailNotificationsNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.EmailNotifications = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) PushNotifications(val sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.PushNotifications = func() sql.Null[bool] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) PushNotificationsFunc(f func() sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.PushNotifications = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetPushNotifications() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.PushNotifications = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomPushNotifications(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.PushNotifications = func() sql.Null[bool] {
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
func (m userPreferenceMods) RandomPushNotificationsNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.PushNotifications = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) AutoHideNSFW(val sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideNSFW = func() sql.Null[bool] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) AutoHideNSFWFunc(f func() sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideNSFW = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetAutoHideNSFW() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideNSFW = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomAutoHideNSFW(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideNSFW = func() sql.Null[bool] {
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
func (m userPreferenceMods) RandomAutoHideNSFWNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideNSFW = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) AutoHideSpoilers(val sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideSpoilers = func() sql.Null[bool] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) AutoHideSpoilersFunc(f func() sql.Null[bool]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideSpoilers = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetAutoHideSpoilers() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideSpoilers = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomAutoHideSpoilers(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideSpoilers = func() sql.Null[bool] {
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
func (m userPreferenceMods) RandomAutoHideSpoilersNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.AutoHideSpoilers = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) CreatedAt(val sql.Null[time.Time]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) CreatedAtFunc(f func() sql.Null[time.Time]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetCreatedAt() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomCreatedAt(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
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
func (m userPreferenceMods) RandomCreatedAtNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_time_Time(f)
			return sql.Null[time.Time]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m userPreferenceMods) UpdatedAt(val sql.Null[time.Time]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UpdatedAt = func() sql.Null[time.Time] { return val }
	})
}

// Set the Column from the function
func (m userPreferenceMods) UpdatedAtFunc(f func() sql.Null[time.Time]) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UpdatedAt = f
	})
}

// Clear any values for the column
func (m userPreferenceMods) UnsetUpdatedAt() UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UpdatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m userPreferenceMods) RandomUpdatedAt(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UpdatedAt = func() sql.Null[time.Time] {
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
func (m userPreferenceMods) RandomUpdatedAtNotNull(f *faker.Faker) UserPreferenceMod {
	return UserPreferenceModFunc(func(_ context.Context, o *UserPreferenceTemplate) {
		o.UpdatedAt = func() sql.Null[time.Time] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_time_Time(f)
			return sql.Null[time.Time]{V: val, Valid: true}
		}
	})
}

func (m userPreferenceMods) WithParentsCascading() UserPreferenceMod {
	return UserPreferenceModFunc(func(ctx context.Context, o *UserPreferenceTemplate) {
		if isDone, _ := userPreferenceWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = userPreferenceWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewUser(ctx, UserMods.WithParentsCascading())
			m.WithUser(related).Apply(ctx, o)
		}
	})
}

func (m userPreferenceMods) WithUser(rel *UserTemplate) UserPreferenceMod {
	return UserPreferenceModFunc(func(ctx context.Context, o *UserPreferenceTemplate) {
		o.r.User = &userPreferenceRUserR{
			o: rel,
		}
	})
}

func (m userPreferenceMods) WithNewUser(mods ...UserMod) UserPreferenceMod {
	return UserPreferenceModFunc(func(ctx context.Context, o *UserPreferenceTemplate) {
		related := o.f.NewUser(ctx, mods...)

		m.WithUser(related).Apply(ctx, o)
	})
}

func (m userPreferenceMods) WithoutUser() UserPreferenceMod {
	return UserPreferenceModFunc(func(ctx context.Context, o *UserPreferenceTemplate) {
		o.r.User = nil
	})
}
