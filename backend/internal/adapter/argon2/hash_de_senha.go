package argon2

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/crypto/argon2"
)

type parametros struct {
	memoria     uint32
	iteracoes   uint32
	paralelismo uint8
	tamanhoSalt uint32
	tamanhoHash uint32
}

// Calibrado para HashSenha ficar entre 150ms e 400ms na máquina de
// referência (design.md §13.6) — é o único freio remanescente a tentativa
// automatizada de login, na ausência de limite de tentativas (4.1).
var parametrosPadrao = parametros{
	memoria:     147456, // 144 MiB
	iteracoes:   3,
	paralelismo: 4,
	tamanhoSalt: 16,
	tamanhoHash: 32,
}

var (
	hashDuracaoSegundos = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "hash_senha_duracao_segundos",
		Help:    "Duração de cada cálculo de hash argon2id (gerar ou conferir).",
		Buckets: prometheus.DefBuckets,
	})
	hashEmEspera = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "hash_senha_em_espera",
		Help: "Chamadas de hash aguardando o semáforo de concorrência.",
	})
)

// HashDeSenha implementa port.HashDeSenha com argon2id (D-01) e um
// semáforo que limita cálculos concorrentes (D-02) — enfileira, nunca
// recusa: não é o controle de limite de tentativas que o dono recusou.
type HashDeSenha struct {
	parametros     parametros
	semaforo       chan struct{}
	hashReferencia string
}

func NovoHashDeSenha(concorrencia int) *HashDeSenha {
	if concorrencia < 1 {
		concorrencia = 4
	}
	h := &HashDeSenha{
		parametros: parametrosPadrao,
		semaforo:   make(chan struct{}, concorrencia),
	}
	ref, err := h.gerarSemMetrica("referencia-de-tempo-constante-l08")
	if err != nil {
		panic(fmt.Sprintf("argon2: falha ao gerar hash de referência: %v", err))
	}
	h.hashReferencia = ref
	return h
}

func (h *HashDeSenha) adquirir() {
	hashEmEspera.Inc()
	h.semaforo <- struct{}{}
	hashEmEspera.Dec()
}

func (h *HashDeSenha) liberar() { <-h.semaforo }

func (h *HashDeSenha) gerarSemMetrica(senha string) (string, error) {
	salt := make([]byte, h.parametros.tamanhoSalt)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(senha), salt, h.parametros.iteracoes, h.parametros.memoria, h.parametros.paralelismo, h.parametros.tamanhoHash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.parametros.memoria, h.parametros.iteracoes, h.parametros.paralelismo,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *HashDeSenha) Gerar(ctx context.Context, senha valueobject.SenhaEmTexto) (valueobject.SenhaHash, error) {
	h.adquirir()
	defer h.liberar()

	inicio := time.Now()
	codificado, err := h.gerarSemMetrica(senha.Revelar())
	hashDuracaoSegundos.Observe(time.Since(inicio).Seconds())
	if err != nil {
		return valueobject.SenhaHash{}, err
	}
	return valueobject.NovaSenhaHash(codificado)
}

func (h *HashDeSenha) Conferir(ctx context.Context, senha valueobject.SenhaEmTexto, hash valueobject.SenhaHash) (bool, error) {
	h.adquirir()
	defer h.liberar()

	inicio := time.Now()
	ok, err := conferirContraCodificado(senha.Revelar(), hash.Codificado())
	hashDuracaoSegundos.Observe(time.Since(inicio).Seconds())
	return ok, err
}

// ConferirDescartavel calcula contra o hash de referência fixo (criado na
// construção do adapter, mesmos parâmetros) — nunca autentica ninguém,
// só consome o mesmo tempo de uma conferência real (L-08).
func (h *HashDeSenha) ConferirDescartavel(ctx context.Context, senha valueobject.SenhaEmTexto) {
	h.adquirir()
	defer h.liberar()

	inicio := time.Now()
	_, _ = conferirContraCodificado(senha.Revelar(), h.hashReferencia)
	hashDuracaoSegundos.Observe(time.Since(inicio).Seconds())
}

func conferirContraCodificado(senha, codificado string) (bool, error) {
	partes := strings.Split(codificado, "$")
	if len(partes) != 6 {
		return false, fmt.Errorf("argon2: formato de hash inesperado")
	}
	var memoria, iteracoes uint32
	var paralelismo uint8
	if _, err := fmt.Sscanf(partes[3], "m=%d,t=%d,p=%d", &memoria, &iteracoes, &paralelismo); err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(partes[4])
	if err != nil {
		return false, err
	}
	hashEsperado, err := base64.RawStdEncoding.DecodeString(partes[5])
	if err != nil {
		return false, err
	}

	calculado := argon2.IDKey([]byte(senha), salt, iteracoes, memoria, paralelismo, uint32(len(hashEsperado)))
	return subtle.ConstantTimeCompare(calculado, hashEsperado) == 1, nil
}

var _ port.HashDeSenha = (*HashDeSenha)(nil)
