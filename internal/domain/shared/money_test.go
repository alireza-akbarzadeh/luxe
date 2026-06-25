package shared

import "testing"

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name    string
		major   float64
		cur     string
		wantErr bool
		want    int64
	}{
		{name: "usd", major: 19.99, cur: "usd", want: 1999},
		{name: "negative", major: -1, cur: "USD", wantErr: true},
		{name: "bad currency", major: 10, cur: "US", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMoney(tt.major, tt.cur)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Cents() != tt.want {
				t.Fatalf("cents = %d, want %d", m.Cents(), tt.want)
			}
		})
	}
}

func TestMoneyAdd(t *testing.T) {
	a, _ := NewMoneyFromCents(100, "USD")
	b, _ := NewMoneyFromCents(250, "USD")
	sum, err := a.Add(b)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Cents() != 350 {
		t.Fatalf("got %d", sum.Cents())
	}
}
