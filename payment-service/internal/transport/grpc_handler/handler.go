package grpc_handler

import (
	"context"
	"payment-service/internal/usecase"
	"payment-service/pkg/payment"
)

type PaymentGRPCHandler struct {
	// Встраиваем нереализованный сервер, чтобы удовлетворить интерфейсу gRPC
	payment.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

func NewPaymentGRPCHandler(uc *usecase.PaymentUseCase) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{useCase: uc}
}

// ProcessPayment — старый метод (обработка одного платежа)
func (h *PaymentGRPCHandler) ProcessPayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	res, err := h.useCase.Authorize(req.OrderId, req.Amount)
	if err != nil {
		return nil, err
	}

	return &payment.PaymentResponse{
		TransactionId: res.TransactionID,
		Status:        res.Status,
	}, nil
}

// ListPayments — НОВЫЙ метод (получение списка)
func (h *PaymentGRPCHandler) ListPayments(ctx context.Context, req *payment.ListPaymentsRequest) (*payment.ListPaymentsResponse, error) {
	// 1. Получаем статус из запроса
	status := req.GetStatus()

	// 2. Вызываем UseCase
	paymentsData, err := h.useCase.GetAll(status)
	if err != nil {
		return nil, err
	}

	// 3. Собираем ответ для gRPC
	var protoPayments []*payment.PaymentResponse
	for _, p := range paymentsData {
		protoPayments = append(protoPayments, &payment.PaymentResponse{
			TransactionId: p.TransactionID,
			Status:        p.Status,
		})
	}

	return &payment.ListPaymentsResponse{
		Payments: protoPayments,
	}, nil
}
