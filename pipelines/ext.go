package pipelines

import "github.com/gratno/govacq/entity"

// publish_time

func FindKeyword2(uuid uint) []entity.GovArticle {
	var arts []entity.GovArticle
	DB.Where("uuid=? and status>0", uuid).Find(&arts)
	return arts
}
