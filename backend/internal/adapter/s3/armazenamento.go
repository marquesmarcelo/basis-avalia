package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Armazenamento — adapter de port.ArmazenamentoDeObjetos sobre a SDK da
// AWS, compatível com CloudServer/MinIO (fundacao-metas.md §9). Path-style
// forçado: os dois servidores S3-compatíveis do projeto não resolvem
// virtual-hosted-style (bucket.host) sem DNS wildcard.
type Armazenamento struct {
	cliente *s3.Client
	bucket  string
}

func Novo(endpoint, regiao, accessKey, secretKey, bucket string) *Armazenamento {
	cfg := aws.Config{
		Region:      regiao,
		Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
	}
	cliente := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
	return &Armazenamento{cliente: cliente, bucket: bucket}
}

func (a *Armazenamento) Gravar(ctx context.Context, chave string, conteudo io.Reader, tipo string, tamanho int64) error {
	_, err := a.cliente.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucket), Key: aws.String(chave), Body: conteudo,
		ContentType: aws.String(tipo), ContentLength: aws.Int64(tamanho),
	})
	return err
}

func (a *Armazenamento) Ler(ctx context.Context, chave string) (io.ReadCloser, string, error) {
	saida, err := a.cliente.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(a.bucket), Key: aws.String(chave)})
	if err != nil {
		return nil, "", err
	}
	tipo := ""
	if saida.ContentType != nil {
		tipo = *saida.ContentType
	}
	return saida.Body, tipo, nil
}

func (a *Armazenamento) Remover(ctx context.Context, chave string) error {
	_, err := a.cliente.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(a.bucket), Key: aws.String(chave)})
	return err
}

// VerificarDisponibilidade sustenta /readyz (design.md §9) — só confirma
// que o bucket responde, nunca devolve o erro do driver ao cliente HTTP
// (isso é feito pelo chamador, que só loga).
func (a *Armazenamento) VerificarDisponibilidade(ctx context.Context) error {
	_, err := a.cliente.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(a.bucket)})
	return err
}
