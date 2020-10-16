package portentries

import (
	"context"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var updatableColumns = []string{
	"code", "name", "city", "province", "country", "alias", "regions", "latitude", "longitude", "timezone", "unlocks",
}

type PortsDB interface {
	List(ctx context.Context) ([]PortEntry, error)
	Upsert(ctx context.Context, port PortEntry) (PortEntry, error)
	BulkUpsert(ctx context.Context, ports []PortEntry) ([]PortEntry, error)
	Fetch(ctx context.Context, id *int64, slug *string) (PortEntry, error)
	Store(ctx context.Context, port *PortEntry) error
	Update(ctx context.Context, port *PortEntry) error
	Delete(ctx context.Context, id *int64, slug *string) error
}

// Datastore is the port entries data layer access object
type Datastore struct {
	db *gorm.DB
}

// NewDatastore returns a new port entries Datastore
func NewDatastore(db *gorm.DB) PortsDB {
	return &Datastore{db: db}
}

// PortEntry represents internal model for ports manipulation
type PortEntry struct {
	ID        int64  `gorm:"primaryKey"`
	Slug      string `gorm:"type:varchar(10)"`
	Code      string `gorm:"type:varchar(10)"`
	Name      string
	City      string
	Province  string
	Country   string
	Alias     pq.StringArray `gorm:"type:varchar(200)[]"`
	Regions   pq.StringArray `gorm:"type:varchar(200)[]"`
	Latitude  float64
	Longitude float64
	Timezone  string
	Unlocks   pq.StringArray `gorm:"type:varchar(10)[]"`
}

// TableName sets proper table name for db queries with GORM ORM
func (PortEntry) TableName() string {
	return "ports.port_entries"
}

// List fetches ports list from the db
func (d Datastore) List(ctx context.Context) ([]PortEntry, error) {
	var ports []PortEntry
	if res := d.db.WithContext(ctx).Model(&PortEntry{}).Find(&ports); res.Error != nil {
		return nil, res.Error
	}

	return ports, nil
}

// Upsert creates or updates port in the db
func (d Datastore) Upsert(ctx context.Context, port PortEntry) (PortEntry, error) {
	res := d.db.WithContext(ctx).Model(&PortEntry{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns(updatableColumns),
	}).Create(&port)
	if res.Error != nil {
		return PortEntry{}, res.Error
	}

	return port, nil
}

// BulkUpsert creates or updates ports in the db
func (d Datastore) BulkUpsert(ctx context.Context, ports []PortEntry) ([]PortEntry, error) {
	res := d.db.WithContext(ctx).Model(&PortEntry{}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns(updatableColumns),
	}).Create(&ports)
	if res.Error != nil {
		return []PortEntry{}, res.Error
	}

	return ports, nil
}

// Fetch fetches single port by id or slug
func (d Datastore) Fetch(ctx context.Context, id *int64, slug *string) (PortEntry, error) {
	port := PortEntry{}
	q := d.db.WithContext(ctx)

	if slug != nil {
		q = q.Where("slug = ?", *slug)
	} else {
		q = q.Where("id = ?", *id)
	}

	return port, q.Find(&port).Error
}

// Store creates new port in the DB
func (d Datastore) Store(ctx context.Context, port *PortEntry) error {
	return d.db.WithContext(ctx).Create(port).Error
}

// Update updates new port in the DB
func (d Datastore) Update(ctx context.Context, port *PortEntry) error {
	return d.db.WithContext(ctx).Save(port).Error
}

// Delete updates new port in the DB
func (d Datastore) Delete(ctx context.Context, id *int64, slug *string) error {
	q := d.db.WithContext(ctx)

	if id != nil {
		q = q.Where("id = ?", *id)
	}
	if slug != nil {
		q = q.Where("slug = ?", *slug)
	}

	return d.db.WithContext(ctx).Delete(&PortEntry{}).Error
}
