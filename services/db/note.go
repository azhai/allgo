package db

import (
	"database/sql"

	"github.com/azhai/allgo/dbutil"
)

// WallNote 壁纸小知识
type WallNote struct {
	Id          int64            `json:"id" form:"id" db:"pk;serial;type:int"`
	DailyId     int64            `json:"daily_id" form:"daily_id" db:"index;type:int"`
	NoteType    string           `json:"note_type" form:"note_type" db:"type:varchar(50)"`
	NoteChinese sql.Null[string] `json:"note_chinese" form:"note_chinese" db:"type:text"`
	NoteEnglish sql.Null[string] `json:"note_english" form:"note_english" db:"type:text"`
}

// TableName WallNote的表名
func (*WallNote) TableName() string {
	return "t_wall_note"
}

// TableComment WallNote的备注
func (*WallNote) TableComment() string {
	return "壁纸小知识"
}

// ForeignIndex WallNote的外键的值
func (m *WallNote) ForeignIndex() any {
	return m.DailyId
}

// SecondaryKey WallNote的次要字段的值
func (m *WallNote) SecondaryKey() string {
	return m.NoteType
}

// ScanFrom 从src中读取数据写入当前对象
func (m *WallNote) ScanFrom(src dbutil.ScanSource, err error) error {
	if err == nil {
		err = src.Scan(&m.Id, &m.DailyId, &m.NoteType,
			&m.NoteChinese, &m.NoteEnglish)
	}
	return err
}

func (m *WallNote) UniqFields() ([]string, []any) {
	return []string{"daily_id", "note_type"}, []any{m.DailyId, m.NoteType}
}

func (m *WallNote) GetPrimary() (string, int64) {
	return "id", m.Id
}

func (*WallNote) PrimaryKey() string {
	return "id"
}

func (m *WallNote) GetId() int64 {
	return m.Id
}

func (m *WallNote) SetId(id int64, err error) error {
	if err == nil && id > 0 {
		m.Id = id
	}
	return err
}

func (m *WallNote) baseInsertSQL() string {
	return "INSERT INTO " + m.TableName() + " (daily_id, note_type, " +
		"note_chinese, note_english) VALUES ($1, $2, $3, $4)"
}

func (m *WallNote) InsertSQL() string {
	return m.baseInsertSQL() + " RETURNING id"
}

func (m *WallNote) UpsertSQL() string {
	return m.baseInsertSQL() + " ON CONFLICT (daily_id, note_type) DO NOTHING"
}

func (m *WallNote) RowValues() []any {
	return []any{m.DailyId, m.NoteType, m.NoteChinese, m.NoteEnglish}
}
