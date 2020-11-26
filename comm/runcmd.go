package comm

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/rpc/client/http"
	dbm "github.com/tendermint/tm-db"
	"strconv"
	"ubc/lib"
	px "ubc/paramer"
	"ubc/tx"
)

var _ types.Application = (*lib.TokenApp)(nil)

var (
	app_cli, _ = http.New("http://localhost:26657", "/websocket")
)

type App struct {
	types.BaseApplication
	Value        int64
	Version      int64
	state        State
	RetainBlocks int64
}

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

type UApp struct {
	app *App

	ValUpdates []types.ValidatorUpdate

	valAddrToPubKeyMap map[string]types.PubKey

	logger log.Logger
}

func Run() {

	walletCmd := &cobra.Command{
		Use:   "init_wallet",
		Short: "Initialize Wallet",
		Run:   func(cmd *cobra.Command, args []string) { px.InitWallet() },
	}

	createAccountCmd := &cobra.Command{
		Use:   "create_account",
		Short: "Create one Account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("Please input account name!")
			}
			label := args[0]
			lib.CreateAccount(label)
			fmt.Println("Account create success!")
			return nil
		},
	}

	issueCmd := &cobra.Command{
		Use:   "issue_tx",
		Short: "Issue coins, Please use [cli issue_tx user]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("Issue coin error! Please use [cli user]")
			}
			val, _ := strconv.Atoi(args[1])
			err := tx.Issue(app_cli, args[0], val)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Issue Success!")
			return nil
		},
	}

	transferCmd := &cobra.Command{
		Use:   "transfer_tx",
		Short: "Transaction detail",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("Issue coin error!")
			}
			err := tx.Transfer(app_cli, args[0], args[1], args[2])
			if err != nil {
				fmt.Println(err)
			}
			return nil
		},
	}

	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Show info",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("query who?")
			}
			label := args[0]
			tx.Query(label, app_cli)
			return nil
		},
	}

	querytxCmd := &cobra.Command{
		Use:   "tx",
		Short: "Add tx info",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("Add tx  error!")
			}
			err := tx.Addtx(app_cli, args[0])
			if err != nil {
				fmt.Println(err)
			}
			return nil
		},
	}

	runAppCmd := &cobra.Command{
		Use:   "run_app",
		Short: "Run a ABCI Application",
		Run:   func(cmd *cobra.Command, args []string) { RunApp() },
	}

	root := commands.RootCmd
	root.AddCommand(commands.GenNodeKeyCmd)
	root.AddCommand(commands.GenValidatorCmd)
	root.AddCommand(commands.InitFilesCmd)
	root.AddCommand(commands.ResetAllCmd)
	root.AddCommand(commands.ShowNodeIDCmd)
	//root.AddCommand(commands.TestnetFilesCmd)

	app := lib.NewTokenApp(lib.GetDbDir())
	nodeProvider := makeNodeProvider(app)
	root.AddCommand(commands.NewRunNodeCmd(nodeProvider))

	root.AddCommand(walletCmd)
	root.AddCommand(issueCmd)
	root.AddCommand(transferCmd)
	root.AddCommand(queryCmd)
	root.AddCommand(runAppCmd)
	root.AddCommand(createAccountCmd)
	root.AddCommand(querytxCmd)

	exec := cli.PrepareBaseCmd(root, "wiz", ".")
	_ = exec.Execute()
}

func RunApp() {
	app := lib.NewTokenApp(lib.GetDbDir())
	svr, err := server.NewServer(":26658", "socket", app)
	if err != nil {
		panic(err)
	}

	_ = svr.Start()
	defer func() { _ = svr.Stop() }()

	fmt.Println("Token Server started!")

	select {}

}

//func NewApp() *App {
//	return &App{}
//}

func makeNodeProvider(app types.Application) node.Provider {
	return func(config *cfg.Config, logger log.Logger) (*node.Node, error) {
		nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
		if err != nil {
			return nil, err
		}

		return node.NewNode(config,
			privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
			nodeKey,
			proxy.NewLocalClientCreator(app),
			node.DefaultGenesisDocProviderFunc(config),
			node.DefaultDBProvider,
			node.DefaultMetricsProvider(config.Instrumentation),
			logger,
		)
	}

}
