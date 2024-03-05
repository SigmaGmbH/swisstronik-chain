use sgx_rand::*;
use sgx_tcrypto::*;
use sgx_tse::*;
use sgx_types::*;

use std::io::{Read, Write};
use std::net::TcpStream;
use std::prelude::v1::*;
use std::ptr;
use std::str;
use std::string::String;
use std::sync::Arc;
use std::vec::Vec;

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
        // This call will automatically verify cert is properly signed
        match crate::attestation::cert::verify_ra_cert(&certs[0].0, None) {
            Ok(_) => Ok(rustls::ClientCertVerified::assertion()),
            Err(crate::attestation::types::AuthResult::SwHardeningAndConfigurationNeeded)
            | Err(crate::attestation::types::AuthResult::GroupOutOfDate) => {
                if self.outdated_ok {
                    println!("outdated_ok is set, overriding outdated error");
                    Ok(rustls::ClientCertVerified::assertion())
                } else {
                    Err(rustls::TLSError::WebPKIError(
                        webpki::Error::ExtensionValueInvalid,
                    ))
                }
            }
            Err(_) => Err(rustls::TLSError::WebPKIError(
                webpki::Error::ExtensionValueInvalid,
            )),
        }
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
        // This call will automatically verify cert is properly signed
        match super::cert::verify_ra_cert(&certs[0].0, None) {
            Ok(_) => Ok(rustls::ClientCertVerified::assertion()),
            Err(super::types::AuthResult::SwHardeningAndConfigurationNeeded) => {
                if self.outdated_ok {
                    println!("outdated_ok is set, overriding outdated error");
                    Ok(rustls::ClientCertVerified::assertion())
                } else {
                    Err(rustls::TLSError::WebPKIError(
                        webpki::Error::ExtensionValueInvalid,
                    ))
                }
            }
            Err(_) => Err(rustls::TLSError::WebPKIError(
                webpki::Error::ExtensionValueInvalid,
            )),
        }
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
        // This call will automatically verify cert is properly signed
        let res = super::cert::verify_ra_cert(&certs[0].0, None);
        match res {
            Ok(_) => Ok(rustls::ServerCertVerified::assertion()),
            Err(super::types::AuthResult::SwHardeningAndConfigurationNeeded) => {
                if self.outdated_ok {
                    println!("outdated_ok is set, overriding outdated error");
                    Ok(rustls::ServerCertVerified::assertion())
                } else {
                    Err(rustls::TLSError::WebPKIError(
                        webpki::Error::ExtensionValueInvalid,
                    ))
                }
            }
            Err(_) => Err(rustls::TLSError::WebPKIError(
                webpki::Error::ExtensionValueInvalid,
            )),
        }
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
        // This call will automatically verify cert is properly signed
        let res = crate::attestation::cert::verify_ra_cert(&certs[0].0, None);
        match res {
            Ok(_) => Ok(rustls::ServerCertVerified::assertion()),
            Err(crate::attestation::types::AuthResult::SwHardeningAndConfigurationNeeded)
            | Err(crate::attestation::types::AuthResult::GroupOutOfDate) => {
                if self.outdated_ok {
                    println!("outdated_ok is set, overriding outdated error");
                    Ok(rustls::ServerCertVerified::assertion())
                } else {
                    Err(rustls::TLSError::WebPKIError(
                        webpki::Error::ExtensionValueInvalid,
                    ))
                }
            }
            Err(_) => Err(rustls::TLSError::WebPKIError(
                webpki::Error::ExtensionValueInvalid,
            )),
        }
    }
}
