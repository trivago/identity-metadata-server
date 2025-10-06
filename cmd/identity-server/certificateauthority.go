package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"identity-metadata-server/internal/shared"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	certificateAuthorityScope = "https://www.googleapis.com/auth/cloud-platform"
)

// https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools.certificateAuthorities#CertificateAuthorityData
type CertificateAuthorityData struct {
	Name       string `json:"name"`
	AccessURLs struct {
		RootCertificate string   `json:"caCertificateAccessUrl"`
		RevocationLists []string `json:"crlAccessUrls"`
	} `json:"accessUrls"`
	RootCertificatePEMs []string `json:"pemCaCertificates"`
}

// https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools/fetchCaCerts#response-body
type FetchCACertsResponse struct {
	CACerts []struct {
		Certificates []string `json:"certificates"`
	} `json:"caCerts"`
}

// InitClientRootCA initializes the client root CA pool by fetching CA
// certificates from Google Cloud's Certificate Authority service.
// See https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools/fetchCaCerts
func InitClientRootCA(project, region, poolName string) ([]*x509.Certificate, *x509.CertPool, error) {
	bearerToken, err := GetIdentityServerToken([]string{certificateAuthorityScope}, context.Background())
	if err != nil {
		return nil, nil, err
	}
	url := fmt.Sprintf("https://privateca.googleapis.com/v1/projects/%s/locations/%s/caPools/%s:fetchCaCerts", project, region, poolName)

	response, err := shared.HttpPOSTJson[FetchCACertsResponse](url, nil, map[string]string{
		"Authorization": "Bearer " + bearerToken,
	}, nil, context.Background())

	if err != nil {
		return nil, nil, errors.Join(err, errors.New("failed to call fetchCaCerts"))
	}

	clientRootCAs := make([]*x509.Certificate, 0, len(response.CACerts))
	clientRootCaPool := x509.NewCertPool()
	for _, chain := range response.CACerts {
		for _, pemData := range chain.Certificates {
			// Decode the PEM block and parse the contained certificate (DER encoded)
			for certData, remain := pem.Decode([]byte(pemData)); certData != nil; certData, remain = pem.Decode(remain) {
				certificate, err := x509.ParseCertificate(certData.Bytes)
				if err != nil {
					log.Warn().Msgf("failed to parse CA cert: %s", err)
					continue
				}

				clientRootCAs = append(clientRootCAs, certificate)
				clientRootCaPool.AddCert(certificate)
			}
		}
	}

	if len(clientRootCAs) == 0 {
		return nil, nil, errors.New("no CA certs found or all failed to parse")
	}

	return clientRootCAs, clientRootCaPool, nil
}

// GetCertificateAuthority retrieves the certificate authority description from the given CA pool.
// See https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools.certificateAuthorities/get
func GetCertificateAuthority(project, region, poolName, name string, ctx context.Context) (*CertificateAuthorityData, error) {
	bearerToken, err := GetIdentityServerToken([]string{certificateAuthorityScope}, ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://privateca.googleapis.com/v1/projects/%s/locations/%s/caPools/%s/certificateAuthorities/%s",
		project, region, poolName, name)

	return shared.HttpGETJson[CertificateAuthorityData](url, nil, map[string]string{
		"Authorization": "Bearer " + bearerToken,
	}, nil, ctx)
}

// GetRevokedCertificates retrieves the revoked certificates from all
// certificate authorities in the given CA pool.
// It returns a map of serial numbers of revoked certificates.
// The keys of the map are the serial numbers, and the values are empty structs.
func GetRevokedCertificates(project, region, poolName, name string, ctx context.Context) ([]x509.RevocationList, error) {
	authority, err := GetCertificateAuthority(project, region, poolName, name, ctx)
	if err != nil {
		return nil, err
	}

	lists := make([]x509.RevocationList, 0)
	processErrors := error(nil)

	for _, crlURL := range authority.AccessURLs.RevocationLists {
		rsp, err := shared.HttpGET(crlURL, nil, map[string]string{}, nil, ctx)
		if err != nil {
			processErrors = errors.Join(processErrors, err)
			continue
		}

		body, _ := io.ReadAll(rsp.Body)
		_ = rsp.Body.Close()

		if rsp.StatusCode != http.StatusOK {
			err := fmt.Errorf("%d: %s", rsp.StatusCode, string(body))
			processErrors = errors.Join(processErrors, err)
			continue
		}

		for pemBlock, remains := pem.Decode(body); pemBlock != nil; pemBlock, remains = pem.Decode(remains) {
			if pemBlock.Type != "X509 CRL" {
				processErrors = errors.Join(processErrors, fmt.Errorf("invalid PEM block type: %s", pemBlock.Type))
				continue
			}
			crl, err := x509.ParseRevocationList(pemBlock.Bytes)
			if err != nil || crl == nil {
				processErrors = errors.Join(processErrors, err)
				continue
			}

			lists = append(lists, *crl)
		}
	}

	return lists, processErrors
}
