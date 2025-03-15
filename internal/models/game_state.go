package models

import "time"

// Score est un type flexible qui peut être soit un nombre, soit une chaîne (pour "crossed")
type Score interface{}

// PlayerScores représente les scores d'un joueur pour chaque combinaison
type PlayerScores struct {
	Ones          Score `json:"ones,omitempty"`
	Twos          Score `json:"twos,omitempty"`
	Threes        Score `json:"threes,omitempty"`
	Fours         Score `json:"fours,omitempty"`
	Fives         Score `json:"fives,omitempty"`
	Sixes         Score `json:"sixes,omitempty"`
	ThreeOfAKind  Score `json:"threeOfAKind,omitempty"`
	FourOfAKind   Score `json:"fourOfAKind,omitempty"`
	FullHouse     Score `json:"fullHouse,omitempty"`
	SmallStraight Score `json:"smallStraight,omitempty"`
	LargeStraight Score `json:"largeStraight,omitempty"`
	Yahtzee       Score `json:"yahtzee,omitempty"`
	Chance        Score `json:"chance,omitempty"`
}

// Player représente un joueur dans la partie en cours
type Player struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Scores PlayerScores `json:"scores"`
}

// HistoryPlayer représente un joueur dans l'historique des parties
type HistoryPlayer struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Score            int    `json:"score"`
}

// GameHistoryEntry représente une entrée dans l'historique des parties
type GameHistoryEntry struct {
	ID        string          `json:"id"`
	Date      time.Time       `json:"date"`
	Players   []HistoryPlayer `json:"players"`
	WinnerID  string          `json:"winnerId"`
}

// GameState représente l'état complet du jeu
type GameState struct {
	Players     []Player          `json:"players"`
	IsStarted   bool              `json:"isStarted"`
	GameHistory []GameHistoryEntry `json:"gameHistory"`
}

// GameStateMessage représente un message WebSocket contenant un état de jeu
type GameStateMessage struct {
	State   GameState `json:"state"`
	Version int       `json:"version"`
}
