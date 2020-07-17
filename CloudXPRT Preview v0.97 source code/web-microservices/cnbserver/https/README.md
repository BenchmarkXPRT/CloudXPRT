# To enable HTTPS/TLS in CNB

##### Generate private key (.key)

```sh
# Key considerations for algorithm "RSA" ≥ 2048-bit
openssl genrsa -out cnbserver.key 2048

# Key considerations for algorithm "ECDSA" ≥ secp384r1
# List ECDSA the supported curves (openssl ecparam -list_curves)
openssl ecparam -genkey -name secp384r1 -out cnbserver.key
```

##### Generation of self-signed(x509) public key (PEM-encodings `.pem`|`.crt`) based on the private (`.key`)

```sh
openssl req -new -x509 -sha256 -key cnbserver.key -out cnbserver.crt -days 3650
```

##### Put generated cnbserver.key and cnbserver.crt in the same directory of cnbserver executable, build and deploy the image

##### The docker image yxiay2k/webserver:v2.0 could server both HTTP(8070) and HTTPS(8443) requests. Sample gobench commands:

```sh
gobench -u http://IP:8070/ocr -c 2 -t 10
gobench -u https://IP:8443/ocr -c 2 -t 10
```

---
