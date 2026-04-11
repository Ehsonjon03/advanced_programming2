### Architecture & Data Flow Diagram

```mermaid
graph TD
    %% Nodes
    Client[Client / Postman]
    
    subgraph Order_Service [Order Service]
        direction TB
        OH[HTTP Handler]
        OUC[Use Case]
        OPG[HTTP Payment Gateway]
        OPR[Postgres Repository]
    end

    subgraph Payment_Service [Payment Service]
        direction TB
        PH[HTTP Handler]
        PUC[Use Case]
        PPR[Postgres Repository]
    end

    ODB[(Order DB)]
    PDB[(Payment DB)]

    %% Connections
    Client -->|POST /orders| OH
    Client -->|GET /orders?min_amount=X| OH
    
    OH --> OUC
    OUC --> OPR
    OUC --> OPG
    OPR --> ODB

    OPG -->|POST /payments| PH
    
    PH --> PUC
    PUC --> PPR
    PPR --> PDB

    %% Styling
    style Order_Service fill:#f9f9f9,stroke:#333,stroke-width:2px
    style Payment_Service fill:#f9f9f9,stroke:#333,stroke-width:2px
    style ODB fill:#e1f5fe,stroke:#01579b
    style PDB fill:#e1f5fe,stroke:#01579b