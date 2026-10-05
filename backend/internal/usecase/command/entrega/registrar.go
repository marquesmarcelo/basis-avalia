package entrega

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/anexo"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	entregadomain "github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// RotaIdempotencia é a chave de rota gravada na tabela `idempotencia`
// (design.md §7) — uma constante, nunca uma string livre repetida em cada
// chamador. Vive aqui (usecase), nunca no adapter Postgres: o valor não é
// SQL, é o identificador estável da operação.
const RotaIdempotencia = "registrar_entrega"

// ArquivoRecebido — o que o handler HTTP já leu e limitou (M-08: nunca
// io.ReadAll sobre um stream ilimitado; o handler aplica MaxBytesReader e
// io.LimitReader ANTES de preencher Conteudo). O use case nunca importa
// mime/multipart nem gin — só bytes já prontos.
type ArquivoRecebido struct {
	NomeOriginal string
	Conteudo     []byte
}

type RegistrarEntregaInput struct {
	Ator           autorizacao.Ator
	ItemPlanoID    uuid.UUID
	Observacao     string
	IdempotencyKey string
	Arquivos       []ArquivoRecebido
}

type RegistrarEntregaResultado struct {
	EntregaID     uuid.UUID
	Reaproveitada bool // 200 em vez de 201 (reenvio com a mesma Idempotency-Key)
}

// RegistrarEntregaUseCase cobre EN-01 a EN-07, EN-09, VG-01, VG-07
// (design.md §5.1, §6, §7). As sete portas são verificadas NESTA ordem —
// a ordem determina qual código o cliente vê.
type RegistrarEntregaUseCase struct {
	repo          port.EntregaRepository
	itemPlanoRepo port.ItemPlanoRepository
	planoRepo     port.PlanoRepository
	armazenamento port.ArmazenamentoDeObjetos
	audit         port.AuditLogger
	uow           port.UnidadeDeTrabalho
}

func NovoRegistrarEntregaUseCase(
	repo port.EntregaRepository, itemPlanoRepo port.ItemPlanoRepository, planoRepo port.PlanoRepository,
	armazenamento port.ArmazenamentoDeObjetos, audit port.AuditLogger, uow port.UnidadeDeTrabalho,
) *RegistrarEntregaUseCase {
	return &RegistrarEntregaUseCase{
		repo: repo, itemPlanoRepo: itemPlanoRepo, planoRepo: planoRepo,
		armazenamento: armazenamento, audit: audit, uow: uow,
	}
}

type anexoValidado struct {
	nomeOriginal string
	tipo         valueobject.TipoDeAnexo
	conteudo     []byte
	hash         string
}

func (uc *RegistrarEntregaUseCase) Executar(ctx context.Context, in RegistrarEntregaInput) (RegistrarEntregaResultado, error) {
	// 1. permissão.
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoCriar, nil)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}

	// 2. item na instituição e na carteira (AplicarEscopo) → 404.
	item, err := uc.itemPlanoRepo.BuscarPorID(ctx, esc, in.ItemPlanoID)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}

	detalhePlano, err := uc.planoRepo.BuscarPorID(ctx, esc, item.PlanoID)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}
	// 3. plano vigente → 409 PLANO_NAO_VIGENTE (M-14: rascunho E
	// encerrado). Depois de QP-3 o coordenador já leu o plano em rascunho
	// — 404 aqui deixaria de esconder e passaria a mentir.
	if detalhePlano.Situacao != string(valueobject.PlanoVigente) {
		return RegistrarEntregaResultado{}, domain.ErrPlanoNaoVigente
	}

	// 4. curso ativo → 409 CURSO_INATIVO.
	ativo, err := uc.planoRepo.CursoValidoParaPlano(ctx, esc, item.CursoID)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}
	if !ativo {
		return RegistrarEntregaResultado{}, domain.ErrCursoInativo
	}

	// 5. curso com designação vigente → 409 CURSO_SEM_COORDENADOR.
	// Defensivo por desenho: a carteira do passo 2 já garante isto na
	// prática, mas o ramo existe para o caminho em que ela é satisfeita
	// por outro caminho (design.md §5.1).
	if detalhePlano.CursoVago {
		return RegistrarEntregaResultado{}, domain.ErrCursoSemCoordenador
	}

	// 6. período aberto → 409 PERIODO_NAO_INICIADO / PERIODO_ENCERRADO.
	per, err := uc.planoRepo.PeriodoDoPlano(ctx, esc, detalhePlano.Plano.PeriodoID)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}
	switch per.Vigencia.SituacaoEm(esc.DataDeReferencia()) {
	case valueobject.Futura:
		return RegistrarEntregaResultado{}, domain.ErrPeriodoNaoIniciado
	case valueobject.Encerrada:
		return RegistrarEntregaResultado{}, domain.ErrPeriodoEncerrado
	}

	// 7. pelo menos um anexo, dentro dos limites — ANTES de gravar
	// qualquer byte (M-08, AN-03).
	anexosValidados, err := validarAnexos(in.Arquivos)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}

	novaEntrega, err := entregadomain.NovaEntrega(item.ID, item.CursoID, item.InstituicaoID, in.Ator.UsuarioID(), in.Observacao)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}

	// Grava os OBJETOS antes da transação (M-09): melhor objeto órfão
	// (invisível, reclamável pela retenção) que linha órfã (download
	// quebrado visível).
	anexosParaGravar, err := uc.gravarObjetos(ctx, item.InstituicaoID, novaEntrega.ID, anexosValidados)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}

	erroTx := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.InserirIdempotencia(ctx, in.Ator.Proprio(), RotaIdempotencia, in.IdempotencyKey, novaEntrega.ID); err != nil {
			return err
		}
		if err := uc.repo.InserirEntregaComAnexos(ctx, esc, novaEntrega, anexosParaGravar); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.RegistrarEntrega, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Entrega"
		evento.RecursoID = &novaEntrega.ID
		evento.Detalhes["item_plano_id"] = item.ID.String()
		evento.Detalhes["curso_id"] = item.CursoID.String()
		return uc.audit.Registrar(ctx, evento)
	})

	if erroTx != nil {
		// best-effort: os objetos já gravados no armazenamento não têm
		// linha que os alcance — órfãos, invisíveis, reclamados pela
		// retenção. Nunca tratado como garantia (M-09).
		removerObjetosOrfaos(ctx, uc.armazenamento, anexosParaGravar)

		if errors.Is(erroTx, domain.ErrChaveDuplicada) {
			return uc.relerPorIdempotencia(ctx, in, esc)
		}
		return RegistrarEntregaResultado{}, erroTx
	}

	return RegistrarEntregaResultado{EntregaID: novaEntrega.ID}, nil
}

// relerPorIdempotencia é chamado FORA da transação que já sofreu
// rollback — o ctx aqui não carrega mais nenhuma *sqlx.Tx (design.md §7,
// M-04): duas requisições simultâneas com a mesma chave se serializam no
// banco, e a segunda relê a entrega original em vez de adivinhar.
func (uc *RegistrarEntregaUseCase) relerPorIdempotencia(ctx context.Context, in RegistrarEntregaInput, esc autorizacao.Escopo) (RegistrarEntregaResultado, error) {
	entregaID, err := uc.repo.BuscarEntregaPorChaveIdempotencia(ctx, in.Ator.Proprio(), RotaIdempotencia, in.IdempotencyKey)
	if err != nil {
		return RegistrarEntregaResultado{}, err
	}
	return RegistrarEntregaResultado{EntregaID: entregaID, Reaproveitada: true}, nil
}

func validarAnexos(arquivos []ArquivoRecebido) ([]anexoValidado, error) {
	if len(arquivos) == 0 {
		return nil, domain.ErrEntregaSemAnexo
	}
	if len(arquivos) > anexo.LimiteArquivosPorEntrega {
		return nil, domain.ErrAnexosAcimaDoLimite
	}
	var somaBytes int64
	validados := make([]anexoValidado, 0, len(arquivos))
	for _, arq := range arquivos {
		tamanho := int64(len(arq.Conteudo))
		if tamanho > anexo.LimiteBytesPorArquivo {
			return nil, domain.ErrAnexoAcimaDoLimite
		}
		somaBytes += tamanho
		if somaBytes > anexo.LimiteBytesPorEntrega {
			return nil, domain.ErrAnexosAcimaDoLimite
		}
		tipo, err := valueobject.DetectarTipoDeAnexo(arq.Conteudo)
		if err != nil {
			return nil, err
		}
		soma := sha256.Sum256(arq.Conteudo)
		validados = append(validados, anexoValidado{
			nomeOriginal: arq.NomeOriginal, tipo: tipo, conteudo: arq.Conteudo, hash: hex.EncodeToString(soma[:]),
		})
	}
	return validados, nil
}

func (uc *RegistrarEntregaUseCase) gravarObjetos(ctx context.Context, instituicaoID, entregaID uuid.UUID, validados []anexoValidado) ([]port.AnexoParaGravar, error) {
	gravados := make([]port.AnexoParaGravar, 0, len(validados))
	for _, a := range validados {
		chave := chaveDoObjeto(instituicaoID, entregaID, a.nomeOriginal)
		if err := uc.armazenamento.Gravar(ctx, chave, bytes.NewReader(a.conteudo), a.tipo.MimeType(), int64(len(a.conteudo))); err != nil {
			removerObjetosOrfaos(ctx, uc.armazenamento, gravados)
			return nil, err
		}
		gravados = append(gravados, port.AnexoParaGravar{
			NomeOriginal: a.nomeOriginal, Tipo: string(a.tipo), TamanhoBytes: int64(len(a.conteudo)),
			ChaveObjeto: chave, HashSHA256: a.hash,
		})
	}
	return gravados, nil
}

func chaveDoObjeto(instituicaoID, entregaID uuid.UUID, nomeOriginal string) string {
	return fmt.Sprintf("entrega-anexo/%s/%s/%s", instituicaoID, entregaID, uuid.Must(uuid.NewV7()))
}

// removerObjetosOrfaos é best-effort — nunca tratado como garantia
// (M-09). Se a remoção falhar, o objeto fica órfão: invisível, sem linha
// que o alcance, reclamado pela rotina de retenção.
func removerObjetosOrfaos(ctx context.Context, armazenamento port.ArmazenamentoDeObjetos, anexos []port.AnexoParaGravar) {
	for _, a := range anexos {
		_ = armazenamento.Remover(ctx, a.ChaveObjeto)
	}
}
