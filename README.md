# Rate Limiter em Go com Redis

Este projeto para avaliação implementa rate limiter em Go language que limita o número de requisições por IP ou token de acesso, usando Redis como mecanismo de armazenamento (facilita a implementação das limitações e testes).

## Funcionalidades

- Middleware para servidores web Go (Gin)
- Limitação de requisições por endereço IP
- Limitação de requisições por token de acesso
- Priorização do limite do token sobre o limite do IP
- Bloqueio temporário de IPs ou tokens que excedem o limite
- Armazenamento em Redis com possibilidade de trocar o backend (Strategy Pattern)
- Configuração via variáveis de ambiente ou arquivo .env

## Requisitos

- Linguagem Go e VS Code ou outas IDE's de desenvolvimento
- Docker c/ WSL 2
- Redis

## Configuração

O rate limiter pode ser configurado através de variáveis de ambiente ou arquivo .env:

```env
# Configurações do Rate Limiter
RATE_LIMITER_IP_MAX_REQUESTS=10        # Máximo de requisições por IP
RATE_LIMITER_TOKEN_MAX_REQUESTS=100    # Máximo de requisições por token
RATE_LIMITER_TIME_WINDOW_SECONDS=1     # Janela de tempo em segundos
BLOCK_DURATION_SECONDS=300             # Duração do bloqueio em segundos

# Configurações do Redis
REDIS_HOST=redis                       # Host do Redis
REDIS_PORT=6379                        # Porta do Redis
REDIS_PASSWORD=                        # Senha do Redis (opcional)
REDIS_DB=0                             # Banco de dados Redis
```

## Script utilizado no Redis

Foi utilizada a linguagem Lua no script desenvolvido para a lógica no Redis.
