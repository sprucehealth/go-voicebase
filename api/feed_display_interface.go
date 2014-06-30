package api

import (
	"time"

	"github.com/sprucehealth/backend/app_url"
)

const (
	DisplayTypeTitleSubtitleActionable = "title_subtitle_actionable"
)

type FeedDisplayInterface interface {
	GetId() int64
	GetTitleAndSubtitle(dataApi DataAPI) (title, subtitle string, err error)
	GetImageUrl() *app_url.SpruceAsset
	ActionUrl(dataApi DataAPI) (*app_url.SpruceAction, error)
	GetDisplayTypes() []string
	GetTimestamp() *time.Time
}
