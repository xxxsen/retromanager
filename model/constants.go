package model

type ActionType int

const (
	ActionCreate ActionType = 1
	ActionModify ActionType = 2
	ActionDelete ActionType = 3
)

type GameState int

const (
	GameStateNormal = 1
	GameStateDelete = 2
)
