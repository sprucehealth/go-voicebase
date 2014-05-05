package api

import (
	"carefront/common"
	"database/sql"
)

func (d *DataService) AddPhoto(uploaderId int64, url, mimetype string) (int64, error) {
	res, err := d.DB.Exec(`
		INSERT INTO photo (uploader_id, url, mimetype) VALUES (?, ?, ?)`,
		uploaderId, url, mimetype)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DataService) GetPhoto(photoId int64) (*common.Photo, error) {
	photo := &common.Photo{
		Id: photoId,
	}
	var claimerType sql.NullString
	var claimerId sql.NullInt64
	if err := d.DB.QueryRow(`
		SELECT uploaded, uploader_id, url, mimetype, claimer_type, claimer_id
		FROM photo
		WHERE id = ?`, photoId,
	).Scan(
		&photo.Uploaded, &photo.UploaderId, &photo.URL, &photo.Mimetype,
		&claimerType, &claimerId,
	); err == sql.ErrNoRows {
		return nil, NoRowsError
	} else if err != nil {
		return nil, err
	}
	photo.ClaimerType = claimerType.String
	photo.ClaimerId = claimerId.Int64
	return photo, nil
}

func (d *DataService) ClaimPhoto(photoId int64, claimerType string, claimerId int64) error {
	return d.claimPhoto(d.DB, photoId, claimerType, claimerId)
}

func (d *DataService) claimPhoto(db db, photoId int64, claimerType string, claimerId int64) error {
	_, err := db.Exec(`
		UPDATE photo
		SET claimer_type = ?, claimer_id = ?
		WHERE id = ?`, claimerType, claimerId, photoId)
	return err
}
