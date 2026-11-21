package utils

import (
	"log"
	"runtime/debug"
)

// SafeGo lance fn dans une goroutine qui récupère une éventuelle panique.
//
// Le middleware Recovery de Gin ne couvre que la goroutine de la requête :
// une panique dans une goroutine lancée à part (notification, e-mail…)
// arrêterait tout le serveur. Ici elle est journalisée avec sa pile, sous le
// nom donné, et le processus continue.
func SafeGo(name string, fn func()) {
	go func() {
		defer RecoverAndLog(name)
		fn()
	}()
}

// RecoverAndLog s'utilise en `defer` en tête d'une goroutine : il récupère une
// panique et la journalise. Sans effet si aucune panique n'est en cours.
func RecoverAndLog(name string) {
	if r := recover(); r != nil {
		log.Printf("[PANIC] goroutine %q : %v\n%s", name, r, debug.Stack())
	}
}
