package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type GenerateRecordsResponse struct {
	DkimRecord    string `json:"dkimRecord"`
	SpfRecord     string `json:"spfRecord"`
	CnameRecord   string `json:"cnameRecord"`
	PublicKeyHash string `json:"publicKeyHash"`
}

type SendEmailRequest struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

const (
	privateKeyFile = "private.pem"
	smtpHost       = "ASPMX.L.GOOGLE.COM" // Replace with your desired SMTP server hostname/IP
	smtpPort       = "25"                 // Replace with your desired SMTP server port
	filePath       = "private.pem"
	// smtpUsername   = "your_smtp_username" // Replace with your SMTP username
	// smtpPassword   = "your_smtp_password" // Replace with your SMTP password
)

func generateRecordsHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	domain, ok := requestBody["domain"]
	if !ok {
		http.Error(w, "Missing 'domain' field in request body", http.StatusBadRequest)
		return
	}

	// Generate DKIM key pair
	privateKey, publicKey, err := generateDKIMKeyPair()
	if err != nil {
		http.Error(w, "Error generating DKIM key pair", http.StatusInternalServerError)
		return
	}

	err = savePrivateKey(privateKey)
	if err != nil {
		http.Error(w, "Error saving private key", http.StatusInternalServerError)
		return
	}

	dkimRecord := generateDKIMRecord(domain, publicKey)

	spfRecord := fmt.Sprintf("v=spf1 include:%s -all", domain)

	cnameRecord := domain

	hash := sha256.New()
	hash.Write(publicKey)
	publicKeyHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Respond with DKIM, SPF, CNAME records, and public key hash
	response := GenerateRecordsResponse{
		DkimRecord:    dkimRecord,
		SpfRecord:     spfRecord,
		CnameRecord:   cnameRecord,
		PublicKeyHash: publicKeyHash,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func verifyRecordsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Verification logic not implemented yet")
}

func getMXRecords(domain string) ([]string, error) {
	mx, err := net.LookupMX(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup MX records for %s: %w", domain, err)
	}
	mxRecords := make([]string, len(mx))
	for i, record := range mx {
		mxRecords[i] = record.Host
	}
	return mxRecords, nil
}

func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody SendEmailRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Load private key from file
	privateKey, err := loadPrivateKey(filePath)
	if err != nil {
		log.Println("Error loading private key:", err)

		http.Error(w, "Error loading private key", http.StatusInternalServerError)
		return
	}

	// Generate a unique Message-ID (e.g., using a library like uuid)
	messageID := fmt.Sprintf("<%s@%s>", generateUniqueID(), "your-domain.com") // Replace with your domain

	// Set headers including Message-ID
	headers := map[string]string{
		"From":         requestBody.From,
		"To":           requestBody.To,
		"Subject":      requestBody.Subject,
		"Content-Type": "text/plain; charset=utf-8",
		"Message-ID":   messageID,
	}

	// Construct email message
	message := "From: " + requestBody.From + "\r\n"
	message += "To: " + requestBody.To + "\r\n"
	message += "Subject: Test Email\r\n" // Add a subject line
	message += "\r\n"                    // Separate headers from body with a blank line
	message += requestBody.Body + "\r\n"

	// Build the email message with headers
	msg := ""
	for key, value := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	msg += "\r\n"  // Separate headers from body
	msg += message // Add your message body here

	// Sign email
	signedMessage, err := signEmail(requestBody.Body, privateKey)
	if err != nil {
		http.Error(w, "Error signing email", http.StatusInternalServerError)
		return
	}

	// Send email
	err = sendEmail(requestBody.To, requestBody.From, requestBody.Subject, string(signedMessage))
	if err != nil {
		// Log the exact error message for debugging
		log.Println("Error sending email:", err)

		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Email sent successfully")
}

func generateUniqueID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Handle error (e.g., log the error)
		return ""
	}
	return fmt.Sprintf("%x-%x", timestamp, randomBytes)
}

func generateDKIMKeyPair() (*rsa.PrivateKey, []byte, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1800)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	// Extract public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	return privateKey, publicKeyBytes, nil
}

func savePrivateKey(privateKey *rsa.PrivateKey) error {
	// Marshal private key to PEM format
	pemPrivateKey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write private key to file
	return ioutil.WriteFile(privateKeyFile, pem.EncodeToMemory(pemPrivateKey), 0600)
}

func loadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	// Read private key from file
	privateKeyPEM, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %w", err) // Wrap error
	}

	// Decode PEM block
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block: %w", err) // Wrap error
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err) // Wrap error
	}

	// Type assertion (ensure success before conversion)
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("parsed key is not an RSA private key: %w", err) // Wrap error
	}

	return rsaPrivateKey, nil
}

func generateDKIMRecord(domain string, publicKey []byte) string {
	return fmt.Sprintf("v=DKIM1; k=rsa; p=%s; s=email; h=sha256;", base64.StdEncoding.EncodeToString(publicKey))
}

func signEmail(message string, privateKey *rsa.PrivateKey) ([]byte, error) {
	// Compute SHA256 hash of the message
	hash := sha256.New()
	_, err := hash.Write([]byte(message))
	if err != nil {
		return nil, fmt.Errorf("failed to compute SHA256 hash: %w", err)
	}
	hashed := hash.Sum(nil)

	// Sign the hashed message
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return signature, nil
}

// Function to send an email
func sendEmail(fromEmail, toEmail, subject, message string) error {
	// Construct email message (headers and body)
	msg := "From: " + fromEmail + "\r\n"
	msg += "To: " + toEmail + "\r\n"
	msg += "Subject: " + subject + "\r\n"
	msg += "\r\n" // Separate headers from body
	msg += message + "\r\n"

	// Extract target server details from email (replace with actual logic)
	targetServer := extractTargetServer(toEmail) // Replace with actual logic

	// Connect to target SMTP server
	conn, err := net.Dial("tcp", targetServer+":25")
	if err != nil {
		return fmt.Errorf("failed to connect to target server: %w", err)
	}
	defer conn.Close() // Ensure connection is closed even in case of errors

	// Create SMTP client with potential error handling
	client, err := smtp.NewClient(conn, targetServer) // No authentication for POC
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit() // Close the connection when done with client

	// Send HELO command
	if err := client.Hello("crms-email-test.io.mslm.io"); err != nil { // Replace with your hostname if needed
		return fmt.Errorf("failed to send HELO command: %w", err)
	}

	// Send MAIL FROM command
	if err := client.Mail(fromEmail); err != nil {
		return fmt.Errorf("failed to send MAIL FROM command: %w", err)
	}

	// Send RCPT TO command for each recipient
	for _, recipient := range strings.Split(toEmail, ",") {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to send RCPT TO command for %s: %w", recipient, err)
		}
	}

	// Send DATA command and email content
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send DATA command: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close DATA writer: %w", err)
	}

	fmt.Println("Email sent successfully")
	return nil
}

// Placeholder function to extract target server from email (replace with your logic)
func extractTargetServer(toEmail string) string {
	// This is a simplified approach, a more robust solution might involve parsing actual email headers
	domain := strings.SplitN(toEmail, "@", 2)[1]
	return fmt.Sprintf("smtp.%s", domain) // Replace with actual logic to extract server from headers
}

func main() {
	// Send email with predefined parameters
	err := sendEmail("asad.mslm@outlook.com", "asad.ali@mslm.io", "Test Subject", "Test Body: This is a test email to create POC for smtp server to send")
	if err != nil {
		log.Fatal("Error sending email:", err)
	}

	// Start HTTP server
	r := mux.NewRouter()
	r.HandleFunc("/generateRecords", generateRecordsHandler).Methods("POST")
	r.HandleFunc("/verifyRecords", verifyRecordsHandler).Methods("GET")
	r.HandleFunc("/sendEmail", sendEmailHandler).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Server listening on port 8080")
	log.Fatal(srv.ListenAndServe())
}
