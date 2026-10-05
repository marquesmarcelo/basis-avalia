package documento

import (
	"context"
	"errors"
	"io"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BaixarDocumentoResultado struct {
	NomeArquivo  string
	TipoConteudo string
	Conteudo     io.ReadCloser
}

// BaixarDocumentoUseCase serve GET /api/v1/documentos/{id}/conteudo —
// tenta o alcance de instituição (PI) e, se negado, o de carteira
// (Coordenador), porque esta rota é alcançada pelos dois perfis
// (design.md §7.3) e o mecanismo de rota só declara uma permissão por
// caminho (design.md §6.1 herdado). A chave de objeto nunca vem do
// cliente: é sempre lida da linha já recortada por Escopo
// (fundacao-metas.md §9, restrição inegociável).
type BaixarDocumentoUseCase struct {
	documentoRepo port.DocumentoRepository
	armazenamento port.ArmazenamentoDeObjetos
	audit         port.AuditLogger
}

func NovoBaixarDocumentoUseCase(documentoRepo port.DocumentoRepository, armazenamento port.ArmazenamentoDeObjetos, audit port.AuditLogger) *BaixarDocumentoUseCase {
	return &BaixarDocumentoUseCase{documentoRepo: documentoRepo, armazenamento: armazenamento, audit: audit}
}

func (uc *BaixarDocumentoUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (BaixarDocumentoResultado, error) {
	esc, err := autorizacaoDoDocumento(ator)
	if err != nil {
		return BaixarDocumentoResultado{}, err
	}
	doc, err := uc.documentoRepo.BuscarPorID(ctx, esc, id)
	if err != nil {
		return BaixarDocumentoResultado{}, err
	}
	conteudo, tipo, err := uc.armazenamento.Ler(ctx, doc.ChaveObjeto)
	if err != nil {
		return BaixarDocumentoResultado{}, err
	}

	usuarioID := ator.UsuarioID()
	evento := auditoria.NovoEvento(auditoria.BaixarDocumento, auditoria.ResultadoSucesso)
	evento.AtorID = &usuarioID
	evento.InstituicaoID = esc.InstituicaoID()
	evento.RecursoTipo = "Documento"
	evento.RecursoID = &doc.ID
	_ = uc.audit.Registrar(ctx, evento)

	return BaixarDocumentoResultado{NomeArquivo: doc.NomeArquivo, TipoConteudo: tipo, Conteudo: conteudo}, nil
}

// autorizacaoDoDocumento tenta PlanosDaInstituicao (PI) e, só se negado,
// PlanosDaCarteira (Coordenador) — a mesma UNIÃO que Perfil.Pode() já faz
// para quem acumula os dois (VI-02), só que decidida aqui porque a rota
// não escolhe entre os dois de antemão.
func autorizacaoDoDocumento(ator autorizacao.Ator) (autorizacao.Escopo, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoBuscar, nil)
	if err == nil {
		return esc, nil
	}
	if !errors.Is(err, domain.ErrPermissaoNegada) {
		return autorizacao.Escopo{}, err
	}
	return autorizacao.Autorizar(ator, autorizacao.PlanosDaCarteira, autorizacao.AcaoBuscar, nil)
}
