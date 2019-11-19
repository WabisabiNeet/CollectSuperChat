// Copyright (C) 2019 RICOH Co., Ltd. All rights reserved.

package ytproxy

import "crypto/tls"

// from: https://github.com/elazarl/goproxy/blob/master/https.go

func TLSConfigFromCA(ca *tls.Certificate, host string) (*tls.Config, error) {
	var err error
	var cert *tls.Certificate

	config := tls.Config{}

	cert, err = signHost(*ca, []string{host})

	if err != nil {
		return nil, err
	}

	config.Certificates = append(config.Certificates, *cert)
	return &config, nil
}
