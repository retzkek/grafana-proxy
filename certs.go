package main

import (
	"bytes"
	"crypto/x509"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// LoadCerts creates a cert pool with the PEM file at path,
// or every file with the. ".pem" extension if path is a directory,
// or the defaul system root CAs if path is empty.
func LoadCerts(path string) (*x509.CertPool, error) {
	if path == "" {
		log.Info("Using system root CA certs")
		return x509.SystemCertPool()
	}
	pool := x509.NewCertPool()

	fi, err := os.Lstat(path)
	if err != nil {
		return pool, err
	}

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		log.WithField("file", path).Info("Loading CA certs from file")
		err := addFile(pool, path)
		if err != nil {
			return pool, err
		}
	case mode.IsDir():
		log.WithField("dir", path).Info("Loading CA certs from directory")
		d, err := os.Open(path)
		if err != nil {
			return pool, err
		}
		defer d.Close()
		files, err := d.Readdirnames(0)
		if err != nil {
			return pool, err
		}
		for _, cert := range files {
			if strings.HasSuffix(cert, ".pem") {
				err := addFile(pool, singleJoiningSlash(path, cert))
				if err != nil {
					return pool, err
				}
			}
		}
	}
	return pool, nil
}

// addFile reads the PEM file at path and adds it to the cert pool
func addFile(pool *x509.CertPool, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		return err
	}
	if ok := pool.AppendCertsFromPEM(buf.Bytes()); !ok {
		return err
	}
	return nil
}
