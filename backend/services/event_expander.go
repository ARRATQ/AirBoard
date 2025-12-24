package services

import (
	"airboard/models"
	"encoding/json"
	"time"
)

// ExpandRecurringEvents expand les événements récurrents dans une plage de dates
// Retourne toutes les instances des événements récurrents entre startDate et endDate
func ExpandRecurringEvents(events []models.Event, startDate, endDate time.Time) []models.RecurringEventInstance {
	var instances []models.RecurringEventInstance

	for _, event := range events {
		if !event.IsRecurring {
			continue
		}

		// Parser la règle de récurrence
		var pattern models.RecurrencePattern
		if err := json.Unmarshal([]byte(event.RecurrenceRule), &pattern); err != nil {
			continue // Ignorer si JSON invalide
		}

		// Parser les exceptions (dates annulées)
		var exceptions []string
		if event.RecurrenceExceptions != "" {
			json.Unmarshal([]byte(event.RecurrenceExceptions), &exceptions)
		}

		// Générer les instances
		eventInstances := generateInstances(event, pattern, startDate, endDate, exceptions)
		instances = append(instances, eventInstances...)
	}

	return instances
}

// generateInstances génère toutes les instances d'un événement récurrent
func generateInstances(
	event models.Event,
	pattern models.RecurrencePattern,
	startDate, endDate time.Time,
	exceptions []string,
) []models.RecurringEventInstance {
	var instances []models.RecurringEventInstance

	// Date de fin de récurrence
	var recurrenceEnd time.Time
	if event.RecurrenceEnd != nil {
		recurrenceEnd = *event.RecurrenceEnd
	} else if pattern.EndType == "on_date" && pattern.EndDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *pattern.EndDate)
		if err == nil {
			recurrenceEnd = parsedDate
		} else {
			recurrenceEnd = endDate.AddDate(10, 0, 0) // Par défaut: 10 ans dans le futur
		}
	} else {
		recurrenceEnd = endDate.AddDate(10, 0, 0) // Par défaut: 10 ans dans le futur
	}

	// Limiter à la fenêtre demandée
	if recurrenceEnd.After(endDate) {
		recurrenceEnd = endDate
	}

	current := event.StartDate
	occurrenceCount := 0
	maxOccurrences := pattern.OccurrenceCount
	if maxOccurrences == 0 {
		maxOccurrences = 1000 // Limite de sécurité
	}

	for current.Before(recurrenceEnd) && occurrenceCount < maxOccurrences {
		// Vérifier si l'instance est dans la plage demandée
		if (current.Equal(startDate) || current.After(startDate)) && current.Before(endDate.AddDate(0, 0, 1)) {
			// Vérifier si l'instance est annulée
			dateStr := current.Format("2006-01-02")
			isCancelled := contains(exceptions, dateStr)

			instance := models.RecurringEventInstance{
				Event:        event,
				InstanceDate: current,
				OriginalDate: event.StartDate,
				IsCancelled:  isCancelled,
			}
			instances = append(instances, instance)
		}

		// Calculer la prochaine occurrence
		current = getNextOccurrence(current, pattern)
		occurrenceCount++

		// Sécurité: arrêter si la date devient invalide
		if current.IsZero() || current.After(time.Now().AddDate(100, 0, 0)) {
			break
		}
	}

	return instances
}

// getNextOccurrence calcule la prochaine occurrence selon le pattern de récurrence
func getNextOccurrence(current time.Time, pattern models.RecurrencePattern) time.Time {
	switch pattern.Type {
	case "daily":
		return current.AddDate(0, 0, pattern.Interval)

	case "weekly":
		// Pour hebdomadaire, on cherche le prochain jour de la semaine correspondant
		if len(pattern.DaysOfWeek) == 0 {
			// Par défaut, même jour de la semaine
			return current.AddDate(0, 0, 7*pattern.Interval)
		}

		// Trouver le prochain jour correspondant
		next := current.AddDate(0, 0, 1) // Commencer par le lendemain
		daysChecked := 0
		maxDays := 7 * pattern.Interval * 2 // Limite de sécurité

		for daysChecked < maxDays {
			weekday := int(next.Weekday())
			if contains(pattern.DaysOfWeek, weekday) {
				// Vérifier que c'est bien dans le bon intervalle de semaines
				weeksSinceStart := int(next.Sub(current).Hours() / 24 / 7)
				if weeksSinceStart >= pattern.Interval || weeksSinceStart == 0 {
					return next
				}
			}
			next = next.AddDate(0, 0, 1)
			daysChecked++
		}

		// Fallback: même jour de la semaine après N semaines
		return current.AddDate(0, 0, 7*pattern.Interval)

	case "monthly":
		next := current.AddDate(0, pattern.Interval, 0)

		// Si un jour spécifique du mois est défini
		if pattern.DayOfMonth > 0 {
			// Construire la date avec le jour spécifié
			year, month, _ := next.Date()
			dayOfMonth := pattern.DayOfMonth

			// Gérer les mois avec moins de jours (ex: 31 février -> 28/29 février)
			lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, next.Location()).Day()
			if dayOfMonth > lastDayOfMonth {
				dayOfMonth = lastDayOfMonth
			}

			next = time.Date(year, month, dayOfMonth, next.Hour(), next.Minute(), next.Second(), next.Nanosecond(), next.Location())
		}

		return next

	case "yearly":
		return current.AddDate(pattern.Interval, 0, 0)

	default:
		// Par défaut, incrémenter d'un jour
		return current.AddDate(0, 0, 1)
	}
}

// contains vérifie si une slice contient une valeur
func contains[T comparable](slice []T, value T) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
