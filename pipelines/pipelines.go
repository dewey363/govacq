package pipelines

import (
	"github.com/gratno/govacq/entity"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

var DB *gorm.DB

const insertTrigger = `
create trigger tasks_settings_insert_trigger AFTER INSERT on gov_tasks for EACH row 
BEGIN
INSERT INTO gov_map_settings(id,created_at,updated_at) VALUES(new.id,new.created_at,new.updated_at);
INSERT INTO gov_art_settings(id,created_at,updated_at) VALUES(new.id,new.created_at,new.updated_at);
INSERT INTO gov_art_update_settings(id,created_at,updated_at) VALUES(new.id,new.created_at,new.updated_at);
END;`

const deleteTrigger = `
create trigger tasks_settings_delete_trigger AFTER DELETE on gov_tasks for EACH row 
BEGIN
DELETE FROM gov_map_settings WHERE id=old.id;
DELETE FROM gov_art_settings WHERE id=old.id;
DELETE FROM gov_art_update_settings WHERE id=old.id;
END;`

func Init(dialect string, args ...interface{}) {
	var err error
	DB, err = gorm.Open(dialect, args...)
	if err != nil {
		logrus.Fatalln(err)
	}
	DB.SetLogger(logrus.StandardLogger())
	DB.DB().Ping()
	DB.DB().SetMaxIdleConns(1)
	DB.DB().SetConnMaxLifetime(time.Minute)
	DB.DB().SetMaxOpenConns(50)

	DB.AutoMigrate(&entity.GovTask{})
	DB.AutoMigrate(&entity.GovMapSetting{})
	DB.AutoMigrate(&entity.GovArtSetting{})
	DB.AutoMigrate(&entity.GovArtUpdateSetting{})
	DB.AutoMigrate(&entity.GovMap{})
	DB.AutoMigrate(&entity.GovArticle{})
	DB.AutoMigrate(&entity.GovArtRule{})

	var count int
	switch strings.ToLower(dialect) {
	case "mysql":
		DB.DB().QueryRow("SELECT count(*) FROM `information_schema`.`triggers` WHERE EVENT_OBJECT_TABLE='gov_tasks'").Scan(&count)
	case "sqlite3":
		DB.DB().QueryRow("SELECT count(*) from sqlite_master where type='trigger' and tbl_name='gov_tasks'").Scan(&count)
	}
	if count != 2 {
		DB.Exec(insertTrigger)
		DB.Exec(deleteTrigger)
	}
}

func CreateGovTask(urls ...string) {
	for _, url := range urls {
		DB.Create(&entity.GovTask{URL: url})
	}
}

func SaveMaps(maps []entity.GovMap) {
	tx := DB.Begin()
	tx.Set("names", "utf8mb4")
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for _, v := range maps {
		tx.Create(&v)
	}
	tx.Commit()
}

func UpdateMapHDTitle(uid uint) {
	maps := FindGovMaps(uid)
	for _, m := range maps {
		if len(m.SourceURL) > 0 {
			var hdMap entity.GovMap
			DB.Model(&entity.GovMap{}).Where("url=?", m.SourceURL).First(&hdMap)
			m.HDID = hdMap.ID
			m.HDTitle = hdMap.Title
			DB.Model(&entity.GovMap{}).Updates(&m)
		}
	}
}

func UpdateTaskStatus(id uint, status int) {
	DB.Model(&entity.GovTask{}).Where("id=?", id).Update("status", status)

}

func SaveArticles(arts []entity.GovArticle) {
	tx := DB.Begin()
	tx.Set("names", "utf8mb4")
	for i, v := range arts {
		tx.Create(&v)
		if i%20 == 0 {
			tx.Commit()
			tx = DB.Begin()
			tx.Set("names", "utf8mb4")
		}
	}
	tx.Commit()
}

func UpdateMapsStatus(uid uint, status int) {
	DB.Model(&entity.GovMap{}).Where("uid=?", uid).Update("status", status)
}

func UpdateArticle(art entity.GovArticle) {
	DB.Model(&entity.GovArticle{}).Update(art)
}

func DelArtUnScoped() {
	DB.Unscoped().Where("deleted_at is not null").Delete(&entity.GovArticle{})
}

func SaveArtRules(artRules []entity.GovArtRule) {
	tx := DB.Begin()
	tx.Set("names", "utf8mb4")
	for _, v := range artRules {
		tx.Create(&v)
	}
	tx.Commit()
}

func FindMapSetting(id uint) entity.GovMapSetting {
	var setting entity.GovMapSetting
	DB.Find(&setting, id)
	return setting
}

func FindArtSetting(id uint) entity.GovArtSetting {
	var setting entity.GovArtSetting
	DB.Find(&setting, id)
	return setting
}

func FindArtUpdateSetting(id uint) entity.GovArtUpdateSetting {
	var setting entity.GovArtUpdateSetting
	DB.Find(&setting, id)
	return setting
}

func FindArtUpdateRules(id uint, _type int) []entity.GovArtRule {
	var rules []entity.GovArtRule
	DB.Model(&entity.GovArtRule{}).Where("uuid=? and type=?", id, _type).Find(&rules)
	return rules
}

func FindGovTask(id uint) entity.GovTask {
	var govTask entity.GovTask
	DB.Find(&govTask, id)
	return govTask
}

func FindGovMaps(uid uint) []entity.GovMap {
	var maps []entity.GovMap
	DB.Where("uid=?", uid).Find(&maps)
	return maps
}

func FindGovArticles(uuid uint, status int) []entity.GovArticle {
	var arts []entity.GovArticle
	DB.Where("uuid=? and status=?", uuid, status).Find(&arts)
	return arts
}

func CountGovArticles(uuid uint, status int) int {
	var count int
	DB.Model(&entity.GovArticle{}).Where("uuid=? and status=?", uuid, status).Count(&count)
	return count
}
