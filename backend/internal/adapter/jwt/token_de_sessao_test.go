package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type relogioFixo struct{ agora time.Time }

func (r relogioFixo) Agora() time.Time { return r.agora }

func decodificarPayload(t *testing.T, token string) map[string]any {
	t.Helper()
	partes := strings.Split(token, ".")
	if len(partes) != 3 {
		t.Fatalf("token com formato inesperado: %d partes", len(partes))
	}
	bruto, err := base64.RawURLEncoding.DecodeString(partes[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	var mapa map[string]any
	if err := json.Unmarshal(bruto, &mapa); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return mapa
}

func TestTokenDeSessao_L17_AdministradorNaoTemClaimIns(t *testing.T) {
	relogio := relogioFixo{agora: time.Now()}
	ts := NovoTokenDeSessao("segredo-de-teste", relogio)

	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}

	token, _, err := ts.Emitir(ator, relogio.agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	payload := decodificarPayload(t, token)
	if _, existe := payload["ins"]; existe {
		t.Fatal("token do administrador não deveria conter a claim ins")
	}
	for _, chave := range []string{"perfil", "nome", "email"} {
		if _, existe := payload[chave]; existe {
			t.Fatalf("token não deveria conter a claim %q", chave)
		}
	}
}

func TestTokenDeSessao_ValidarRecusaAssinaturaAlterada(t *testing.T) {
	relogio := relogioFixo{agora: time.Now()}
	ts := NovoTokenDeSessao("segredo-de-teste", relogio)
	ator, _ := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	token, _, _ := ts.Emitir(ator, relogio.agora)

	adulterado := token[:len(token)-4] + "abcd"
	if _, err := ts.Validar(adulterado); err == nil {
		t.Fatal("esperava erro para assinatura adulterada")
	}
}

func TestTokenDeSessao_ValidarRecusaAlgNone(t *testing.T) {
	relogio := relogioFixo{agora: time.Now()}
	ts := NovoTokenDeSessao("segredo-de-teste", relogio)

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"` + uuid.Must(uuid.NewV7()).String() + `","exp":` + itoa(relogio.agora.Add(time.Hour).Unix()) + `}`))
	tokenAlgNone := header + "." + claims + "."

	if _, err := ts.Validar(tokenAlgNone); err == nil {
		t.Fatal("esperava erro para token com alg=none")
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func TestTokenDeSessao_SE05_TokenEmitidoHaMaisDe8HorasEhRecusado(t *testing.T) {
	emissao := time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC)
	ator, _ := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)

	tsNaEmissao := NovoTokenDeSessao("segredo-de-teste", relogioFixo{agora: emissao})
	token, _, err := tsNaEmissao.Emitir(ator, emissao)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	nove := emissao.Add(9 * time.Hour)
	tsDepois := NovoTokenDeSessao("segredo-de-teste", relogioFixo{agora: nove})
	if _, err := tsDepois.Validar(token); err == nil {
		t.Fatal("esperava erro: token emitido há mais de 8 horas")
	}

	seteEMeia := emissao.Add(7*time.Hour + 30*time.Minute)
	tsDentro := NovoTokenDeSessao("segredo-de-teste", relogioFixo{agora: seteEMeia})
	if _, err := tsDentro.Validar(token); err != nil {
		t.Fatalf("token dentro das 8 horas não deveria ser recusado: %v", err)
	}
}

func TestTokenDeSessao_EmtSobreviveAIdaEVoltaComPrecisaoDeMicrossegundo(t *testing.T) {
	relogio := relogioFixo{agora: time.Now()}
	ts := NovoTokenDeSessao("segredo-de-teste", relogio)
	conjunto, _ := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	ator, _ := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, ptrUUID(uuid.Must(uuid.NewV7())))

	instante := time.Now().Truncate(time.Microsecond)
	token, _, err := ts.Emitir(ator, instante)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	claims, err := ts.Validar(token)
	if err != nil {
		t.Fatalf("Validar: %v", err)
	}
	if claims.EmitidoEmMicro != instante.UnixMicro() {
		t.Fatalf("emt não sobreviveu com precisão de microssegundo: esperado %d, obtido %d", instante.UnixMicro(), claims.EmitidoEmMicro)
	}
}

func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }
