package anexo

import (
	"context"
	"errors"
	"io"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BaixarResultado struct {
	NomeArquivo  string
	TipoConteudo string
	Conteudo     io.ReadCloser
}

// BaixarUseCase serve GET /api/v1/anexos/{id}/conteudo (design.md §6.5,
// AN-04 a AN-07). A chave do objeto NUNCA vem do cliente — é sempre lida
// da linha já recortada por Escopo, só então entregue ao armazenamento.
type BaixarUseCase struct {
	repo          port.EntregaRepository
	armazenamento port.ArmazenamentoDeObjetos
	audit         port.AuditLogger
}

func NovoBaixarUseCase(repo port.EntregaRepository, armazenamento port.ArmazenamentoDeObjetos, audit port.AuditLogger) *BaixarUseCase {
	return &BaixarUseCase{repo: repo, armazenamento: armazenamento, audit: audit}
}

func (uc *BaixarUseCase) Executar(ctx context.Context, ator autorizacao.Ator, anexoID uuid.UUID) (BaixarResultado, error) {
	esc, err := autorizarAnexoLeitura(ator)
	if err != nil {
		return BaixarResultado{}, err
	}
	nomeOriginal, tipoBruto, chaveObjeto, err := uc.repo.BuscarAnexoParaDownload(ctx, esc, anexoID)
	if err != nil {
		uc.auditarNegado(ctx, ator, esc, anexoID)
		return BaixarResultado{}, err
	}

	conteudo, _, err := uc.armazenamento.Ler(ctx, chaveObjeto)
	if err != nil {
		return BaixarResultado{}, err
	}

	tipo, _ := valueobject.NovoTipoDeAnexo(tipoBruto)
	usuarioID := ator.UsuarioID()
	evento := auditoria.NovoEvento(auditoria.BaixarAnexo, auditoria.ResultadoSucesso)
	evento.AtorID = &usuarioID
	evento.InstituicaoID = esc.InstituicaoID()
	evento.RecursoTipo = "Anexo"
	evento.RecursoID = &anexoID
	_ = uc.audit.Registrar(ctx, evento)

	return BaixarResultado{NomeArquivo: nomeOriginal, TipoConteudo: tipo.MimeType(), Conteudo: conteudo}, nil
}

func (uc *BaixarUseCase) auditarNegado(ctx context.Context, ator autorizacao.Ator, esc autorizacao.Escopo, anexoID uuid.UUID) {
	usuarioID := ator.UsuarioID()
	evento := auditoria.NovoEvento(auditoria.BaixarAnexo, auditoria.ResultadoNegado)
	evento.AtorID = &usuarioID
	evento.InstituicaoID = esc.InstituicaoID()
	evento.RecursoTipo = "Anexo"
	evento.RecursoID = &anexoID
	_ = uc.audit.Registrar(ctx, evento)
}

func autorizarAnexoLeitura(ator autorizacao.Ator) (autorizacao.Escopo, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.EntregasDaInstituicao, autorizacao.AcaoListar, nil)
	if err == nil {
		return esc, nil
	}
	if !errors.Is(err, domain.ErrPermissaoNegada) {
		return autorizacao.Escopo{}, err
	}
	return autorizacao.Autorizar(ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoListar, nil)
}
