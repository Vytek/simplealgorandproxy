package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/client/kmd"
	"github.com/algorand/go-algorand-sdk/transaction"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Local KMD
const kmdAddress = "http://127.0.0.1:7833"

//const kmdToken = "<token_svilupp>"  //Sviluppo
const kmdToken = "<token_esercizio>" //Esercizio

// Purestake
const algodAddress = "https://testnet-algorand.api.purestake.io/ps1"
const algodToken = "<token_testnet>"

// To and From Addresses
const toAddr = "N2CNZ5VZD3DNMOWPN5Z3CXANEQWYZMF7YEUONO4355M4QHHBPRG3AHRASI"

//const fromAddr = "H66CAVGV64Z76KEN3TOW2KDDNFXUYQK5SYHWJWFUPFEDVABMUKLEUCPW5U"
const fromAddr = "N2CNZ5VZD3DNMOWPN5Z3CXANEQWYZMF7YEUONO4355M4QHHBPRG3AHRASI"

// Secret Algorand password
const WalletPassword = "<wallet_password>"

// WalletID
const WalletID = "<wallet_name>"

// JSON to send in transaction
type DataJSON struct {
	Type          string
	HashFile      string
	HashFilePADES string
	DateTime      string
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/notarizeservices/:hashfile/:hashfilepades/:datetime", notarize)
	e.PUT("/notarizeservices/:hashfile/:hashfilepades/:datetime", notarize)
	e.GET("/notarizeservices/:hashfile/:hashfilepades/:datetime", notarize)

	// Start server
	e.Logger.Fatal(e.Start(":1324"))
}

//https://play.golang.org/p/RnEBFCJ9h0
func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

//https://gist.github.com/is73/de4f38e1d8da157fe33e
func BytesToString(data []byte) string {
	return string(data[:])
}

// Handler
func notarize(c echo.Context) error {
	vHash := c.Param("hashfile")
	vHashPADES := c.Param("hashfilepades")
	vDateTime := c.Param("datetime")

	//Simple check data inserted
	if (len(vHash) == 64) && (len(vHashPADES) == 64) && (len(vDateTime) == 14) {
		//Notarize on Algorand
		// Create a kmd client
		kmdClient, err := kmd.MakeClient(kmdAddress, kmdToken)
		if err != nil {
			response := fmt.Sprintf("failed to make kmd client: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}
		fmt.Println("Made a kmd client")

		// uncomment if using Purestake
		var headers []*algod.Header
		headers = append(headers, &algod.Header{"X-API-Key", algodToken})
		// Create an algod client
		algodClient, err := algod.MakeClientWithHeaders(algodAddress, "", headers)
		if err != nil {
			response := fmt.Sprintf("failed to make algod client: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		fmt.Println("Made an algod client")

		// Get the list of wallets
		listResponse, err := kmdClient.ListWallets()
		if err != nil {
			response := fmt.Sprintf("error listing wallets: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		// Find our wallet name in the list
		var exampleWalletID string
		fmt.Printf("Got %d wallet(s):\n", len(listResponse.Wallets))
		for _, wallet := range listResponse.Wallets {
			fmt.Printf("ID: %s\tName: %s\n", wallet.ID, wallet.Name)
			if wallet.Name == WalletID {
				fmt.Printf("found wallet '%s' with ID: %s\n", wallet.Name, wallet.ID)
				exampleWalletID = wallet.ID
			}
		}

		// Get a wallet handle
		initResponse, err := kmdClient.InitWalletHandle(exampleWalletID, WalletPassword)
		if err != nil {
			response := fmt.Sprintf("Error initializing wallet handle: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		// Extract the wallet handle
		exampleWalletHandleToken := initResponse.WalletHandleToken

		// Get the suggested transaction parameters
		txParams, err := algodClient.SuggestedParams()
		if err != nil {
			response := fmt.Sprintf("error getting suggested tx params: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		// Create JSON for transaction
		dataBase64ToSend := DataJSON{
			Type:          "Data",
			HashFile:      vHash,
			HashFilePADES: vHashPADES,
			DateTime:      vDateTime,
		}
		dataBase64, err := json.Marshal(dataBase64ToSend)
		if err != nil {
			response := fmt.Sprintf("Error creating JSON for transaction: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		// Make transaction
		genID := txParams.GenesisID
		tx, err := transaction.MakePaymentTxn(fromAddr, toAddr, 1, 100000, txParams.LastRound, txParams.LastRound+1000, []byte(dataBase64), "", genID, txParams.GenesisHash)
		if err != nil {
			response := fmt.Sprintf("Error creating transaction: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		// Sign the transaction - change to your wallet password
		signResponse, err := kmdClient.SignTransaction(exampleWalletHandleToken, WalletPassword, tx)
		if err != nil {
			response := fmt.Sprintf("Failed to sign transaction with kmd: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		fmt.Printf("kmd made signed transaction with bytes: %x\n", signResponse.SignedTransaction)

		// Broadcast the transaction to the network
		// **** Note that this transaction will get rejected because the accounts do not have any tokens
		// **** copy off the Generated address 1 in the output below and past into the testnet dispenser
		// https://bank.testnet.algorand.network/
		txHeaders := append([]*algod.Header{}, &algod.Header{"Content-Type", "application/x-binary"})
		sendResponse, err := algodClient.SendRawTransaction(signResponse.SignedTransaction, txHeaders...)
		if err != nil {
			response := fmt.Sprintf("failed to send transaction: %s\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": response})
		}

		fmt.Printf("Transaction ID: %s\n", sendResponse.TxID)

		return c.JSON(http.StatusOK, map[string]string{
			"hash":      vHash,
			"hashpades": vHashPADES,
			"datetime":  vDateTime,
			"txid":      sendResponse.TxID})
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Inserted data was not in expeted format!"})
	}
}
