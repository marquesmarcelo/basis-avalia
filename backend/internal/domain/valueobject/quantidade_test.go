package valueobject

import "testing"

// TestNovaQuantidade_IT04_InvalidaAbaixoDeUm cobre IT-04 (0, -1).
func TestNovaQuantidade_IT04_InvalidaAbaixoDeUm(t *testing.T) {
	for _, bruta := range []int{0, -1} {
		if _, err := NovaQuantidade(bruta); err == nil {
			t.Fatalf("quantidade %d deveria ser inválida", bruta)
		}
	}
	q, err := NovaQuantidade(4)
	if err != nil || q.Int() != 4 {
		t.Fatalf("quantidade 4 deveria ser válida, obtido q=%v err=%v", q, err)
	}
}
