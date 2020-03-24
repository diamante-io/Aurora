package bitcoin

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

func TestAddressGenerator(t *testing.T) {
	// Generated using https://iancoleman.github.io/bip39/
	// Root key:
	// xprv9s21ZrQH143K2Cfj4mDZBcEecBmJmawReGwwoAou2zZzG45bM6cFPJSvobVTCB55L6Ld2y8RzC61CpvadeAnhws3CHsMFhNjozBKGNgucYm
	// Derivation Path m/44'/0'/0'/0:
	// xprvA1y8DJefYknMwXkdUrSk57z26Z3Fjr3rVpk8NzQKRQWjy3ogV43qr4eqTuF1rg5rrw28mqbDHfWsmoBbeDPcQ34teNgDyohSu6oyodoJ6Bu
	// xpub6ExUcpBZP8LfA1q6asykSFvkeask9Jmhs3fjBNovyk3iqr8q2bN6PryKKCvLLkMs1u2667wJnoM5LRQc3JcsGbQAhjUqJavxhtdk363GbP2
	generator, err := NewAddressGenerator("xpub6ExUcpBZP8LfA1q6asykSFvkeask9Jmhs3fjBNovyk3iqr8q2bN6PryKKCvLLkMs1u2667wJnoM5LRQc3JcsGbQAhjUqJavxhtdk363GbP2", &chaincfg.MainNetParams)
	assert.NoError(t, err)

	expectedChildren := []struct {
		index   uint32
		address string
	}{
		{0, "1Q74qRud8bXUn6FMtXWZwJa5pj56s3mdyf"},
		{1, "1CSauQLNjb3RVQN34bDZAnmKuHScsP3xuC"},
		{2, "17HCcV6BseYXaZaBXAPZqtCGQTJB9ZKsYS"},
		{3, "1MLEi1UXggrJP9ArUbxNPE9N6JUMnXErxb"},
		{4, "1cwGtdn8kqGakhXji1qDAnFjp58zN5qTn"},
		{5, "13X3CERUszAkQ2YG8yJ3eDQ8w2ATosRJWk"},
		{6, "16sgaW7RPaebPNB1umpNMxiJLjhRnNsJWY"},
		{7, "1D8xepkjsM6hfA56E1j3NWP2zcyrTMsrQM"},
		{8, "1DAEFQpKEqchA7caGKQBRacexcGJWvjXfP"},
		{9, "1N3nPpuLiZtDxuM9F3qtbTNJun3kSwC83C"},

		{100, "14C4sYrxXMCN17gUK2BMjuHSgmsp4X1oYu"},
		{101, "1G8unQbMMSrGh9SHwyUCVVGu5NTjAEPRYY"},
		{102, "1HeyVCFJr95VGJwJAuUSfBenCwk1jSjjsQ"},
		{103, "18hSmMYJ43AHrE1x5Q9gHjaEMJmbwaUQQo"},
		{104, "18sVLpqDyz4dfmBy6bwNw9yYJme8ybQxeh"},
		{105, "1EjPpuUU2Mh2vgmQgdmQvF6TqkR3YJEypn"},
		{106, "17zJ3LxbZFVpNANXeJfHvCGSsytfMYMeVh"},
		{107, "1555pj7ZWw2Qmv7chn1ziJgDYkaauw9BLD"},
		{108, "1KUaZb5Znqu8XF7AV7phhGDuVPvosJeoa"},
		{109, "144w7WJhkpm9M9k9xYxQdmyNxgPiY33L6v"},

		{1000, "1JGxd9xgBpYp4z7XHS9ezonfUTEuSoQv7y"},
	}

	for _, child := range expectedChildren {
		address, err := generator.Generate(child.index)
		assert.NoError(t, err)
		assert.Equal(t, child.address, address)
	}
}
