## Example running TLS secured communication between plugins and framework

### Overview

The TLS demo scenario is here. 
* Demo is supposed to test TLS communication
in the scenario where all parties are running on a localhost.
* Steps for the demo were tested on Ubuntu 16.04.

Overview of the steps:
1. Generate own CA (Certificate Authority) certificate for demo.
2. Generate and sign demo keys for server- (plugins) and client- 
(server) gRPC side.
3. Start Snap.
4. Build snap-plugin-lib-go example plugins
5. Load plugins with certificates given in arguments.
6. Create a task.

#### Generating demo CA certificate

```sh
mkdir CA; cd CA
mkdir certs crl newcerts private
echo "01" > serial
cp /dev/null index.txt
cp /etc/ssl/openssl.cnf openssl.cnf
grep openssl.cnf -ie DemoCA

## alter OpenSSL configuration
sed 's#./DemoCA#.#gI' -i openssl.cnf
sed 's#cacert.pem#taocacert.pem#gI' -i openssl.cnf
sed 's#cakey.pem#taocakey.pem#gI' -i openssl.cnf
## manual step - add an entry to section [ usr_cert ] for local address
    subjectAltName=IP:127.0.0.1
## diff to original openssl conf should look like this:

diff openssl.cnf /etc/ssl/openssl.cnf                                                                                                       [1/5307]
42c42
< dir           = .             # Where everything is kept
---
> dir           = ./demoCA              # Where everything is kept
50c50
< certificate   = $dir/taocacert.pem    # The CA certificate
---
> certificate   = $dir/cacert.pem       # The CA certificate
55c55
< private_key   = $dir/private/taocakey.pem# The private key
---
> private_key   = $dir/private/cakey.pem# The private key
200d199
< subjectAltName=IP:127.0.0.1
331c330
< dir           = .             # TSA root directory
---
> dir           = ./demoCA              # TSA root directory
336c335
< certs         = $dir/taocacert.pem    # Certificate chain to include in reply
---
> certs         = $dir/cacert.pem       # Certificate chain to include in reply

## generate and install CA certificate
## enter any data, but rememeber the password encrypting the private key
openssl req -new -x509 -keyout private/taocakey.pem -out taocacert.pem -days 365 -config openssl.cnf
sudo cp taocacert.pem /usr/local/share/ca-certificates/taocacert.crt
## this will link demo certificate as trusted in system
sudo update-ca-certificates
```

#### Generate and sign demo keys for server- (plugins) and client- (server) gRPC side.

```sh
## request and sign server certificate
## upon request enter any data for all fields except CN - give `127.0.0.1`
openssl req -nodes -new -x509 -keyout taoserverkey.pem -out taoserverreq.pem -days 365 -config openssl.cnf
openssl x509 -x509toreq -in taoserverreq.pem -signkey taoserverkey.pem -out tmp.pem
openssl ca -config openssl.cnf -policy policy_anything -out taoservercert.pem -infiles tmp.pem
rm tmp.pem
## request and sign client certificate
## upon request enter any data for all fields except CN - give `127.0.0.1`
openssl req -nodes -new -x509 -keyout taoclientkey.pem -out taoclientreq.pem -days 365 -config openssl.cnf
openssl x509 -x509toreq -in taoclientreq.pem -signkey taoclientkey.pem -out tmp.pem
openssl ca -config openssl.cnf -policy policy_anything -out taoclientcert.pem -infiles tmp.pem
rm tmp.pem
sudo cp taoserver*.pem taoclient*.pem /tmp
```

#### Start Snap as usual

`snapteld --plugin-trust 0 --log-level 1` 

#### Build snap-plugin-lib-go example plugins

`for k in collector processor publisher; do go build -o example-$k examples/$k/main.go; done`

#### Load plugins with certificates given in arguments

```sh
snaptel plugin load example-processor --plugin-cert=/tmp/taoclientcert.pem --plugin-key=/tmp/taoclientkey.pem
snaptel plugin load example-collector --plugin-cert=/tmp/taoclientcert.pem --plugin-key=/tmp/taoclientkey.pem
snaptel plugin load example-publisher --plugin-cert=/tmp/taoclientcert.pem --plugin-key=/tmp/taoclientkey.pem
```

#### Create and watch a task

```sh
snaptel task create -t examples/task.yml
#
#Using task manifest to create task
#Task created
#ID: 9eb91d45-3427-4847-b119-47d3cd43b793
#Name: Task-9eb91d45-3427-4847-b119-47d3cd43b793
#State: Running
#
snaptel task watch 9eb91d45-3427-4847-b119-47d3cd43b793
#
#Watching Task (9eb91d45-3427-4847-b119-47d3cd43b793):

snaptel task watch 
```

- where `examples/task.yml` is the example given in [readme for examples/](../README.md).  

### Known issues

Client doesn't yet authenticate to servers - ie.: framework doesn't
provide its certificate. Additionally, the plugins do not yet demand the
certificate from framework.

