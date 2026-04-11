package payment

import (
	"google.golang.org/protobuf/runtime/protoimpl"
)

type PaymentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OrderId string `protobuf:"bytes,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	Amount  int64  `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *PaymentRequest) Reset()         { *x = PaymentRequest{} }
func (x *PaymentRequest) String() string { return "PaymentRequest" }
func (*PaymentRequest) ProtoMessage()    {}

func (x *PaymentRequest) GetOrderId() string {
	if x != nil {
		return x.OrderId
	}
	return ""
}

func (x *PaymentRequest) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

type PaymentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransactionId string `protobuf:"bytes,1,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	Status        string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *PaymentResponse) Reset()         { *x = PaymentResponse{} }
func (x *PaymentResponse) String() string { return "PaymentResponse" }
func (*PaymentResponse) ProtoMessage()    {}

func (x *PaymentResponse) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

func (x *PaymentResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}
