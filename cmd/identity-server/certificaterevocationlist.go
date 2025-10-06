package main

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type CertificateRevocationList struct {

	// updateFunctionGuard is used to prevent multiple updates from happening at the same time.
	// This mutex MUST only be used in the Update function to avoid deadlocks between the listGuard
	// and the updateFunctionGuard.
	updateFunctionGuard *sync.Mutex

	// listGuard is used to protect the revoked certificates map. Concurrent reads are
	// allowed, but writes are exclusive. This is to allow multiple reads at the same time
	// while still allowing updates to the map.
	// A write lock on the listGuard MUST only be used in the Update function.
	listGuard *sync.RWMutex

	// timer is used to refresh the revoked certificates after a certain interval.
	// The timer is being managed by the Update function.
	timer *time.Timer

	// timerDone is used to signal the timer goroutine to stop
	timerDone chan struct{}

	// revoked is a map of revoked certificates. The key is the serial number of the
	// certificate in hex format. The value is an empty struct as we only test for existence.
	revoked map[string]struct{}

	// project, region, poolName and caName are used to identify the CA pool
	// and the CA certificate authority. These values are used to fetch the revoked
	// certificates from the CA pool.
	project  string
	region   string
	poolName string
	caName   string

	// clientRootCAs contains the CA certificates this CRL is based on.
	// This can be used for further verification of certificates in the CA pool.
	clientRootCAs []*x509.Certificate

	// maxUpdateInterval is the maximum duration after which the revoked certificates are refreshed.
	// The interval is reset after each call to Update.
	maxUpdateInterval time.Duration
}

// NewCertificateRevocationList creates a new CertificateRevocationList instance.
// In order to initialise the list, the Update function must be called once.
func NewCertificateRevocationList(clientRootCAs []*x509.Certificate, project, region, poolName, name string, updateInterval time.Duration) *CertificateRevocationList {
	crl := &CertificateRevocationList{
		timer:               nil,
		listGuard:           new(sync.RWMutex),
		updateFunctionGuard: new(sync.Mutex),
		revoked:             make(map[string]struct{}),
		project:             project,
		region:              region,
		poolName:            poolName,
		caName:              name,
		maxUpdateInterval:   updateInterval,
		clientRootCAs:       clientRootCAs,
		timerDone:           make(chan struct{}),
	}

	return crl
}

// IsCertFromPool checks if the given certificate is from the CA pool stored
// in the CertificateRevocationList.
func (crl *CertificateRevocationList) IsCertFromPool(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}

	for _, caCert := range crl.clientRootCAs {
		if cert.CheckSignatureFrom(caCert) == nil {
			return true
		}
	}
	return false
}

// IsRevoked checks if the given certificate is revoked.
// If the certificate is nil or has no serial number, it is considered revoked.
func (crl *CertificateRevocationList) IsRevoked(cert *x509.Certificate) bool {
	if cert == nil {
		return true
	}

	if cert.SerialNumber == nil {
		return true
	}

	hexSerial := hex.EncodeToString(cert.SerialNumber.Bytes())
	return crl.IsSerialRevoked(hexSerial)
}

// IsSerialRevoked checks if the given serial number is revoked.
// The serial number is expected to be in hex format.
func (crl *CertificateRevocationList) IsSerialRevoked(hexSerial string) bool {
	crl.listGuard.RLock()
	defer crl.listGuard.RUnlock()

	_, isRevoked := crl.revoked[hexSerial]
	return isRevoked
}

// StartUpdateTimer starts the timer to refresh the revoked certificates.
func (crl *CertificateRevocationList) StartUpdateTimer() {
	crl.updateFunctionGuard.Lock()
	defer crl.updateFunctionGuard.Unlock()

	// Don't start the timer if it is already running
	if crl.timer == nil {
		nextInvocation := time.Now().Add(crl.maxUpdateInterval)
		crl.timer = time.NewTimer(time.Until(nextInvocation))

		// Start the goroutine that listens to the timer
		go crl.timerLoop()
	}
}

// timerLoop runs in a goroutine and handles timer events
func (crl *CertificateRevocationList) timerLoop() {
	for {
		select {
		case <-crl.timerDone:
			return
		case <-crl.timer.C:
			// Call Update when the timer fires
			if err := crl.Update(context.Background()); err != nil {
				log.Error().Err(err).Msg("Failed to update revoked certificates from timer")
			}
		}
	}
}

// StopUpdateTimer stops the timer to refresh the revoked certificates.
func (crl *CertificateRevocationList) StopUpdateTimer() {
	crl.updateFunctionGuard.Lock()
	defer crl.updateFunctionGuard.Unlock()

	if crl.timer != nil {
		// signal timerLoop to stop
		close(crl.timerDone)
		crl.timer.Stop()
		crl.timer = nil
	}
}

// UpdateRevokedCertificates updates the revoked certificates in the
// RevokedCertificates map. It fetches the revoked certificates from the
// given CA pool and updates the map with the serial numbers of the revoked
// certificates. The map is cleared before updating it with the new revoked
// certificates. The function also sets a timer to refresh the revoked
// certificates after a certain interval. The timer is stopped if it is
// already running. The interval is determined by the CRL refresh interval
// configured in the server settings.
func (crl *CertificateRevocationList) Update(ctx context.Context) error {
	crl.updateFunctionGuard.Lock()
	defer crl.updateFunctionGuard.Unlock()

	// Stop the timer if it is running (we might get called directly).
	// This is to prevent the timer from firing while we are updating the revoked certificates
	// and to prevent multiple updates from happening at the same time.
	if crl.timer != nil {
		crl.timer.Stop()
	}

	// nextInvocation can be changed later in the code.
	// this is intended so that the deferred function uses the correct time for
	// the next update
	nextInvocation := time.Now().Add(crl.maxUpdateInterval)
	defer func() {
		log.Info().Time("nextUpdate", nextInvocation).Msg("Setting timer for next revoked certificate update")
		if crl.timer != nil {
			crl.timer.Reset(time.Until(nextInvocation))
		}
	}()

	log.Info().Msg("Updating revoked certificate list")

	// Get the revoked certificates from the CA pool
	crls, err := GetRevokedCertificates(crl.project, crl.region, crl.poolName, crl.caName, ctx)
	if err != nil {
		return err
	}

	// Use a temporary map so we can still use the old revoked certificates
	// while we are reading the new ones.
	revokedCertificates := make(map[string]struct{})

	for _, list := range crls {

		// We only accept CRLs that are signed by the client root CAs
		// This is to prevent CRLs from other CAs from being accepted
		validRoot := false
		for _, caCert := range crl.clientRootCAs {
			if list.CheckSignatureFrom(caCert) == nil {
				validRoot = true
				break
			}
		}

		if !validRoot {
			log.Error().Msg("CRL is not signed by any of the client root CAs")
			continue
		}

		// Make sure we adhere to the CRL next update time
		log.Info().Time("nextUpdate", list.NextUpdate).Msg("CRL next update")
		if list.NextUpdate.Before(nextInvocation) && list.NextUpdate.After(time.Now()) {
			nextInvocation = list.NextUpdate
		}

		// Convert the revoked certificates to hex format and add them to the map
		// for faster lookups.
		for _, revokedCert := range list.RevokedCertificateEntries {
			if revokedCert.SerialNumber == nil {
				log.Warn().Msg("Revoked certificate has no serial number")
				continue
			}
			hexSerial := hex.EncodeToString(revokedCert.SerialNumber.Bytes())
			revokedCertificates[hexSerial] = struct{}{}
		}
	}

	// Attention: The write lock MUST only be used in this function.
	// If a write lock is used somewhere else, you will likely create a deadlock because
	// of the timerLock being active here.

	crl.listGuard.Lock()
	defer crl.listGuard.Unlock()

	log.Info().Int("count", len(revokedCertificates)).Msg("Updated revoked certificate list")
	crl.revoked = revokedCertificates

	return nil
}
