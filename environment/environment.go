// Package environment provides a way in which for us to set and
// pull in the current environment stage in any package that we see it being
// necessary to do so.
package environment

import "sync"

const (
	Corp    = "corp"
	Dev     = "dev"
	Prod    = "prod"
	Test    = "test"
	Staging = "staging"
	Demo    = "demo"
	Local   = "local"
)

var current = Test
var once sync.Once

// SetCurrent should be called at startup to set the current environment variable
// so as to make it possible for any package to pull in the current state to act on it
func SetCurrent(env string) {
	once.Do(func() {
		switch env {
		case Dev, Test, Staging, Prod, Demo, Corp, Local:
			current = env
		default:
			panic("unexpected environment: " + env)
		}

	})

}

func GetCurrent() string {
	return current
}

func IsTest() bool {
	return current == Test
}

func IsLocal() bool {
	return current == Local
}

func IsCorp() bool {
	return current == Corp
}

func IsDev() bool {
	return current == Dev
}

func IsStaging() bool {
	return current == Staging
}

func IsProd() bool {
	return current == Prod
}

func IsDemo() bool {
	return current == Demo
}
