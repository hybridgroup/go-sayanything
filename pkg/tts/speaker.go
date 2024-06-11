package tts

type Speaker interface {
	Connect(string) error
	Close()
	Speech(text string) ([]byte, error)
}
