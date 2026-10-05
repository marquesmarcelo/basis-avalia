package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// HealthHandler expõe os endpoints obrigatórios de observabilidade
// (ver CLAUDE.md, seção "Observabilidade").
type HealthHandler struct {
	db                     *sqlx.DB
	redis                  *redis.Client
	versao                 string
	verificarArmazenamento func(ctx context.Context) error
	verificarSMTP          func(ctx context.Context) error
}

func NewHealthHandler(db *sqlx.DB, redisClient *redis.Client, versao string) *HealthHandler {
	return &HealthHandler{db: db, redis: redisClient, versao: versao}
}

// ComVerificacaoDeArmazenamento liga a checagem do MinIO/S3 ao /readyz
// (specs/plano-acao/design.md §9) — opcional para não acoplar health_handler.go
// a um cliente concreto: quem chama passa só a função de checagem
// (tipicamente HeadBucket via port.ArmazenamentoDeObjetos).
func (h *HealthHandler) ComVerificacaoDeArmazenamento(f func(ctx context.Context) error) *HealthHandler {
	h.verificarArmazenamento = f
	return h
}

// ComVerificacaoDeSMTP liga a checagem do servidor de e-mail ao /readyz
// (specs/metas-coordenacao/design.md §13) — mesmo padrão da verificação
// de armazenamento, opcional para não acoplar este arquivo a um cliente
// SMTP concreto.
func (h *HealthHandler) ComVerificacaoDeSMTP(f func(ctx context.Context) error) *HealthHandler {
	h.verificarSMTP = f
	return h
}

// Liveness — o processo está vivo. Nunca verifica dependência externa:
// se o Postgres cair, o pod não deve ser reiniciado por isso (é o /readyz
// que sai da rotação, não o /healthz que mata o processo).
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness — as dependências (Postgres, Redis) estão acessíveis. O corpo
// da resposta nunca carrega o erro do driver (usuário, banco, IP, porta,
// SQLSTATE) — só o nome da dependência indisponível. O erro real vai só
// para o log do servidor (design.md §16 T-103, achado M-2).
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	falhas := []string{}

	if err := h.db.PingContext(ctx); err != nil {
		log.Printf("readyz: postgres indisponível: %v", err)
		falhas = append(falhas, "postgres")
	}
	if err := h.redis.Ping(ctx).Err(); err != nil {
		log.Printf("readyz: redis indisponível: %v", err)
		falhas = append(falhas, "redis")
	}
	if h.verificarArmazenamento != nil {
		if err := h.verificarArmazenamento(ctx); err != nil {
			log.Printf("readyz: armazenamento de objetos indisponível: %v", err)
			falhas = append(falhas, "armazenamento")
		}
	}
	if h.verificarSMTP != nil {
		if err := h.verificarSMTP(ctx); err != nil {
			log.Printf("readyz: servidor SMTP indisponível: %v", err)
			falhas = append(falhas, "smtp")
		}
	}

	if len(falhas) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "indisponivel", "falhas": falhas})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// VersaoResponse é o corpo de GET /version.
type VersaoResponse struct {
	Versao string `json:"versao"`
}

// Versao godoc
// @Summary      Versão da aplicação em execução
// @Tags         sistema
// @Produce      json
// @Success      200 {object} VersaoResponse
// @Router       /version [get]
func (h *HealthHandler) Versao(c *gin.Context) {
	c.JSON(http.StatusOK, VersaoResponse{Versao: h.versao})
}
