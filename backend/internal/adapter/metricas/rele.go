package metricas

// Rele implementa port.MetricasDoRele — a interface fica em internal/port
// (T-117), a implementação concreta do cliente Prometheus continua aqui,
// junto das demais.
type Rele struct{}

func NovoRele() Rele { return Rele{} }

func (Rele) RegistrarNotificacoesPendentes(n int)      { RegistrarNotificacoesPendentes(n) }
func (Rele) RegistrarNotificacaoEnviada(evento string) { RegistrarNotificacaoEnviada(evento) }
