package launcher

import (
	"github.com/pkg/errors"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
)

var validatorIDFlag = cli.UintFlag{
	Name:  "validator.id",
	Usage: "ID of a validator to create events from",
	Value: 0,
}

var validatorPubkeyFlag = cli.StringFlag{
	Name:  "validator.pubkey",
	Usage: "Public key of a validator to create events from",
	Value: "",
}

var validatorPasswordFlag = cli.StringFlag{
	Name:  "validator.password",
	Usage: "Password to unlock validator private key",
	Value: "",
}

// setValidatorID retrieves the validator ID either from the directly specified
// command line flags or from the keystore if CLI indexed.
func setValidator(ctx *cli.Context, cfg *emitter.Config) error {
	// Extract the current validator address, new flag overriding legacy one
	var err error
	if ctx.GlobalIsSet(FakeNetFlag.Name) {
		var num idx.Validator
		cfg.Validator.ID, num, err = parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
		if err != nil {
			return err
		}
		validators := makegenesis.GetFakeValidators(num)
		cfg.Validator.PubKey = validators.Map()[cfg.Validator.ID].PubKey
	}
	if ctx.GlobalIsSet(validatorIDFlag.Name) {
		cfg.Validator.ID = idx.ValidatorID(ctx.GlobalInt(validatorIDFlag.Name))
	}
	if ctx.GlobalIsSet(validatorPubkeyFlag.Name) {
		cfg.Validator.PubKey, err = validatorpk.FromString(ctx.GlobalString(validatorPubkeyFlag.Name))
		if err != nil {
			return err
		}
	}

	if cfg.Validator.ID != 0 && cfg.Validator.PubKey.Empty() {
		return errors.New("validator public key is not set")
	}
	return nil
}
