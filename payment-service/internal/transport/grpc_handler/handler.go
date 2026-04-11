package grpc_handler

import (
	"context"
	"payment-service/internal/usecase"
	"payment-service/pkg/payment" // Твой пакет gRPC
)

type PaymentGRPCHandler struct {
	payment.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

func NewPaymentGRPCHandler(uc *usecase.PaymentUseCase) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{useCase: uc}
}

// Метод должен называться в точности как в .proto файле
func (h *PaymentGRPCHandler) ProcessPayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// Вызываем твой UseCase метод Authorize
	res, err := h.useCase.Authorize(req.OrderId, req.Amount)
	if err != nil {
		return nil, err
	}

	// Возвращаем ответ клиенту
	return &payment.PaymentResponse{
		TransactionId: res.TransactionID,
		Status:        res.Status,
	}, nil
}
