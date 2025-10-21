package apple

import "github.com/alvinbaena/passkit"

type capturingSigner struct {
	pass      *passkit.Pass
	template  passkit.PassTemplate
	signing   *passkit.SigningInformation
	payload   []byte
	callCount int
}

func (c *capturingSigner) CreateSignedAndZippedPassArchive(pass *passkit.Pass, template passkit.PassTemplate, info *passkit.SigningInformation) ([]byte, error) {
	c.pass = pass
	c.template = template
	c.signing = info
	c.callCount++
	c.payload = []byte("signed-pass-" + pass.SerialNumber)
	return c.payload, nil
}
