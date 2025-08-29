package main

func main() {
	debug := true
	InitState(debug, "state.json") // true = use MockState
	InitSDKInterface(debug)        // enable mock env/sdk
}
