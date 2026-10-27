package document

import "testing"

func TestCell(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"a | b\nnext", "a \\| b<br>next"},
		{"a\r\nb\rc", "a<br>b<br>c"},
		{"  `region` <script> & \"quoted\"  ", "`region` &lt;script&gt; &amp; \"quoted\""},
	} {
		t.Run(tc.input, func(t *testing.T) {
			if got := Cell(tc.input); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	for _, tc := range []struct {
		input string
		valid bool
	}{
		{"kind: Composition\n", true},
		{"kind: Composition\n---\nkind: ConfigMap", false},
		{"kind: Composition\n---\ninvalid: [", false},
		{"kind: [", false},
		{"", false},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var target map[string]interface{}
			if err := Decode([]byte(tc.input), &target); (err == nil) != tc.valid {
				t.Fatalf("valid=%v, got error %v", tc.valid, err)
			}
		})
	}
}
