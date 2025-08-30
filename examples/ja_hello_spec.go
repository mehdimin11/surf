package main

import (
	"log"

	"github.com/enetx/surf"
	tls "github.com/refraction-networking/utls"
	"github.com/refraction-networking/utls/dicttls"
)

func main() {
	spec := tls.ClientHelloSpec{
		TLSVersMin: tls.VersionTLS12,
		TLSVersMax: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		CompressionMethods: []uint8{
			0x0, // no compression
		},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
			&tls.ExtendedMasterSecretExtension{},
			&tls.RenegotiationInfoExtension{
				Renegotiation: tls.RenegotiateOnceAsClient,
			},
			&tls.SupportedCurvesExtension{
				Curves: []tls.CurveID{
					tls.X25519,
					tls.CurveP256,
					tls.CurveP384,
					tls.CurveP521,
					256,
					257,
				},
			},
			&tls.SupportedPointsExtension{
				SupportedPoints: []uint8{
					0x0, // uncompressed
				},
			},
			&tls.SessionTicketExtension{},
			&tls.ALPNExtension{
				AlpnProtocols: []string{
					"h2",
					"http/1.1",
				},
			},
			&tls.StatusRequestExtension{},
			&tls.FakeDelegatedCredentialsExtension{
				SupportedSignatureAlgorithms: []tls.SignatureScheme{
					tls.ECDSAWithP256AndSHA256,
					tls.ECDSAWithP384AndSHA384,
					tls.ECDSAWithP521AndSHA512,
					tls.ECDSAWithSHA1,
				},
			},
			&tls.KeyShareExtension{
				KeyShares: []tls.KeyShare{
					{
						Group: tls.X25519,
					},
					{
						Group: tls.CurveP256,
					},
				},
			},
			&tls.SupportedVersionsExtension{
				Versions: []uint16{
					tls.VersionTLS13,
					tls.VersionTLS12,
				},
			},
			&tls.SignatureAlgorithmsExtension{
				SupportedSignatureAlgorithms: []tls.SignatureScheme{
					tls.ECDSAWithP256AndSHA256,
					tls.ECDSAWithP384AndSHA384,
					tls.ECDSAWithP521AndSHA512,
					tls.PSSWithSHA256,
					tls.PSSWithSHA384,
					tls.PSSWithSHA512,
					tls.PKCS1WithSHA256,
					tls.PKCS1WithSHA384,
					tls.PKCS1WithSHA512,
					tls.ECDSAWithSHA1,
					tls.PKCS1WithSHA1,
				},
			},
			&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
				tls.PskModeDHE,
			}},
			&tls.FakeRecordSizeLimitExtension{
				Limit: 0x4001,
			},
			&tls.GREASEEncryptedClientHelloExtension{
				CandidateCipherSuites: []tls.HPKESymmetricCipherSuite{
					{
						KdfId:  dicttls.HKDF_SHA256,
						AeadId: dicttls.AEAD_AES_128_GCM,
					},
					{
						KdfId:  dicttls.HKDF_SHA256,
						AeadId: dicttls.AEAD_CHACHA20_POLY1305,
					},
				},
				CandidatePayloadLens: []uint16{223},
			},
		},
	}

	r := surf.NewClient().
		Builder().
		JA3().
		SetHelloSpec(spec).
		Build().
		// Get("https://tls.peet.ws/api/clean").
		Get("https://tls.peet.ws/api/all").
		Do()

	if r.IsErr() {
		log.Fatal(r.Err())
	}

	// "ja3_hash": "b5001237acdf006056b409cc433726b0"
	// "ja4": "t13d1715h2_5b57614c22b0_2764158f9823"
	// "peetprint_hash": "b9c611f928c8c1f20c414a48c66abf27"
	r.Ok().Body.String().Print()
}
