package db

import (
	"fmt"
)

// WallDailyForeign 关联查询字段
type WallDailyForeign struct {
	ImageUrl string               `json:"image_url" form:"image_url" db:"-"`
	ThumbUrl string               `json:"thumb_url" form:"thumb_url" db:"-"`
	Thumb    *WallImage           `json:"thumb" form:"thumb" db:"-"`
	Image    *WallImage           `json:"image" form:"image" db:"-"`
	Notes    map[string]*WallNote `json:"notes" form:"notes" db:"-"`
}

// ImageUrlMixin 图片URL
type ImageUrlMixin struct {
	FileName string `json:"file_name" form:"file_name" db:"type:varchar(100)"`
	ImgMd5   string `json:"img_md5" form:"img_md5" db:"index;type:char(32)"`
}

// GetUrl 获取图片的URL地址
// 若图片的MD5码不为空，则取后8位作为版本号
func (m *ImageUrlMixin) GetUrl() string {
	url := m.FileName
	if len(m.ImgMd5) > 24 {
		url += "?v=" + m.ImgMd5[24:]
	}
	return url
}

// ImageDimMixin 图片宽高
type ImageDimMixin struct {
	ImgWidth  int `json:"img_width" form:"img_width" db:"type:int"`
	ImgHeight int `json:"img_height" form:"img_height" db:"type:int"`
}

// GeDims 获取图片的尺寸
func (m *ImageDimMixin) GeDims() string {
	return fmt.Sprintf("%dx%d", m.ImgWidth, m.ImgHeight)
}
