## Архитектура системы (gRPC)

Ниже представлена схема взаимодействия сервисов через gRPC:

```mermaid
sequenceDiagram
    participant Client as Postman / User
    participant Order as Order Service (Port 8080)
    participant DB1 as Order DB (PostgreSQL)
    participant Payment as Payment Service (Port 50051)
    participant DB2 as Payment DB (PostgreSQL)

    Client->>Order: POST /orders (ID, Amount)
    Order->>DB1: Сохранение заказа (Status: Pending)
    
    Note over Order,Payment: Общение через gRPC (Protobuf)
    
    Order->>Payment: gRPC: ProcessPayment(OrderID, Amount)
    
    alt Amount > 100,000
        Payment-->>Order: gRPC Response (Status: Rejected)
    else Amount <= 100,000
        Payment->>DB2: Сохранение транзакции
        Payment-->>Order: gRPC Response (Status: Authorized)
    end

    Order-->>Client: JSON Response (Order Created & Processed)