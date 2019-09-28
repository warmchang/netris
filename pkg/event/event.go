package event

type Event struct {
	Player  int
	Message string
}

type ScoreEvent struct {
	Event
	Score int
}
