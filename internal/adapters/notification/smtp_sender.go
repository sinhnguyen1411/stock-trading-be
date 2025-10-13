package notification

import (
    "context"
    "crypto/tls"
    "fmt"
    "net"
    "net/smtp"
    "net/url"
    "strings"
    "time"
)

// SMTPSender sends verification emails using an SMTP server.
type SMTPSender struct {
    host     string
    port     int
    username string
    password string
    from     string
    useTLS   bool
    timeout  time.Duration
    verifyURLBase string
}

type SMTPSenderConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    From     string
    UseTLS   bool
    Timeout  time.Duration
    // VerificationURLBase should end with "token=" so the token can be appended.
    // Example: http://127.0.0.1:18080/users/verify?token=
    VerificationURLBase string
}

// NewSMTPSender constructs an SMTPSender using provided configuration.
func NewSMTPSender(cfg SMTPSenderConfig) (*SMTPSender, error) {
    if cfg.Host == "" {
        return nil, fmt.Errorf("smtp host is required")
    }
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("smtp from address is required")
	}
    if cfg.Timeout == 0 {
        cfg.Timeout = 10 * time.Second
    }
    if cfg.VerificationURLBase == "" {
        // No-op; email body will still include the raw token if base URL not provided.
        // It is recommended to set this in configuration.
    }
    return &SMTPSender{
        host:     cfg.Host,
        port:     cfg.Port,
        username: cfg.Username,
        password: cfg.Password,
        from:     cfg.From,
        useTLS:   cfg.UseTLS,
        timeout:  cfg.Timeout,
        verifyURLBase: cfg.VerificationURLBase,
    }, nil
}

func (s *SMTPSender) address() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// SendVerificationEmail composes and sends a verification email.
func (s *SMTPSender) SendVerificationEmail(ctx context.Context, email, token, purpose string) error {
    subject := "Verify your account"
    if purpose != "" && purpose != "register" {
        subject = fmt.Sprintf("User verification (%s)", purpose)
    }
    var link string
    if s.verifyURLBase != "" {
        // Best effort URL composition. Expect base like: http://host/users/verify?token=
        link = s.verifyURLBase + url.QueryEscape(token)
    }
    body := fmt.Sprintf(
        "Hello,\n\nPlease verify your account using the information below:\n\nToken: %s\n%s\n\nThank you.\n",
        token,
        func() string { if link != "" { return "Link: " + link } ; return "" }(),
    )
    msg := buildMessage(s.from, email, subject, body)

	d := net.Dialer{Timeout: s.timeout}
	conn, err := d.DialContext(ctx, "tcp", s.address())
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer conn.Close()

	var client *smtp.Client
	if s.useTLS {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: s.host})
		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("tls handshake: %w", err)
		}
		client, err = smtp.NewClient(tlsConn, s.host)
	} else {
		client, err = smtp.NewClient(conn, s.host)
	}
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	helloName := s.host
	if helloName == "" {
		helloName = "localhost"
	}
	if err := client.Hello(helloName); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}

	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(email); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data open: %w", err)
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return fmt.Errorf("smtp data write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return client.Quit()
}

func buildMessage(from, to, subject, body string) string {
	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=\"UTF-8\"",
	}
	var builder strings.Builder
	for k, v := range headers {
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(v)
		builder.WriteString("\r\n")
	}
	builder.WriteString("\r\n")
	builder.WriteString(body)
	return builder.String()
}
