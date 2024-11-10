package words

type (
	Turn struct {
		ID       string
		GameID   string
		PlayerID string
		Round    int
		Word     Word
	}
)
