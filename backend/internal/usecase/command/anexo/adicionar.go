package anexo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/anexo"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// ArquivoRecebido — mesma forma de usecase/command/entrega.ArquivoRecebido,
// duplicada de propósito: os dois pacotes não devem depender um do outro
// só para compartilhar um struct de duas linhas.
type ArquivoRecebido struct {
	NomeOriginal string
	Conteudo     []byte
}

type AdicionarAnexoInput struct {
	Ator      autorizacao.Ator
	EntregaID uuid.UUID
	Arquivos  []ArquivoRecebido
}

// AdicionarAnexoUseCase — POST /entregas/{id}/anexos, SEM Idempotency-Key
// (M-06, design.md §7.3): a duplicata aqui não corrompe número nenhum,
// porque nunca se contam arquivos. Risco residual formalmente aceito.
type AdicionarAnexoUseCase struct {
	repo          port.EntregaRepository
	armazenamento port.ArmazenamentoDeObjetos
}

func NovoAdicionarAnexoUseCase(repo port.EntregaRepository, armazenamento port.ArmazenamentoDeObjetos) *AdicionarAnexoUseCase {
	return &AdicionarAnexoUseCase{repo: repo, armazenamento: armazenamento}
}

func (uc *AdicionarAnexoUseCase) Executar(ctx context.Context, in AdicionarAnexoInput) ([]port.AnexoResponse, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoEditar, nil)
	if err != nil {
		return nil, err
	}
	detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
	if err != nil {
		return nil, err
	}
	if len(in.Arquivos) == 0 {
		return nil, domain.ErrEntregaSemAnexo
	}

	somaAtual := int64(0)
	for _, existente := range detalhe.Anexos {
		somaAtual += existente.TamanhoBytes
	}
	if len(detalhe.Anexos)+len(in.Arquivos) > anexo.LimiteArquivosPorEntrega {
		return nil, domain.ErrAnexosAcimaDoLimite
	}

	prontos := make([]port.AnexoParaGravar, 0, len(in.Arquivos))
	for _, arq := range in.Arquivos {
		tamanho := int64(len(arq.Conteudo))
		if tamanho > anexo.LimiteBytesPorArquivo {
			return nil, domain.ErrAnexoAcimaDoLimite
		}
		somaAtual += tamanho
		if somaAtual > anexo.LimiteBytesPorEntrega {
			return nil, domain.ErrAnexosAcimaDoLimite
		}
		tipo, err := valueobject.DetectarTipoDeAnexo(arq.Conteudo)
		if err != nil {
			return nil, err
		}
		soma := sha256.Sum256(arq.Conteudo)
		chave := fmt.Sprintf("entrega-anexo/%s/%s", in.EntregaID, uuid.Must(uuid.NewV7()))
		if err := uc.armazenamento.Gravar(ctx, chave, bytes.NewReader(arq.Conteudo), tipo.MimeType(), tamanho); err != nil {
			return nil, err
		}
		prontos = append(prontos, port.AnexoParaGravar{
			NomeOriginal: arq.NomeOriginal, Tipo: string(tipo), TamanhoBytes: tamanho,
			ChaveObjeto: chave, HashSHA256: hex.EncodeToString(soma[:]),
		})
	}

	return uc.repo.AdicionarAnexos(ctx, esc, in.EntregaID, prontos)
}
