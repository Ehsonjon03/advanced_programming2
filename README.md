## Как работает мой проект

```mermaid
graph TD
    %% Внешний мир
    Postman[("🚀 Postman / Browser")]
    style Postman fill:#ff6c37,stroke:#333,color:#fff

    subgraph OrderApp ["📦 Order Service (Port 8080)"]
        Gin["🌐 Gin Framework (HTTP)"]
        OrderLogic["🧠 Business Logic"]
        GRPC_Client["🔌 gRPC Client"]
    end

    subgraph PaymentApp ["💳 Payment Service (Port 50051)"]
        GRPC_Server["⚙️ gRPC Server"]
        PayLogic["🧠 Payment Logic"]
    end

    %% Потоки
    Postman -->|HTTP Request| Gin
    Gin --> OrderLogic
    OrderLogic -->|Call| GRPC_Client

    %% Тот самый gRPC переход
    GRPC_Client -.->|Protobuf / gRPC| GRPC_Server

    GRPC_Server --> PayLogic

    %% Базы
    OrderLogic --- DB1[("🛢️ order_db")]
    PayLogic --- DB2[("🛢️ payment_db")]

    %% Стилизация
    style OrderApp fill:#e1f5fe,stroke:#01579b
    style PaymentApp fill:#fff3e0,stroke:#ef6c00