package plano

import (
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func dadosValidos() DadosDoPlano {
	return DadosDoPlano{Descricao: "Descrição", ObjetivoGeral: "Objetivo", ResultadosEsperados: "Resultados"}
}

func periodoDe(t *testing.T, inicio, fim string) periodo.Periodo {
	t.Helper()
	di, err := valueobject.DataLocalTexto(inicio)
	if err != nil {
		t.Fatalf("data inicio: %v", err)
	}
	df, err := valueobject.DataLocalTexto(fim)
	if err != nil {
		t.Fatalf("data fim: %v", err)
	}
	p, err := periodo.NovoPeriodo(uuid.Must(uuid.NewV7()), "2026.1", di, df)
	if err != nil {
		t.Fatalf("NovoPeriodo: %v", err)
	}
	return *p
}

func hoje(t *testing.T, data string) valueobject.DataLocal {
	t.Helper()
	d, err := valueobject.DataLocalTexto(data)
	if err != nil {
		t.Fatalf("hoje: %v", err)
	}
	return d
}

// TestSituacaoEfetiva_SI01_RascunhoNaoDependeDoPeriodo prova SI-01: um
// plano em rascunho é sempre "rascunho", mesmo com período aberto.
func TestSituacaoEfetiva_SI01_RascunhoNaoDependeDoPeriodo(t *testing.T) {
	per := periodoDe(t, "2026-01-01", "2026-07-30")
	pl, err := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err != nil {
		t.Fatalf("NovoPlano: %v", err)
	}
	if got := pl.SituacaoEfetiva(per, hoje(t, "2026-03-15")); got != valueobject.PlanoRascunho {
		t.Fatalf("esperava rascunho, obtido %v", got)
	}
}

// TestSituacaoEfetiva_SI11_EncerradoDerivadoDoPeriodo prova SI-11: plano
// vigente cujo período terminou aparece como encerrado sem ninguém
// executar nada.
func TestSituacaoEfetiva_SI11_EncerradoDerivadoDoPeriodo(t *testing.T) {
	per := periodoDe(t, "2026-01-01", "2026-07-30")
	pl, err := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err != nil {
		t.Fatalf("NovoPlano: %v", err)
	}
	if err := pl.Publicar(true, false); err != nil {
		t.Fatalf("Publicar: %v", err)
	}
	if got := pl.SituacaoEfetiva(per, hoje(t, "2026-07-30")); got != valueobject.PlanoVigente {
		t.Fatalf("no último dia do período deveria ser vigente, obtido %v", got)
	}
	if got := pl.SituacaoEfetiva(per, hoje(t, "2026-07-31")); got != valueobject.PlanoEncerrado {
		t.Fatalf("um dia depois deveria ser encerrado, obtido %v", got)
	}
}

// TestSituacaoEfetiva_SI13_ReabrirComPeriodoEncerradoPermaneceEncerrado
// prova SI-13: reabrir limpa o encerramento antecipado, mas a situação
// derivada do período não anula — não é bug, é a resposta correta.
func TestSituacaoEfetiva_SI13_ReabrirComPeriodoEncerradoPermaneceEncerrado(t *testing.T) {
	per := periodoDe(t, "2025-08-01", "2025-12-20")
	pl, err := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err != nil {
		t.Fatalf("NovoPlano: %v", err)
	}
	if err := pl.Publicar(true, false); err != nil {
		t.Fatalf("Publicar: %v", err)
	}
	if err := pl.Encerrar("Encerramento antecipado de teste", time.Now()); err != nil {
		t.Fatalf("Encerrar: %v", err)
	}
	pl.Reabrir()
	if got := pl.SituacaoEfetiva(per, hoje(t, "2026-03-15")); got != valueobject.PlanoEncerrado {
		t.Fatalf("período já encerrado: reabrir não deveria mudar a situação efetiva, obtido %v", got)
	}
}

// TestSemAprovacao_SI07_RascunhoSemAprovacaoNaoAvisa prova SI-07: plano em
// rascunho sem aprovação não recebe aviso nenhum — é o estado normal.
func TestSemAprovacao_SI07_RascunhoSemAprovacaoNaoAvisa(t *testing.T) {
	per := periodoDe(t, "2026-01-01", "2026-07-30")
	pl, err := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err != nil {
		t.Fatalf("NovoPlano: %v", err)
	}
	if pl.SemAprovacao(per, hoje(t, "2026-03-15")) {
		t.Fatal("rascunho sem aprovação não deveria acusar pendência")
	}
}

// TestSemAprovacao_SI05_VigenteSemAprovacaoAvisa prova SI-05: a falta de
// aprovação fica visível quando o plano já está cobrando.
func TestSemAprovacao_SI05_VigenteSemAprovacaoAvisa(t *testing.T) {
	per := periodoDe(t, "2026-01-01", "2026-07-30")
	pl, err := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err != nil {
		t.Fatalf("NovoPlano: %v", err)
	}
	if err := pl.Publicar(true, false); err != nil {
		t.Fatalf("Publicar: %v", err)
	}
	if !pl.SemAprovacao(per, hoje(t, "2026-03-15")) {
		t.Fatal("vigente sem aprovação deveria acusar pendência")
	}
}

// TestPublicar_SI02_ExigePeloMenosUmItem prova SI-02.
func TestPublicar_SI02_ExigePeloMenosUmItem(t *testing.T) {
	per := periodoDe(t, "2026-01-01", "2026-07-30")
	pl, _ := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err := pl.Publicar(false, false); err != domain.ErrPlanoSemItem {
		t.Fatalf("esperava ErrPlanoSemItem, obtido %v", err)
	}
}

// TestPublicar_SI09_PeriodoEncerradoRecusa prova SI-09.
func TestPublicar_SI09_PeriodoEncerradoRecusa(t *testing.T) {
	per := periodoDe(t, "2026-01-01", "2026-07-30")
	pl, _ := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err := pl.Publicar(true, true); err != domain.ErrPeriodoEncerrado {
		t.Fatalf("esperava ErrPeriodoEncerrado, obtido %v", err)
	}
}

// TestEncerrar_SI12_MotivoObrigatorio prova SI-12.
func TestEncerrar_SI12_MotivoObrigatorio(t *testing.T) {
	per := periodoDe(t, "2026-01-01", "2026-07-30")
	pl, _ := NovoPlano(per.InstituicaoID, uuid.Must(uuid.NewV7()), per.ID, dadosValidos())
	if err := pl.Encerrar("", time.Now()); err != domain.ErrMotivoObrigatorio {
		t.Fatalf("esperava ErrMotivoObrigatorio, obtido %v", err)
	}
}

// TestNovoPlano_TextosObrigatorios cobre os três campos de texto vazios.
func TestNovoPlano_TextosObrigatorios(t *testing.T) {
	instituicaoID, cursoID, periodoID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	if _, err := NovoPlano(instituicaoID, cursoID, periodoID, DadosDoPlano{}); err == nil {
		t.Fatal("esperava erro com todos os campos de texto vazios")
	}
}

// TestNovoPlano_PL03_AprovacaoIncompletaPropaga garante que o construtor
// do plano não aceita aprovação pela metade.
func TestNovoPlano_PL03_AprovacaoIncompletaPropaga(t *testing.T) {
	instituicaoID, cursoID, periodoID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	d := dadosValidos()
	d.AprovacaoData = "2026-02-10"
	if _, err := NovoPlano(instituicaoID, cursoID, periodoID, d); err != domain.ErrAprovacaoIncompleta {
		t.Fatalf("esperava ErrAprovacaoIncompleta, obtido %v", err)
	}
}
