// Code generated by HashPost Generated Code. DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jaswdr/faker/v2"
	models "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
)

type RoleKeyMod interface {
	Apply(context.Context, *RoleKeyTemplate)
}

type RoleKeyModFunc func(context.Context, *RoleKeyTemplate)

func (f RoleKeyModFunc) Apply(ctx context.Context, n *RoleKeyTemplate) {
	f(ctx, n)
}

type RoleKeyModSlice []RoleKeyMod

func (mods RoleKeyModSlice) Apply(ctx context.Context, n *RoleKeyTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// RoleKeyTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type RoleKeyTemplate struct {
	KeyID        func() uuid.UUID
	RoleName     func() string
	Scope        func() string
	KeyData      func() []byte
	KeyVersion   func() int32
	Capabilities func() types.JSON[json.RawMessage]
	CreatedAt    func() sql.Null[time.Time]
	ExpiresAt    func() time.Time
	IsActive     func() sql.Null[bool]
	CreatedBy    func() int64

	r roleKeyR
	f *Factory
}

type roleKeyR struct {
	KeyKeyUsageAudits []*roleKeyRKeyKeyUsageAuditsR
	CreatedByUser     *roleKeyRCreatedByUserR
}

type roleKeyRKeyKeyUsageAuditsR struct {
	number int
	o      *KeyUsageAuditTemplate
}
type roleKeyRCreatedByUserR struct {
	o *UserTemplate
}

// Apply mods to the RoleKeyTemplate
func (o *RoleKeyTemplate) Apply(ctx context.Context, mods ...RoleKeyMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.RoleKey
// according to the relationships in the template. Nothing is inserted into the db
func (t RoleKeyTemplate) setModelRels(o *models.RoleKey) {
	if t.r.KeyKeyUsageAudits != nil {
		rel := models.KeyUsageAuditSlice{}
		for _, r := range t.r.KeyKeyUsageAudits {
			related := r.o.BuildMany(r.number)
			for _, rel := range related {
				rel.KeyID = o.KeyID // h2
				rel.R.KeyRoleKey = o
			}
			rel = append(rel, related...)
		}
		o.R.KeyKeyUsageAudits = rel
	}

	if t.r.CreatedByUser != nil {
		rel := t.r.CreatedByUser.o.Build()
		rel.R.CreatedByRoleKeys = append(rel.R.CreatedByRoleKeys, o)
		o.CreatedBy = rel.UserID // h2
		o.R.CreatedByUser = rel
	}
}

// BuildSetter returns an *models.RoleKeySetter
// this does nothing with the relationship templates
func (o RoleKeyTemplate) BuildSetter() *models.RoleKeySetter {
	m := &models.RoleKeySetter{}

	if o.KeyID != nil {
		val := o.KeyID()
		m.KeyID = &val
	}
	if o.RoleName != nil {
		val := o.RoleName()
		m.RoleName = &val
	}
	if o.Scope != nil {
		val := o.Scope()
		m.Scope = &val
	}
	if o.KeyData != nil {
		val := o.KeyData()
		m.KeyData = &val
	}
	if o.KeyVersion != nil {
		val := o.KeyVersion()
		m.KeyVersion = &val
	}
	if o.Capabilities != nil {
		val := o.Capabilities()
		m.Capabilities = &val
	}
	if o.CreatedAt != nil {
		val := o.CreatedAt()
		m.CreatedAt = &val
	}
	if o.ExpiresAt != nil {
		val := o.ExpiresAt()
		m.ExpiresAt = &val
	}
	if o.IsActive != nil {
		val := o.IsActive()
		m.IsActive = &val
	}
	if o.CreatedBy != nil {
		val := o.CreatedBy()
		m.CreatedBy = &val
	}

	return m
}

// BuildManySetter returns an []*models.RoleKeySetter
// this does nothing with the relationship templates
func (o RoleKeyTemplate) BuildManySetter(number int) []*models.RoleKeySetter {
	m := make([]*models.RoleKeySetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.RoleKey
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use RoleKeyTemplate.Create
func (o RoleKeyTemplate) Build() *models.RoleKey {
	m := &models.RoleKey{}

	if o.KeyID != nil {
		m.KeyID = o.KeyID()
	}
	if o.RoleName != nil {
		m.RoleName = o.RoleName()
	}
	if o.Scope != nil {
		m.Scope = o.Scope()
	}
	if o.KeyData != nil {
		m.KeyData = o.KeyData()
	}
	if o.KeyVersion != nil {
		m.KeyVersion = o.KeyVersion()
	}
	if o.Capabilities != nil {
		m.Capabilities = o.Capabilities()
	}
	if o.CreatedAt != nil {
		m.CreatedAt = o.CreatedAt()
	}
	if o.ExpiresAt != nil {
		m.ExpiresAt = o.ExpiresAt()
	}
	if o.IsActive != nil {
		m.IsActive = o.IsActive()
	}
	if o.CreatedBy != nil {
		m.CreatedBy = o.CreatedBy()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.RoleKeySlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use RoleKeyTemplate.CreateMany
func (o RoleKeyTemplate) BuildMany(number int) models.RoleKeySlice {
	m := make(models.RoleKeySlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableRoleKey(m *models.RoleKeySetter) {
	if m.RoleName == nil {
		val := random_string(nil, "100")
		m.RoleName = &val
	}
	if m.Scope == nil {
		val := random_string(nil, "100")
		m.Scope = &val
	}
	if m.KeyData == nil {
		val := random___byte(nil)
		m.KeyData = &val
	}
	if m.Capabilities == nil {
		val := random_types_JSON_json_RawMessage_(nil)
		m.Capabilities = &val
	}
	if m.ExpiresAt == nil {
		val := random_time_Time(nil)
		m.ExpiresAt = &val
	}
	if m.CreatedBy == nil {
		val := random_int64(nil)
		m.CreatedBy = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.RoleKey
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *RoleKeyTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.RoleKey) (context.Context, error) {
	var err error

	isKeyKeyUsageAuditsDone, _ := roleKeyRelKeyKeyUsageAuditsCtx.Value(ctx)
	if !isKeyKeyUsageAuditsDone && o.r.KeyKeyUsageAudits != nil {
		ctx = roleKeyRelKeyKeyUsageAuditsCtx.WithValue(ctx, true)
		for _, r := range o.r.KeyKeyUsageAudits {
			var rel0 models.KeyUsageAuditSlice
			ctx, rel0, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachKeyKeyUsageAudits(ctx, exec, rel0...)
			if err != nil {
				return ctx, err
			}
		}
	}

	return ctx, err
}

// Create builds a roleKey and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *RoleKeyTemplate) Create(ctx context.Context, exec bob.Executor) (*models.RoleKey, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a roleKey and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *RoleKeyTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.RoleKey {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a roleKey and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *RoleKeyTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.RoleKey {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a roleKey and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *RoleKeyTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.RoleKey, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableRoleKey(opt)

	if o.r.CreatedByUser == nil {
		RoleKeyMods.WithNewCreatedByUser().Apply(ctx, o)
	}

	rel1, ok := userCtx.Value(ctx)
	if !ok {
		ctx, rel1, err = o.r.CreatedByUser.o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	opt.CreatedBy = &rel1.UserID

	m, err := models.RoleKeys.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = roleKeyCtx.WithValue(ctx, m)

	m.R.CreatedByUser = rel1

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple roleKeys and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o RoleKeyTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.RoleKeySlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple roleKeys and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o RoleKeyTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.RoleKeySlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple roleKeys and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o RoleKeyTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.RoleKeySlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple roleKeys and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o RoleKeyTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.RoleKeySlice, error) {
	var err error
	m := make(models.RoleKeySlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// RoleKey has methods that act as mods for the RoleKeyTemplate
var RoleKeyMods roleKeyMods

type roleKeyMods struct{}

func (m roleKeyMods) RandomizeAllColumns(f *faker.Faker) RoleKeyMod {
	return RoleKeyModSlice{
		RoleKeyMods.RandomKeyID(f),
		RoleKeyMods.RandomRoleName(f),
		RoleKeyMods.RandomScope(f),
		RoleKeyMods.RandomKeyData(f),
		RoleKeyMods.RandomKeyVersion(f),
		RoleKeyMods.RandomCapabilities(f),
		RoleKeyMods.RandomCreatedAt(f),
		RoleKeyMods.RandomExpiresAt(f),
		RoleKeyMods.RandomIsActive(f),
		RoleKeyMods.RandomCreatedBy(f),
	}
}

// Set the model columns to this value
func (m roleKeyMods) KeyID(val uuid.UUID) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) KeyIDFunc(f func() uuid.UUID) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyID = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetKeyID() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomKeyID(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) RoleName(val string) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.RoleName = func() string { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) RoleNameFunc(f func() string) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.RoleName = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetRoleName() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.RoleName = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomRoleName(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.RoleName = func() string {
			return random_string(f, "100")
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) Scope(val string) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Scope = func() string { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) ScopeFunc(f func() string) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Scope = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetScope() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Scope = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomScope(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Scope = func() string {
			return random_string(f, "100")
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) KeyData(val []byte) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyData = func() []byte { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) KeyDataFunc(f func() []byte) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyData = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetKeyData() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyData = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomKeyData(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyData = func() []byte {
			return random___byte(f)
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) KeyVersion(val int32) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyVersion = func() int32 { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) KeyVersionFunc(f func() int32) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyVersion = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetKeyVersion() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyVersion = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomKeyVersion(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.KeyVersion = func() int32 {
			return random_int32(f)
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) Capabilities(val types.JSON[json.RawMessage]) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Capabilities = func() types.JSON[json.RawMessage] { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) CapabilitiesFunc(f func() types.JSON[json.RawMessage]) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Capabilities = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetCapabilities() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Capabilities = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomCapabilities(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.Capabilities = func() types.JSON[json.RawMessage] {
			return random_types_JSON_json_RawMessage_(f)
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) CreatedAt(val sql.Null[time.Time]) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.CreatedAt = func() sql.Null[time.Time] { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) CreatedAtFunc(f func() sql.Null[time.Time]) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetCreatedAt() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m roleKeyMods) RandomCreatedAt(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
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
func (m roleKeyMods) RandomCreatedAtNotNull(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
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
func (m roleKeyMods) ExpiresAt(val time.Time) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.ExpiresAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) ExpiresAtFunc(f func() time.Time) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.ExpiresAt = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetExpiresAt() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.ExpiresAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomExpiresAt(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.ExpiresAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) IsActive(val sql.Null[bool]) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.IsActive = func() sql.Null[bool] { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) IsActiveFunc(f func() sql.Null[bool]) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.IsActive = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetIsActive() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.IsActive = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m roleKeyMods) RandomIsActive(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.IsActive = func() sql.Null[bool] {
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
func (m roleKeyMods) RandomIsActiveNotNull(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.IsActive = func() sql.Null[bool] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_bool(f)
			return sql.Null[bool]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m roleKeyMods) CreatedBy(val int64) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.CreatedBy = func() int64 { return val }
	})
}

// Set the Column from the function
func (m roleKeyMods) CreatedByFunc(f func() int64) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.CreatedBy = f
	})
}

// Clear any values for the column
func (m roleKeyMods) UnsetCreatedBy() RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.CreatedBy = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m roleKeyMods) RandomCreatedBy(f *faker.Faker) RoleKeyMod {
	return RoleKeyModFunc(func(_ context.Context, o *RoleKeyTemplate) {
		o.CreatedBy = func() int64 {
			return random_int64(f)
		}
	})
}

func (m roleKeyMods) WithParentsCascading() RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		if isDone, _ := roleKeyWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = roleKeyWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewUser(ctx, UserMods.WithParentsCascading())
			m.WithCreatedByUser(related).Apply(ctx, o)
		}
	})
}

func (m roleKeyMods) WithCreatedByUser(rel *UserTemplate) RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		o.r.CreatedByUser = &roleKeyRCreatedByUserR{
			o: rel,
		}
	})
}

func (m roleKeyMods) WithNewCreatedByUser(mods ...UserMod) RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		related := o.f.NewUser(ctx, mods...)

		m.WithCreatedByUser(related).Apply(ctx, o)
	})
}

func (m roleKeyMods) WithoutCreatedByUser() RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		o.r.CreatedByUser = nil
	})
}

func (m roleKeyMods) WithKeyKeyUsageAudits(number int, related *KeyUsageAuditTemplate) RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		o.r.KeyKeyUsageAudits = []*roleKeyRKeyKeyUsageAuditsR{{
			number: number,
			o:      related,
		}}
	})
}

func (m roleKeyMods) WithNewKeyKeyUsageAudits(number int, mods ...KeyUsageAuditMod) RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		related := o.f.NewKeyUsageAudit(ctx, mods...)
		m.WithKeyKeyUsageAudits(number, related).Apply(ctx, o)
	})
}

func (m roleKeyMods) AddKeyKeyUsageAudits(number int, related *KeyUsageAuditTemplate) RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		o.r.KeyKeyUsageAudits = append(o.r.KeyKeyUsageAudits, &roleKeyRKeyKeyUsageAuditsR{
			number: number,
			o:      related,
		})
	})
}

func (m roleKeyMods) AddNewKeyKeyUsageAudits(number int, mods ...KeyUsageAuditMod) RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		related := o.f.NewKeyUsageAudit(ctx, mods...)
		m.AddKeyKeyUsageAudits(number, related).Apply(ctx, o)
	})
}

func (m roleKeyMods) WithoutKeyKeyUsageAudits() RoleKeyMod {
	return RoleKeyModFunc(func(ctx context.Context, o *RoleKeyTemplate) {
		o.r.KeyKeyUsageAudits = nil
	})
}
