package provider

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

// EmailProvider - интерфейс адаптера
type EmailProvider interface {
	SendEmail(to string, body string) error
}

// SimulatedEmailProvider - имитация внешнего API (как требует задание)
type SimulatedEmailProvider struct{}

func (s *SimulatedEmailProvider) SendEmail(to string, body string) error {
	// Имитируем задержку сети
	time.Sleep(1 * time.Second)

	// Имитируем случайную ошибку в 30% случаев для проверки Retries
	if rand.Float32() < 0.3 {
		return errors.New("external email provider timeout")
	}

	log.Printf("[EMAIL SENT] To: %s | Body: %s", to, body)
	return nil
}
