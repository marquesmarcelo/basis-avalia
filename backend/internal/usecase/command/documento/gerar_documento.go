package documento

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/documento"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type GerarDocumentoInput struct {
	Ator    autorizacao.Ator
	Alcance autorizacao.Alcance // PlanosDaInstituicao (PI) ou PlanosDaCarteira (Coordenador)
	PlanoID uuid.UUID
}

type GerarDocumentoResultado struct {
	DocumentoID uuid.UUID
	NomeArquivo string
	Conteudo    []byte
}

// GerarDocumentoUseCase cobre DO-01 a DO-06. Nenhuma ramificação por
// situação do plano (design.md §7.3, P-10): o único recorte é o Escopo —
// instituição para o PI, instituição e carteira para o coordenador — e
// isso já produz 404 fora dele. Grava o objeto primeiro, a linha depois
// (design.md §7.5): se a transação falhar, sobra um objeto órfão
// (invisível, reclamado por retenção) em vez de uma linha órfã (download
// quebrado visível ao usuário).
type GerarDocumentoUseCase struct {
	planoRepo     port.PlanoRepository
	documentoRepo port.DocumentoRepository
	armazenamento port.ArmazenamentoDeObjetos
	gerador       port.GeradorDeDocumento
	audit         port.AuditLogger
	uow           port.UnidadeDeTrabalho
}

func NovoGerarDocumentoUseCase(
	planoRepo port.PlanoRepository, documentoRepo port.DocumentoRepository,
	armazenamento port.ArmazenamentoDeObjetos, gerador port.GeradorDeDocumento, audit port.AuditLogger, uow port.UnidadeDeTrabalho,
) *GerarDocumentoUseCase {
	return &GerarDocumentoUseCase{
		planoRepo: planoRepo, documentoRepo: documentoRepo,
		armazenamento: armazenamento, gerador: gerador, audit: audit, uow: uow,
	}
}

func (uc *GerarDocumentoUseCase) Executar(ctx context.Context, in GerarDocumentoInput) (GerarDocumentoResultado, error) {
	esc, err := autorizacao.Autorizar(in.Ator, in.Alcance, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return GerarDocumentoResultado{}, err
	}

	detalhe, err := uc.planoRepo.BuscarPorID(ctx, esc, in.PlanoID)
	if err != nil {
		return GerarDocumentoResultado{}, err
	}
	// PeriodoDoPlano, nunca port.PeriodoRepository: o alcance de carteira
	// do coordenador não confere PeriodoGerenciar, e período é entidade da
	// instituição, não do curso — ver comentário em port/plano_repository.go.
	per, err := uc.planoRepo.PeriodoDoPlano(ctx, esc, detalhe.Plano.PeriodoID)
	if err != nil {
		return GerarDocumentoResultado{}, err
	}
	instituicaoNome, instituicaoSigla, err := uc.planoRepo.InstituicaoDaSessao(ctx, esc)
	if err != nil {
		return GerarDocumentoResultado{}, err
	}

	dados := montarDadosDoDocumento(detalhe, per, instituicaoNome, instituicaoSigla, esc.DataDeReferencia())
	conteudo, err := uc.gerador.Gerar(dados)
	if err != nil {
		return GerarDocumentoResultado{}, err
	}

	agora := time.Now()
	nomeArquivo := fmt.Sprintf("plano-acao-%s-%s.docx", detalhe.CursoNome, agora.Format("2006-01-02"))
	chave := fmt.Sprintf("documento-plano/%s/%s/%d-%s.docx", detalhe.Plano.InstituicaoID, detalhe.Plano.ID, agora.UnixNano(), uuid.Must(uuid.NewV7()))
	if err := uc.armazenamento.Gravar(ctx, chave, bytes.NewReader(conteudo), "application/vnd.openxmlformats-officedocument.wordprocessingml.document", int64(len(conteudo))); err != nil {
		return GerarDocumentoResultado{}, err
	}

	novoDocumento := documento.NovoDocumento(detalhe.Plano.ID, detalhe.Plano.CursoID, detalhe.Plano.InstituicaoID, chave, nomeArquivo, detalhe.Situacao, in.Ator.UsuarioID())
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.documentoRepo.Inserir(ctx, esc, novoDocumento); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.GerarDocumento, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Documento"
		evento.RecursoID = &novoDocumento.ID
		evento.Detalhes["plano_id"] = detalhe.Plano.ID.String()
		return uc.audit.Registrar(ctx, evento)
	})
	if erro != nil {
		return GerarDocumentoResultado{}, erro
	}

	return GerarDocumentoResultado{DocumentoID: novoDocumento.ID, NomeArquivo: nomeArquivo, Conteudo: conteudo}, nil
}

func montarDadosDoDocumento(detalhe port.DetalhePlano, per periodo.Periodo, instituicaoNome, instituicaoSigla string, hoje valueobject.DataLocal) port.DadosDoDocumento {
	itens := make([]port.ItemDoDocumento, 0, len(detalhe.Itens))
	for _, item := range detalhe.Itens {
		indicadores := make([]string, 0, len(item.Indicadores))
		for _, ind := range item.Indicadores {
			origem := "Do INEP"
			if ind.Escopo == "instituicao" {
				origem = "Da instituição"
			}
			indicadores = append(indicadores, fmt.Sprintf("%s (%s)", ind.Codigo, origem))
		}
		itens = append(itens, port.ItemDoDocumento{MetaNome: item.MetaNome, Indicadores: indicadores, Quantidade: item.Quantidade})
	}

	coordenadorNome, coordenadorPortaria := "", ""
	if detalhe.CoordenadorNome != nil {
		coordenadorNome = *detalhe.CoordenadorNome
	}
	if detalhe.CoordenadorPortaria != nil {
		coordenadorPortaria = *detalhe.CoordenadorPortaria
	}

	marcaSituacao := ""
	switch detalhe.Situacao {
	case string(valueobject.PlanoRascunho):
		marcaSituacao = "RASCUNHO"
	case string(valueobject.PlanoEncerrado):
		fim := per.Vigencia.Fim()
		dataTexto := ""
		if fim != nil {
			dataTexto = formatarDataBR(fim.String())
		}
		marcaSituacao = "ENCERRADO — período encerrado em " + dataTexto
	}

	aprovacaoData, aprovacaoOrgao := "", ""
	if detalhe.Plano.Aprovacao != nil {
		aprovacaoData = formatarDataBR(detalhe.Plano.Aprovacao.Data().String())
		aprovacaoOrgao = rotuloOrgao(detalhe.Plano.Aprovacao.Orgao())
	}

	return port.DadosDoDocumento{
		InstituicaoNome: instituicaoNome, InstituicaoSigla: instituicaoSigla,
		CursoNome:       detalhe.CursoNome,
		CursoGrau:       detalhe.CursoGrau,
		CursoModalidade: detalhe.CursoModalidade,
		CursoCodigoEMec: detalhe.CursoCodigoEMec,
		CoordenadorNome: coordenadorNome, CoordenadorPortaria: coordenadorPortaria,
		PeriodoNome: detalhe.PeriodoNome, PeriodoInicio: formatarDataBR(per.Vigencia.Inicio().String()),
		PeriodoFim: periodoFimTexto(per.Vigencia.Fim()),
		Descricao:  detalhe.Plano.Descricao, ObjetivoGeral: detalhe.Plano.ObjetivoGeral, ResultadosEsperados: detalhe.Plano.ResultadosEsperados,
		AlinhamentoPDI: detalhe.Plano.AlinhamentoPDI, AlinhamentoPPC: detalhe.Plano.AlinhamentoPPC,
		Itens: itens, AprovacaoData: aprovacaoData, AprovacaoOrgao: aprovacaoOrgao,
		AprovacaoPendente: detalhe.SemAprovacao, MarcaSituacao: marcaSituacao,
		GeradoEmTexto: hoje.String(),
	}
}

func rotuloOrgao(o valueobject.OrgaoDeAprovacao) string {
	if o == valueobject.OrgaoNDE {
		return "NDE"
	}
	return "Colegiado de curso"
}

func periodoFimTexto(fim *valueobject.DataLocal) string {
	if fim == nil {
		return ""
	}
	return formatarDataBR(fim.String())
}

func formatarDataBR(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("02/01/2006")
}
