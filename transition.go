package fsm

type transition struct {
	dst       string
	condition string
	action    string
}
