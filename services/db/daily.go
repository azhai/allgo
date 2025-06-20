package db

import (
	"time"

	"github.com/azhai/allgo/dbutil"
)

// WallDaily 每日壁纸
type WallDaily struct {
	Id               int64     `json:"id" form:"id" db:"pk;type:int"`
	Guid             string    `json:"guid" form:"guid" db:"type:char(10)"`
	BingDate         time.Time `json:"bing_date" form:"bing_date" db:"unique;type:date"`
	BingSku          string    `json:"bing_sku" form:"bing_sku" db:"index;type:varchar(100)"`
	Title            string    `json:"title" form:"title" db:"index;type:varchar(255)"`
	Headline         string    `json:"headline" form:"headline" db:"type:varchar(255)"`
	Color            string    `json:"color" form:"color" db:"type:varchar(15)"`
	MaxDpi           string    `json:"max_dpi" form:"max_dpi" db:"type:varchar(15)"`
	WallDailyForeign `json:",inline" db:"-"`
}

// TableName WallDaily的表名
func (*WallDaily) TableName() string {
	return "t_wall_daily"
}

// TableComment WallDaily的备注
func (*WallDaily) TableComment() string {
	return "每日壁纸"
}

func (*WallDaily) PrimaryKey() string {
	return "id"
}

func (m *WallDaily) GetId() int64 {
	return m.Id
}

func (m *WallDaily) SetId(int64, error) error {
	return nil
}

// ForeignIndex WallDaily的外键的值
func (m *WallDaily) ForeignIndex() any {
	return m.BingDate.Format("2006-01-02")
}

func (m *WallDaily) UniqFields() ([]string, []any) {
	return []string{"bing_date"}, []any{m.BingDate.Format("2006-01-02")}
}

// ScanFrom 从src中读取数据写入当前对象
func (m *WallDaily) ScanFrom(src dbutil.ScanSource, err error) error {
	if err == nil {
		err = src.Scan(&m.Id, &m.Guid, &m.BingDate,
			&m.BingSku, &m.Title, &m.Headline, &m.Color, &m.MaxDpi)
	}
	return err
}

func (m *WallDaily) InsertSQL() string {
	return "INSERT INTO " + m.TableName() + " (id, guid, bing_date, bing_sku, " +
		"title, headline, color, max_dpi) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
}

func (m *WallDaily) UpsertSQL() string {
	return m.InsertSQL() + " ON CONFLICT (bing_date) DO NOTHING"
}

func (m *WallDaily) RowValues() []any {
	return []any{m.Id, m.Guid, m.BingDate, m.BingSku, m.Title, m.Headline, m.Color, m.MaxDpi}
}

// WallDailyList 每日壁纸列表
type WallDailyList []*WallDaily

// GetIds 获取所有的主键ID
func (ms WallDailyList) GetIds() (ids []any) {
	for _, row := range ms {
		ids = append(ids, row.Id)
	}
	return
}

// GetDates 获取所有的主键ID
func (ms WallDailyList) GetDates() (dates []any) {
	for _, row := range ms {
		bingDate := row.BingDate.Format("2006-01-02")
		dates = append(dates, bingDate)
	}
	return
}
