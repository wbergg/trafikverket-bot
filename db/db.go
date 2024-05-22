package db

import (
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DBobject struct {
	db    *sqlx.DB
	Setup int
}

type DBTrainMessage struct {
	EventID                                string         `db:"event_id"`
	CountyNo                               sql.NullString `db:"county_no"`
	Deleted                                bool           `db:"deleted"`
	ExternalDescription                    sql.NullString `db:"external_description"`
	GeometrySweref99Tm                     sql.NullString `db:"geometry_sweref99_tm"`
	GeometryWgs84                          sql.NullString `db:"geometry_wgs84"`
	Header                                 sql.NullString `db:"header"`
	StartDateTime                          time.Time      `db:"start_date_time"`
	PrognosticatedEndDateTimeTrafficImpact time.Time      `db:"prognosticated_end_date_time_traffic_impact"`
	LastUpdateDateTime                     time.Time      `db:"last_update_date_time"`
	ModifiedTime                           time.Time      `db:"modified_time"`
}

func Open() DBobject {
	db, err := sqlx.Connect("sqlite3", "./db/db.sql")
	if err != nil {
		panic(err)
	}

	setup := initDB(db)

	return DBobject{
		db:    db,
		Setup: setup,
	}
}

func initDB(db *sqlx.DB) int {

	// Check if table exists
	dbQuery1 := `SELECT COUNT(name) as P FROM sqlite_master WHERE type='table' AND name='train_messages';`

	var s1 int
	err := db.Get(&s1, dbQuery1)
	if err != nil {
		panic(err)
	}

	// If tables doesn't exist, create it from the schema
	if s1 == 0 {
		dat, err := os.ReadFile("db/db.schema")
		if err != nil {
			panic(err)
		}
		db.MustExec(string(dat))
		//fmt.Println(dat)
	}
	return s1
}

func (d *DBobject) Close() {
	d.db.Close()
}

func (d *DBobject) GetAllItems() ([]DBTrainMessage, error) {
	databaseResp := []DBTrainMessage{}

	dbQuery := "SELECT * FROM train_messages"

	err := d.db.Select(&databaseResp, dbQuery)
	if err != nil {
		panic(err)
	}

	return databaseResp, nil
}

func (d *DBobject) GetRowItemByPid(pid int64) error {

	var result string

	dbQuery := "SELECT * FROM train_messages WHERE id=?"

	row := d.db.QueryRow(dbQuery, pid)
	err := row.Scan(&result)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	return err
}

func (d *DBobject) InsertDBTrainMessage(EventID string, CountyNo int, Deleted bool, ExternalDescription string, GeometrySweref99Tm string, GeometryWgs84 string,
	Header string, StartDateTime time.Time, PrognosticatedEndDateTimeTrafficImpact time.Time, LastUpdateDateTime time.Time, ModifiedTime time.Time) error {
	dbQuery := `
		INSERT INTO train_messages (
			event_id, county_no, deleted, external_description, 
			geometry_sweref99_tm, geometry_wgs84, header, start_date_time, 
			prognosticated_end_date_time_traffic_impact, last_update_date_time, 
			modified_time
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	_, err := d.db.Exec(dbQuery,
		EventID,
		CountyNo,
		Deleted,
		ExternalDescription,
		GeometrySweref99Tm,
		GeometryWgs84,
		Header,
		StartDateTime,
		PrognosticatedEndDateTimeTrafficImpact,
		LastUpdateDateTime,
		ModifiedTime)

	if err != nil {
		panic(err)
	}
	return err
}

func (d *DBobject) GetMessagesByPid(pid string) (DBTrainMessage, error) {

	databaseItem := DBTrainMessage{}

	dbQuery := `SELECT * FROM train_messages WHERE event_id = $1`
	err := d.db.Get(&databaseItem, dbQuery, pid)
	return databaseItem, err
}

func (d *DBobject) UpdateChangeByPid(column string, value string, pid int64) error {

	dbQuery := `UPDATE train_messages SET ` + strings.ToLower(column) + ` = $1 WHERE id = $2`

	_, err := d.db.Exec(dbQuery, value, pid)

	return err

}
