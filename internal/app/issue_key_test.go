package app

import "testing"

func TestParseIssueKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "standard key",
			input: "COMMUNITY-102",
			want:  "COMMUNITY-102",
		},
		{
			name:  "lowercase normalized",
			input: "community-102",
			want:  "COMMUNITY-102",
		},
		{
			name:  "trimmed",
			input: "  COMMUNITY-102  ",
			want:  "COMMUNITY-102",
		},
		{
			name:    "missing number",
			input:   "COMMUNITY",
			wantErr: true,
		},
		{
			name:    "missing hyphen",
			input:   "COMMUNITY102",
			wantErr: true,
		},
		{
			name:    "extra suffix",
			input:   "COMMUNITY-102-extra",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseIssueKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("unexpected issue key: %s", got)
			}
		})
	}
}
