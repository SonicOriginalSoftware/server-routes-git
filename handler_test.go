package git_test

import (
	"context"
	"crypto/tls"
	"testing"

	lib "git.nathanblair.rocks/server"

	"git.nathanblair.rocks/routes/git"
)

var cert = []byte(`-----BEGIN CERTIFICATE-----
MIIC+TCCAeGgAwIBAgIQfiWczLKRh0VC9Ylb9jG3XjANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTIyMDIyMDA0Mzc1MVoXDTIzMDIyMDA0Mzc1
MVowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBALJr4P72n2ucixAi2VM2+i/cSr9g85iZSuDTYfoPWsbBgoXfYuH3hdl+
ihz2WGEgNXT6NZkN5C9ro/cEJilO++wdtHKmdvVU9WNgu4FFT3lVs6/Eoi751jTX
R/OllBfeWIYhIAVVpL9HYCvEc0NWi0qP0krGy1fNzEzVolkdNplUX/5EAQ8Y41Qf
1nnY0IePLo9KYC8BH3YqH5K8CjqhLsVEazn/Jv6TIkzP9aaaw5HAfMq5JaAhzmX9
CK66q3NksdIQOkXdZeO1s0dVsYRaWtH1XGjrdSOmo+Wg3Ok8o5EKKc7B1eunkNAI
RNCmK8IUxuBi5xikpLFWQ2MvGyxzdJsCAwEAAaNLMEkwDgYDVR0PAQH/BAQDAgWg
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFAYDVR0RBA0wC4IJ
bG9jYWxob3N0MA0GCSqGSIb3DQEBCwUAA4IBAQBPwTN9eRx1hybciJsRLdXzN7wb
s9RrvyXzQxz0Zgq3ykA0THOhncKLG+ZX5/3S3zx4wTMaigrL0g607r8n2Ki/Dsyg
HR7L25tTgyiQ/rH+RghGdsMvI3vYOGWpYa94PQQWWIXv+oZButDP9+7kz+tOp4WY
3JGg73js37SSkza7V2XM5puDZPgs6J64CeX82grjOf80ehtSA0eagUKUdwqpqox/
f++oYhpCSej9QDlXcUkVU5jLjkVhyFVTysJj5iv5cGaXe72CvGGK9SKSCV2aTSMa
Shi4KcroO5cYKtMQLFSZZELf3cQ7RHqLCVqwTLYsIB67dtsBcSAiIiEnI/O7
-----END CERTIFICATE-----`)

var key = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCya+D+9p9rnIsQ
ItlTNvov3Eq/YPOYmUrg02H6D1rGwYKF32Lh94XZfooc9lhhIDV0+jWZDeQva6P3
BCYpTvvsHbRypnb1VPVjYLuBRU95VbOvxKIu+dY010fzpZQX3liGISAFVaS/R2Ar
xHNDVotKj9JKxstXzcxM1aJZHTaZVF/+RAEPGONUH9Z52NCHjy6PSmAvAR92Kh+S
vAo6oS7FRGs5/yb+kyJMz/WmmsORwHzKuSWgIc5l/QiuuqtzZLHSEDpF3WXjtbNH
VbGEWlrR9Vxo63UjpqPloNzpPKORCinOwdXrp5DQCETQpivCFMbgYucYpKSxVkNj
Lxssc3SbAgMBAAECggEBAJ9N6QddBjxb+mI+61H8bWfcRjUYCIfAnvWcZweRPBjo
YHTNXompqq3l6MUxQvn9ex1l5gMxPwMEFdMQtk39wrswTpRrgx1LbQn9LY2fZ/VL
CvOuGqzcz2BAs3Kc7VxeXyXrX57DuCQ9Q2XwsoV2OUoqnbW8R4SvMHGN8bWlesFs
8YbgQut0bXLfnqEUQIztyLw53al5Ko6MH/FOoVF1/sIYIlVr+g2J42fcD85hYFTp
mhMAtkDNBFxo8XwyM7qOoQxJufeyGhehFbL9QQzrpwVed3e877p6j+rufzwlfF28
uGm1AtlkJmyWwlnfBzsx8944OqXuvK5PwalmRUlm21ECgYEA3wE9tVLo4RGC19ql
dpeGRqhkLPxnbdZvW4VBleRbGSe1XB5fFiohfMdXIDvo9qh4crhAp+jYJkpEBC84
QR372VoHISfCZOQU1AXm0MT26VKWCWj9cbNyEZDfEZD6WY9CPaEz5F3pqMlIrCbm
r2y5ZXoo7BF3FsOQzjQatqMaE7MCgYEAzNHzJcifvbL4aCxUAM94eQysm9Kxi8Fy
22iaAhkWuYSgYKeMPaPi+/MVkU0h2TQemZOVxGC7i2lBExIHG5pnv8tnz2pXzs5T
jD/ZfsUZW7bvuXOKUcmTU6vCf3na+hNhh9mMnHRv5/8zaiDeDHF02MTcS8nxy1k0
eN3u9kjwx3kCgYA0qgFdse/HPzBsM3mB8TTHuPq3WQBAAzUXIvgjIuOUpDkDQTTp
cheodRcRSLSyk4Smavbx8F4jZMR9TH13e1I/uTAX12DkHK0CiUZCJVG+Nj+yhzXb
RSp6FYFoj5lfzyIwlcJAeyE0OBzOcv1ljkKWQWwqm9FI8fRfjhSE7y24WwKBgFsu
vIh9oF/bZSs7UMprkr6RHebhDZmiLXfwQV/du3gryxo8fPqUE2EG/vsI06DWyyij
w3EBf3y5BvdudyuaucVw0G5OcXjn8dnmMvV02a0y69Yr0dBHZQdC1/vYS9w49Jp+
B1M/ovItcr40k6YGfHZkbY5wAOz+cZW9d7y9DDRBAoGBANRW3JXy47p5zPkw2WpD
Q6UQzTZDC3pD15hwr0iV6YO1kkbaeOZSEW0RC1I2uqmzal1o/TuDavB/L3KnRfd9
7TsTx0CtZtO1Ycfhn3f1SmnUxEa2De9JJH/i1fYL0rbKN7s9WPXNSh+lVhTuGoKa
lSLm75AMeTQghz1CfVEJnXc9
-----END PRIVATE KEY-----`)

func TestHandler(t *testing.T) {
	subdomains := []lib.SubdomainHandler{
		git.New(),
	}

	cert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.FailNow()
	}
	certs := []tls.Certificate{cert}

	ctx, cancelCtx := context.WithCancel(context.Background())

	// TODO Use a channel and have the Run loop execute in a goroutine
	// Wait for a brief period, send a request to the server,
	// check the response (should be a not implemented response),
	// then cancel the context and make sure the server shuts down
	// successfully

	if exitCode := lib.Run(ctx, subdomains, certs); exitCode != 0 {
		t.FailNow()
	}

	cancelCtx()
}
