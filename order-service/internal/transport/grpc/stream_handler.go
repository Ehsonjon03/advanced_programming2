package grpc

import (
	"log"
	"time"

	"order-service/internal/generated/order"
	"order-service/internal/repository"

	"google.golang.org/protobuf/types/known/timestamppb"
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

	// Извлекаем контекст из gRPC стрима
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Клиент отключился от заказа: %s", req.OrderId)
			return nil
		default:
			// Читаем данные через репозиторий (который использует кэш и БД)
			ord, err := h.repo.GetByID(ctx, req.OrderId)
			if err != nil {
				log.Printf("Ошибка получения данных: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Если статус изменился со времени последней проверки — отправляем клиенту
			if ord.Status != lastStatus {
				err := stream.Send(&order.OrderStatusUpdate{
					OrderId:   ord.ID,
					Status:    ord.Status,
					UpdatedAt: timestamppb.Now(),
				})
				if err != nil {
					return err
				}
				lastStatus = ord.Status
				log.Printf("Статус заказа %s изменен на: %s", ord.ID, ord.Status)
			}

			// Опрос раз в 2 секунды
			time.Sleep(2 * time.Second)
		}
	}
}
