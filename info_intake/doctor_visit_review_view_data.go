package info_intake

type PhotoData struct {
	Title          string `json:"title"`
	PhotoID        int64  `json:"-"`
	PhotoUrl       string `json:"photo_url"`
	PlaceholderURL string `json:"placeholder_url"`
}

type TitlePhotoListData struct {
	Title  string      `json:"title"`
	Photos []PhotoData `json:"photos"`
}
type CheckedUncheckedData struct {
	Value     string `json:"value"`
	IsChecked bool   `json:"is_checked"`
}

type TitleSubItemsDescriptionContentData struct {
	Title    string                    `json:"title"`
	SubItems []*DescriptionContentData `json:"subitems"`
}

type DescriptionContentData struct {
	Description string `json:"description"`
	Content     string `json:"content"`
}
