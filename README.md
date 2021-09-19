# ssl-pairgen

The `ssl-pairgen` utility is for generating user private client pkcs#12 bundles and certificate authorities, for use in a [Kubernetes](https://kubernetes.io) cluster.

## Rationale and Application

With the growth of Kubernetes, an interesting type of application has also grown in popularity - the Single User Pod. First observed in the [Gitpod](https://gitpod.io) online coding environment offering, and later in the [Eclipse Che](https://www.eclipse.org/che) "Kubernetes-Native IDE". To simplify the creation and use of this variety of pod, the concept of the User Private Certificate Authority was created.

Conceptually this operates in similar fashion to ssh key pairs, with the caveat that a User Private Certificate Authority can only grant a single user access to a single pod,(*) unlike `ssh` which can grant several users access to a host.

## How it Works

Behind the curtains, this tool generates a temporary CA, then creates and signs a user certificate for client authentication. The CA signing key is immediately discarded, leaving two useful artifacts:

### User CA
The `<user>-ca.crt` public CA certificate, which gets loaded to the cluster with a user-specific name and/or namespace (this package is not prescriptive). This CA cert is later referenced during creation of the user pod ingress definition, taking advantage of the Kubernetes [nginx-ingress](https://kubernetes.github.io/ingress-nginx/) standard [client certificate authentication](https://kubernetes.github.io/ingress-nginx/examples/auth/client-certs/) feature.

### User Client Certificate
The `<user>.p12` encrypted PKCS12 bundle actually includes two cryptography objects:
* The user's public certificate, signed by the temporary CA key and presented to the ingress controller to be validated by the user's CA cert
* The user's private key, used to prove the user has the private key corresponding to the signed public key

This bundle can be imported with it's passphrase by any modern browser, and the protocol allows the browser to easily identify the required certificate for authentication.

### User PEM
The user PEM file is informational only - the public certificate part of the **p12** file. It can be examined with `openssl`:
```
$ openssl x509 -in david.pem -text
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            <...>
        Signature Algorithm: sha256WithRSAEncryption
        Issuer: O = Linux Jedi, CN = david-privateCA
        Validity
            Not Before: Sep 19 20:46:26 2021 GMT
            Not After : Sep 19 20:46:26 2022 GMT
        Subject: O = Linux Jedi, CN = david
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                RSA Public-Key: (2048 bit)
                Modulus:
                    <...>
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Extended Key Usage: 
                TLS Web Client Authentication
            X509v3 Authority Key Identifier: 
                keyid:<...>

            X509v3 Subject Alternative Name: 
                IP Address:127.0.0.1
    Signature Algorithm: sha256WithRSAEncryption
        ...
-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
```

## Other Thoughts and Applications

### Short-Lived Use
Since the typical user workflow would be to create a certificate and immediately load it in to the desired browser(s), constraining the CA validity to a matter of months is less cumbersome to the user than similarly frequent required password changes.

### Trivial Revocation
Revoking the user's certificate is trivially and permanently done by deleting the user CA from the cluster.

### Private Developer Clusters
To simplify authentication and authorization to single-user clusters, this mechanism can be easily used to protect all cluster web applications.(**)

## Usage

### User Usage
Users can download and un-tar the appropriate release for their platform, and run `./ssl-pairgen`:
```
$ ssl-pairgen "Linux Jedi" david
Generating CA and user keys and certificates ...
Enter passphrase for encrypting david.p12: 
Re-enter passphrase: 
Wrote user private CA cert david-ca.crt, browser package david.p12, and informational david.pem
```

In this example, `parsley42-ca.crt` is provided to the cluster administrator (or uploaded with custom tooling), serving as an analog to an **ssh** public key. The **pkcs#12** file `parsley42.p12` is kept private, and imported into the browser(s) desired by the user.

### Cluster Administrator Use
The trivial `kube-ca.sh` script is provided, mostly as a reference, for generating the user CA secret. To illustrate a full example, continuing from above, the cluster administrator could create the private CA secret in a `parsley42` namespace:
```
$ ./kube-ca.sh parsley42-ca.crt parsley42-ca parsley42
apiVersion: v1
data:
  ca.crt: <...>
kind: Secret
metadata:
  creationTimestamp: null
  name: parsley42-ca
  namespace: parsley42
```
This could be loaded to the cluster with:
```
$ ./kube-ca.sh parsley42-ca.crt parsley42-ca parsley42 | kubectl apply -f -`
```
Then, assuming the use of the `ingress-nginx` ingress controller, the administrator could create a service only accessible by the given user by applying these `ingress-nginx`-specific annotations:
```yaml
    nginx.ingress.kubernetes.io/auth-tls-secret: "parsley42/parsley42-ca"
    nginx.ingress.kubernetes.io/auth-tls-verify-client: "on"
```

> (*) - Provided the user does not share their private client certificate and passphrase.

> (**) - This was (and is) the initial application for the author.
