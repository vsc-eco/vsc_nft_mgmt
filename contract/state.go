// state.go
package main

type State interface {
	Set(key, value string)
	Get(key string) *string
	Delete(key string)
}
