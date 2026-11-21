package handlers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// explicitFalseColumns renvoie les colonnes demandées que le corps JSON envoie
// explicitement à false. Un champ absent ou null n'y figure pas : le défaut du
// modèle s'applique. Le corps doit avoir été lu avec ShouldBindBodyWith(…, binding.JSON).
func explicitFalseColumns(c *gin.Context, columns ...string) map[string]interface{} {
	values := map[string]interface{}{}
	var body map[string]json.RawMessage
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		return values
	}
	for _, column := range columns {
		if raw, ok := body[column]; ok && string(raw) == "false" {
			values[column] = false
		}
	}
	return values
}

// keepExplicitFalse réapplique après un Create les false envoyés par le client.
// GORM omet à l'insertion un booléen false dont le champ porte default:true, et la
// base y met alors true : un élément créé « inactif » serait enregistré actif.
// model doit être le pointeur passé à Create (sa clé primaire est renseignée).
func keepExplicitFalse(db *gorm.DB, c *gin.Context, model interface{}, columns ...string) error {
	values := explicitFalseColumns(c, columns...)
	if len(values) == 0 {
		return nil
	}
	return db.Model(model).Omit(clause.Associations).Updates(values).Error
}

// createKeepingFalse insère model et réapplique les false explicites, dans une transaction.
func createKeepingFalse(db *gorm.DB, c *gin.Context, model interface{}, columns ...string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(model).Error; err != nil {
			return err
		}
		return keepExplicitFalse(tx, c, model, columns...)
	})
}
