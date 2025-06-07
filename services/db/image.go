package db

import (
	"strings"

	"github.com/azhai/allgo/dbutil"
)

// WallImage 壁纸图片
type WallImage struct {
	Id            int64 `json:"id" form:"id" db:"pk;type:int"`
	DailyId       int64 `json:"daily_id" form:"daily_id" db:"index;type:int"`
	ImageUrlMixin `json:",inline" db:"inline"`
	ImgSize       int64 `json:"img_size" form:"img_size" db:"index;type:int"`
	ImgOffset     int64 `json:"img_offset" form:"img_offset" db:"type:int"`
	ImageDimMixin `json:",inline" db:"inline"`
}

// TableName WallImage的表名
func (*WallImage) TableName() string {
	return "t_wall_image"
}

// TableComment WallImage的备注
func (*WallImage) TableComment() string {
	return "壁纸图片"
}

// ForeignIndex WallImage的外键的值
func (m *WallImage) ForeignIndex() any {
	return m.DailyId
}

// SecondaryKey WallImage的次要字段的值
func (m *WallImage) SecondaryKey() string {
	if strings.HasPrefix(m.FileName, "thumb/") {
		return "thumb"
	} else {
		return "image"
	}
}

// ScanFrom 从src中读取数据写入当前对象
func (m *WallImage) ScanFrom(src dbutil.ScanSource, err error) error {
	if err == nil {
		err = src.Scan(&m.Id, &m.DailyId, &m.FileName,
			&m.ImgMd5, &m.ImgSize, &m.ImgOffset, &m.ImgWidth, &m.ImgHeight)
	}
	return err
}

func (m *WallImage) UniqFields() ([]string, []any) {
	return []string{"id"}, []any{m.Id}
}

func (*WallImage) PrimaryKey() string {
	return "id"
}

func (m *WallImage) GetId() int64 {
	return m.Id
}

func (m *WallImage) SetId(int64, error) error {
	return nil
}

func (m *WallImage) InsertSQL() string {
	return "INSERT INTO " + m.TableName() + " (id, daily_id, file_name, img_md5, " +
		"img_size, img_offset, img_width, img_height) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
}

func (m *WallImage) UpsertSQL() string {
	return m.InsertSQL() + " ON CONFLICT (id) DO NOTHING"
}

func (m *WallImage) RowValues() []any {
	return []any{m.Id, m.DailyId, m.FileName, m.ImgMd5, m.ImgSize, m.ImgOffset, m.ImgWidth, m.ImgHeight}
}
