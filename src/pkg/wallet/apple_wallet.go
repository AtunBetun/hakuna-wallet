package wallet

import (
	"context"
	"os"

	"github.com/alvinbaena/passkit"
	"github.com/atunbetun/hakuna-wallet/pkg/logger"
)

func CreateAppleWalletTicket(ctx context.Context, folderPath string) ([]byte, error) {
	template := passkit.NewFolderPassTemplate(folderPath)
	pass := passkit.Pass{}
	signer := passkit.NewMemoryBasedSigner()
	signInfo, err := passkit.LoadSigningInformationFromFiles("/home/user/pass_cert.p12", "pass_cert_password", "/home/user/AppleWWDRCA.cer")
	if err != nil {
		panic(err)
	}

	z, err := signer.CreateSignedAndZippedPassArchive(&pass, template, signInfo)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("/home/user/pass.pkpass", z, 0644)
	if err != nil {
		panic(err)
	}

	logger.Logger.Info("Pass signed successfully")
	return nil, nil
}
