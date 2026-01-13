package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mdp/qrterminal/v3"
)

const baseURL = "http://localhost:8080"

type EnrollResponse struct {
	Secret        string   `json:"Secret"`
	OTPAuthURL    string   `json:"OTPAuthURL"`
	RecoveryCodes []string `json:"RecoveryCodes"`
}

type StatusResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== TOTP Backend Interactive Demo ===")
	fmt.Print("Enter a username for this session: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Println("Username cannot be empty")
		return
	}

	// 1. Enroll
	fmt.Printf("\n[1] Enrolling user '%s'...\n", username)
	resp, err := http.Post(
		baseURL+"/enroll",
		"application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"user_id": "%s"}`, username)),
	)
	if err != nil {
		fmt.Printf("Failed to connect to backend at %s/enroll: %v\nMake sure 'go run cmd/api/main.go' is running in another terminal.\n", baseURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Enrollment failed: %s\n", body)
		return
	}

	var enrollData EnrollResponse
	json.NewDecoder(resp.Body).Decode(&enrollData)

	fmt.Println(">>> ENROLLMENT SUCCESSFUL <<<")
	fmt.Printf("Secret: %s\n", enrollData.Secret)
	fmt.Println("Scan the QR code below with your Authenticator App:")
	fmt.Println("")

	// Generate QR in terminal
	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}
	qrterminal.GenerateWithConfig(enrollData.OTPAuthURL, config)

	fmt.Println("")
	fmt.Println("Recovery Codes (SAVE THESE!):")
	for _, code := range enrollData.RecoveryCodes {
		fmt.Println(" -", code)
	}

	// 2. Verify
	fmt.Println("\n[2] Verification (First Time)")
	fmt.Println("Please enter the code from your app to enable 2FA:")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	verifyResp, err := http.Post(
		baseURL+"/verify",
		"application/json",
		bytes.NewBufferString(fmt.Sprintf(`{"user_id": "%s", "code": "%s"}`, username, code)),
	)
	if err != nil {
		panic(err)
	}
	defer verifyResp.Body.Close()

	if verifyResp.StatusCode == http.StatusOK {
		fmt.Println(">>> 2FA ENABLED SUCCESSFULLY! <<<")
	} else {
		body, _ := io.ReadAll(verifyResp.Body)
		fmt.Printf("Verification failed: %s\n", body)
		return
	}

	// 3. Validation Loop
	for {
		fmt.Println("\n[3] Test Validation / Recovery")
		fmt.Println("Type a 6-digit code to validate, or a recovery code to recover.")
		fmt.Println("Type 'exit' to quit.")
		fmt.Print("Input: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		if len(input) == 6 {
			// Assume TOTP
			res, err := http.Post(
				baseURL+"/validate",
				"application/json",
				bytes.NewBufferString(fmt.Sprintf(`{"user_id": "%s", "code": "%s"}`, username, input)),
			)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			fmt.Printf("Status: %d | Body: %s\n", res.StatusCode, body)

		} else {
			// Try Recovery
			res, err := http.Post(
				baseURL+"/recover",
				"application/json",
				bytes.NewBufferString(fmt.Sprintf(`{"user_id": "%s", "code": "%s"}`, username, input)),
			)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			fmt.Printf("Status: %d | Body: %s\n", res.StatusCode, body)
		}
	}
}
