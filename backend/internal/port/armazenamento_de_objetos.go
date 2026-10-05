package port

import (
	"context"
	"io"
)

// ArmazenamentoDeObjetos — S3/MinIO (fundacao-metas.md §9). Nenhum método
// recebe uuid.UUID: a chave de objeto é sempre construída no adapter a
// partir de identificadores que já passaram por um repositório com
// Escopo — por isso esta porta nunca precisa de autorizacao.Escopo, e
// TestNenhumaPortaNovaSemEscopo continua verde sem ampliar a lista das
// duas portas estreitas.
type ArmazenamentoDeObjetos interface {
	Gravar(ctx context.Context, chave string, conteudo io.Reader, tipo string, tamanho int64) error
	Ler(ctx context.Context, chave string) (io.ReadCloser, string, error)
	Remover(ctx context.Context, chave string) error
}
