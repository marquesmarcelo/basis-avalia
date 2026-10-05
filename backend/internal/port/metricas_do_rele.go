package port

// MetricasDoRele — as duas métricas que o relê em segundo plano precisa
// emitir e que NENHUMA rota HTTP produz (T-117, design.md de
// autenticacao-usuarios §4.2/§15.1): o relê não tem rota, então o
// middleware global (contador por rota+método+status) não gera nada sobre
// ele. O gauge de pendentes é o sinal de "relê parado" — o mesmo papel
// que o CLAUDE.md dá ao backlog de outbox.
//
// Critério que separa este caso do que foi removido em T-117 (as quatro
// métricas de indicador_plataforma, que SAÍRAM em vez de virar port):
// antes de criar métrica de negócio, perguntar se o middleware global já
// a produz. Se a ação tem rota 1:1, ele produz — métrica própria nesse
// caso diverge da rota no dia em que alguém esquecer de chamá-la num
// caminho novo. Métrica própria só se justifica para o que NÃO passa por
// rota: fila, lote, backlog — exatamente o caso do relê.
type MetricasDoRele interface {
	RegistrarNotificacoesPendentes(n int)
	RegistrarNotificacaoEnviada(evento string)
}
