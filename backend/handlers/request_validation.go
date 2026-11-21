package handlers

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// Bornes de la longueur maximale d'un commentaire (mêmes valeurs que le champ de l'admin).
const (
	minCommentLengthSetting = 100
	maxCommentLengthSetting = 5000
)

// msgEndBeforeStart est renvoyé quand une période finit avant de commencer.
const msgEndBeforeStart = "La date de fin doit être postérieure à la date de début"

// endBeforeStart indique une période incohérente : les deux bornes sont fournies
// et la fin précède le début. Une fin égale au début est acceptée.
func endBeforeStart(start, end *time.Time) bool {
	return start != nil && end != nil && end.Before(*start)
}

// validCommentLength vérifie le réglage « longueur maximale d'un commentaire » :
// 0 ou une valeur trop basse bloquerait tout commentaire.
func validCommentLength(value int) bool {
	return value >= minCommentLengthSetting && value <= maxCommentLengthSetting
}

// isUniqueViolation reconnaît une violation de contrainte d'unicité PostgreSQL
// (SQLSTATE 23505), par exemple un nom ou un slug déjà pris.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
