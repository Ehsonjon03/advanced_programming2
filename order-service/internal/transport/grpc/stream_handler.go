package grpc

import (
	"log"
	"time"

	"order-service/internal/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
	"order-service/internal/generated/order" // Твой сгенерированный код
)

type OrderStreamHandler struct {
	order.UnimplementedOrderServiceServer
	repo *repository.OrderRepo
}

func NewOrderStreamHandler(repo *repository.OrderRepo) *OrderStreamHandler {
	return &OrderStreamHandler{repo: repo}
}

func (h *OrderStreamHandler) SubscribeToOrderUpdates(req *order.OrderRequest, stream order.OrderService_SubscribeToOrderUpdatesServer) error {
	lastStatus := ""
	log.Printf("Клиент подписался на обновления заказа: %s", req.OrderId)

	for {
		// Проверка на закрытие соединения клиентом
		select {
		case <-stream.Context().Done():
			log.Printf("Клиент отключился от заказа: %s", req.OrderId)
			return nil
		default:
			// ЧИТАЕМ ИЗ БД (Критическое требование задания)
			// Убедись, что метод GetByID возвращает актуальный статус
			ord, err := h.repo.GetByID(req.OrderId)
			if err != nil {
				log.Printf("Ошибка получения данных из БД: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Если статус изменился — пушим в стрим [cite: 44]
			if ord.Status != lastStatus {
				err := stream.Send(&order.OrderStatusUpdate{
					OrderId:   ord.ID,
					Status:    ord.Status,
					UpdatedAt: timestamppb.Now(), // Соответствие google.protobuf.Timestamp
				})
				if err != nil {
					return err
				}
				lastStatus = ord.Status
				log.Printf("Статус заказа %s изменен на: %s", ord.ID, ord.Status)
			}

			// Проверяем каждые 2 секунды (не слишком часто, чтобы не грузить БД)
			time.Sleep(2 * time.Second)
		}
	}
}
