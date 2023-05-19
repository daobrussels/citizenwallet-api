package community

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/daobrussels/cw/pkg/common/ethrequest"
	"github.com/daobrussels/smartcontracts/pkg/contracts/accfactory"
	"github.com/daobrussels/smartcontracts/pkg/contracts/gateway"
	"github.com/daobrussels/smartcontracts/pkg/contracts/grfactory"
	"github.com/daobrussels/smartcontracts/pkg/contracts/paymaster"
	"github.com/daobrussels/smartcontracts/pkg/contracts/profactory"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Community struct {
	es      *ethrequest.EthService
	key     *ecdsa.PrivateKey
	address common.Address
	chainID *big.Int

	EntryPoint       common.Address
	Gateway          *gateway.Gateway
	Paymaster        *paymaster.Paymaster
	AccountFactory   *accfactory.Accfactory
	GratitudeFactory *grfactory.Grfactory
	ProfileFactory   *profactory.Profactory
}

func New(es *ethrequest.EthService, key *ecdsa.PrivateKey, address common.Address, chainID *big.Int, gaddr, paddr, accaddr, graddr, proaddr common.Address) (*Community, error) {
	// instantiate gateway contract
	g, err := gateway.NewGateway(gaddr, es.Client())
	if err != nil {
		return nil, err
	}

	// instantiate paymaster contract
	p, err := paymaster.NewPaymaster(paddr, es.Client())
	if err != nil {
		return nil, err
	}

	// instantiate account factory contract
	acc, err := accfactory.NewAccfactory(accaddr, es.Client())
	if err != nil {
		return nil, err
	}

	// instantiate gratitude factory contract
	gr, err := grfactory.NewGrfactory(graddr, es.Client())
	if err != nil {
		return nil, err
	}

	// instantiate profile factory contract
	pro, err := profactory.NewProfactory(proaddr, es.Client())
	if err != nil {
		return nil, err
	}

	return &Community{
		es:               es,
		key:              key,
		address:          address,
		chainID:          chainID,
		EntryPoint:       gaddr,
		Gateway:          g,
		Paymaster:        p,
		AccountFactory:   acc,
		GratitudeFactory: gr,
		ProfileFactory:   pro,
	}, nil
}

func Deploy(es *ethrequest.EthService, key *ecdsa.PrivateKey, address common.Address, chainID *big.Int) (*Community, error) {
	c := &Community{
		es:      es,
		key:     key,
		address: address,
		chainID: chainID,
	}

	// instantiate gateway contract
	err := c.DeployGateway()
	if err != nil {
		return nil, err
	}

	// deploy paymaster contract
	err = c.DeployPaymaster()
	if err != nil {
		return nil, err
	}

	// deploy account factory contract
	err = c.DeployAccountFactory()
	if err != nil {
		return nil, err
	}

	// deploy gratitude factory contract
	err = c.DeployGratitudeFactory()
	if err != nil {
		return nil, err
	}

	// deploy profile factory contract
	err = c.DeployProfileFactory()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Community) NewTransactor() (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(c.key, c.chainID)
}

func (c *Community) NextNonce() (uint64, error) {
	return c.es.NextNonce(c.address.Hex())
}

func (c *Community) DeployGateway() error {
	auth, err := c.NewTransactor()
	if err != nil {
		return err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	// deploy the gateway contract
	addr, _, g, err := gateway.DeployGateway(auth, c.es.Client())
	if err != nil {
		return err
	}

	c.EntryPoint = addr
	c.Gateway = g

	return nil
}

func (c *Community) DeployPaymaster() error {
	auth, err := c.NewTransactor()
	if err != nil {
		return err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	// deploy the paymaster contract
	_, _, p, err := paymaster.DeployPaymaster(auth, c.es.Client(), c.EntryPoint)
	if err != nil {
		return err
	}

	c.Paymaster = p

	return nil
}

func (c *Community) FundPaymaster(amount *big.Int) error {
	auth, err := c.NewTransactor()
	if err != nil {
		return err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = amount

	_, err = c.Paymaster.Deposit(auth)
	if err != nil {
		return err
	}

	return nil
}

func (c *Community) DeployAccountFactory() error {
	auth, err := c.NewTransactor()
	if err != nil {
		return err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	// deploy the account factory contract
	_, _, acc, err := accfactory.DeployAccfactory(auth, c.es.Client(), c.EntryPoint)
	if err != nil {
		return err
	}

	c.AccountFactory = acc

	return nil
}

func (c *Community) DeployGratitudeFactory() error {
	auth, err := c.NewTransactor()
	if err != nil {
		return err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	// deploy the gratitude factory contract
	_, _, gr, err := grfactory.DeployGrfactory(auth, c.es.Client(), c.EntryPoint)
	if err != nil {
		return err
	}

	c.GratitudeFactory = gr

	return nil
}

func (c *Community) CreateGratitudeApp(owner common.Address) (*common.Address, error) {
	auth, err := c.NewTransactor()
	if err != nil {
		return nil, err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return nil, err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	// create the gratitude app
	_, err = c.GratitudeFactory.CreateGratitudeToken(auth, owner, big.NewInt(int64(nonce)))
	if err != nil {
		return nil, err
	}

	addr, err := c.GratitudeFactory.GetGratitudeTokenAddress(&bind.CallOpts{}, owner, big.NewInt(int64(nonce)))
	if err != nil {
		return nil, err
	}

	return &addr, nil
}

func (c *Community) CreateAccount(owner common.Address) (*common.Address, error) {
	auth, err := c.NewTransactor()
	if err != nil {
		return nil, err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return nil, err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	_, err = c.AccountFactory.CreateAccount(auth, owner, big.NewInt(int64(nonce)))
	if err != nil {
		return nil, err
	}

	addr, err := c.AccountFactory.GetAddress(&bind.CallOpts{}, owner, big.NewInt(int64(nonce)))
	if err != nil {
		return nil, err
	}

	return &addr, nil
}

func (c *Community) DeployProfileFactory() error {
	auth, err := c.NewTransactor()
	if err != nil {
		return err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	// deploy profile factory contract
	_, _, pr, err := profactory.DeployProfactory(auth, c.es.Client(), c.EntryPoint)
	if err != nil {
		return err
	}

	c.ProfileFactory = pr

	return nil
}

func (c *Community) CreateProfile(owner common.Address) (*common.Address, error) {
	auth, err := c.NewTransactor()
	if err != nil {
		return nil, err
	}

	// get the next nonce for the main wallet
	nonce, err := c.NextNonce()
	if err != nil {
		return nil, err
	}

	// set default parameters
	setDefaultParameters(auth, nonce)

	_, err = c.ProfileFactory.CreateProfile(auth, owner, big.NewInt(int64(nonce)))
	if err != nil {
		return nil, err
	}

	addr, err := c.ProfileFactory.GetProfileAddress(&bind.CallOpts{}, owner, big.NewInt(int64(nonce)))
	if err != nil {
		return nil, err
	}

	return &addr, nil
}

func setDefaultParameters(auth *bind.TransactOpts, nonce uint64) {
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(30000000 - 1)
}
