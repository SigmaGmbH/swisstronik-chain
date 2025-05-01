use std::prelude::v1::*;

pub struct ClientAuth {
    outdated_ok: bool,
}

impl ClientAuth {
    pub fn new(outdated_ok: bool) -> ClientAuth {
        ClientAuth { outdated_ok }
    }
}

#[cfg(all(feature = "hardware_mode", not(feature = "mainnet")))]
impl rustls::ClientCertVerifier for ClientAuth {
    fn client_auth_root_subjects(
        &self,
        _sni: Option<&webpki::DNSName>,
    ) -> Option<rustls::DistinguishedNames> {
        Some(rustls::DistinguishedNames::new())
    }

    fn verify_client_cert(
        &self,
        certs: &[rustls::Certificate],
        _sni: Option<&webpki::DNSName>,
    ) -> Result<rustls::ClientCertVerified, rustls::TLSError> {
        if certs.is_empty() {
            println!("[Enclave] No certs provided for Client Auth");
            return Err(rustls::TLSError::NoCertificatesPresented);
        }

        crate::attestation::cert::verify_dcap_cert(&certs[0].0).map_err(|err| {
            println!(
                "[Attestastion Server] Cannot verify DCAP cert. Reason: {:?}",
                err
            );
            rustls::TLSError::WebPKIError(webpki::Error::ExtensionValueInvalid)
        })?;
        Ok(rustls::ClientCertVerified::assertion())
    }
}

#[cfg(all(feature = "hardware_mode", feature = "mainnet"))]
impl rustls::ClientCertVerifier for ClientAuth {
    fn client_auth_root_subjects(
        &self,
        _sni: Option<&webpki::DNSName>,
    ) -> Option<rustls::DistinguishedNames> {
        Some(rustls::DistinguishedNames::new())
    }

    fn verify_client_cert(
        &self,
        certs: &[rustls::Certificate],
        _sni: Option<&webpki::DNSName>,
    ) -> Result<rustls::ClientCertVerified, rustls::TLSError> {
        if certs.is_empty() {
            println!("[Enclave] No certs provided for Client Auth");
            return Err(rustls::TLSError::NoCertificatesPresented);
        }

        crate::attestation::cert::verify_dcap_cert(&certs[0].0).unwrap();
        Ok(rustls::ClientCertVerified::assertion())
    }
}

pub struct ServerAuth {
    outdated_ok: bool,
}

impl ServerAuth {
    pub fn new(outdated_ok: bool) -> ServerAuth {
        ServerAuth { outdated_ok }
    }
}

#[cfg(all(feature = "hardware_mode", feature = "mainnet"))]
impl rustls::ServerCertVerifier for ServerAuth {
    fn verify_server_cert(
        &self,
        _roots: &rustls::RootCertStore,
        certs: &[rustls::Certificate],
        _hostname: webpki::DNSNameRef,
        _ocsp: &[u8],
    ) -> Result<rustls::ServerCertVerified, rustls::TLSError> {
        if certs.is_empty() {
            println!("[Enclave] No certs provided for Server Auth");
            return Err(rustls::TLSError::NoCertificatesPresented);
        }

        crate::attestation::cert::verify_dcap_cert(&certs[0].0).unwrap();
        Ok(rustls::ServerCertVerified::assertion())
    }
}

#[cfg(all(feature = "hardware_mode", not(feature = "mainnet")))]
impl rustls::ServerCertVerifier for ServerAuth {
    fn verify_server_cert(
        &self,
        _roots: &rustls::RootCertStore,
        certs: &[rustls::Certificate],
        _hostname: webpki::DNSNameRef,
        _ocsp: &[u8],
    ) -> Result<rustls::ServerCertVerified, rustls::TLSError> {
        if certs.is_empty() {
            println!("[Enclave] No certs provided for Server Auth");
            return Err(rustls::TLSError::NoCertificatesPresented);
        }

        crate::attestation::cert::verify_dcap_cert(&certs[0].0).unwrap();
        Ok(rustls::ServerCertVerified::assertion())
    }
}
