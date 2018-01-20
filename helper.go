package hst

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// HTTPGet 获取http内容
func HTTPGet(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return bs, nil
}

// HTTPSGet 获取https内容
func HTTPSGet(url string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(url)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bs, err := ioutil.ReadAll(res.Body)
	return bs, nil
}

// TLSSGet 获取tls内容
func TLSSGet(url, ca, crt, key string) ([]byte, error) {
	caCrt, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCrt)

	cliCrt, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			Certificates:       []tls.Certificate{cliCrt},
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(url)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bs, err := ioutil.ReadAll(res.Body)
	return bs, nil
}

/*
MakeTLSFile 生成TLS双向认证证书
# 1.创建根证书密钥文件(自己做CA)root.key：
openssl genrsa -des3 -passout pass:123 -out ssl/root.key 2048
# 2.创建根证书的申请文件root.csr：
openssl req -passin pass:123 -new -subj "/C=CN/ST=Shanghai/L=Shanghai/O=MyCompany/OU=MyCompany/CN=localhost/emailAddress=hk@cdeyun.com" -key ssl/root.key -out ssl/root.csr
# 3.创建根证书root.crt：
openssl x509 -passin pass:123 -req -days 3650 -sha256 -extensions v3_ca -signkey ssl/root.key -in ssl/root.csr -out ssl/root.crt
rm -rf ssl/root.csr

# 1.创建客户端证书私钥
openssl genrsa -des3 -passout pass:456 -out ssl/ssl.key 2048
# 2.去除key口令
openssl rsa -passin pass:456 -in ssl/ssl.key -out ssl/ssl.key
# 3.创建客户端证书申请文件ssl.csr
openssl req -new -subj "/C=CN/ST=Shanghai/L=Shanghai/O=MyCompany/OU=MyCompany/CN=localhost/emailAddress=hk@cdeyun.com" -key ssl/ssl.key -out ssl/ssl.csr
# 4.创建客户端证书文件ssl.crt
openssl x509 -passin pass:123 -req -days 365 -sha256 -extensions v3_req -CA ssl/root.crt -CAkey ssl/root.key -CAcreateserial -in ssl/ssl.csr -out ssl/ssl.crt
rm -rf ssl/ssl.csr
rm -rf ssl/root.srl
# 5.将客户端证书文件ssl.crt和客户端证书密钥文件ssl.key合并成客户端证书安装包ssl.pfx
openssl pkcs12 -export -passout pass:789 -in ssl/ssl.crt -inkey ssl/ssl.key -out ssl/ssl.pfx
*/
func MakeTLSFile(passRoot, passKey, passPfx, path, domain, email string) bool {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	domain = path + domain
	os.MkdirAll(path, 0755)
	// 1.1.创建根证书密钥文件(自己做CA)root.key：
	exec.Command(`openssl`, `genrsa`, `-des3`, `-passout`, `pass:`+passRoot, `-out`, domain+`.ca.key`, `2048`).CombinedOutput()

	// 1.2.创建根证书的申请文件root.csr：
	exec.Command(`openssl`, `req`, `-passin`, `pass:`+passRoot, `-new`, `-subj`, `/C=CN/ST=Shanghai/L=Shanghai/O=MyCompany/OU=MyCompany/CN=`+domain+`/emailAddress=`+email, `-key`, domain+`.ca.key`, `-out`, domain+`.ca.csr`).CombinedOutput()

	// 1.3.创建根证书root.crt：
	exec.Command(`openssl`, `x509`, `-passin`, `pass:`+passRoot, `-req`, `-days`, `3650`, `-sha256`, `-extensions`, `v3_ca`, `-signkey`, domain+`.ca.key`, `-in`, domain+`.ca.csr`, `-out`, domain+`.ca.crt`).CombinedOutput()
	exec.Command(`rm`, domain+`.ca.csr`).CombinedOutput()

	// 2.1.创建客户端证书私钥
	exec.Command(`openssl`, `genrsa`, `-des3`, `-passout`, `pass:`+passKey, `-out`, domain+`.ssl.key`, `2048`).CombinedOutput()

	// 2.2.去除key口令
	exec.Command(`openssl`, `rsa`, `-passin`, `pass:`+passKey, `-in`, domain+`.ssl.key`, `-out`, domain+`.ssl.key`).CombinedOutput()

	// 2.3.创建客户端证书申请文件ssl.csr
	exec.Command(`openssl`, `req`, `-new`, `-subj`, `/C=US/ST=Mars/L=iTranswarp/O=iTranswarp/OU=iTranswarp/CN=`+domain+`/emailAddress=`+email, `-key`, domain+`.ssl.key`, `-out`, domain+`.ssl.csr`).CombinedOutput()

	// 2.4.创建客户端证书文件ssl.crt
	exec.Command(`openssl`, `x509`, `-passin`, `pass:`+passRoot, `-req`, `-days`, `365`, `-sha256`, `-extensions`, `v3_req`, `-CA`, domain+`.ca.crt`, `-CAkey`, domain+`.ca.key`, `-CAcreateserial`, `-in`, domain+`.ssl.csr`, `-out`, domain+`.ssl.crt`).CombinedOutput()
	exec.Command(`rm`, domain+`.ssl.csr`).CombinedOutput()

	// 2.5.将客户端证书文件ssl.crt和客户端证书密钥文件ssl.key合并成客户端证书安装包ssl.pfx
	exec.Command(`openssl`, `pkcs12`, `-export`, `-passout`, `pass:`+passPfx, `-in`, domain+`.ssl.crt`, `-inkey`, domain+`.ssl.key`, `-out`, domain+`.ssl.pfx`).CombinedOutput()
	exec.Command(`rm`, domain+`.srl`).CombinedOutput()

	// 3.校验
	bs1, _ := exec.Command(`openssl`, `x509`, `-noout`, `-modulus`, `-in`, domain+`.ssl.crt`).CombinedOutput()
	bs2, _ := exec.Command(`openssl`, `rsa`, `-noout`, `-modulus`, `-in`, domain+`.ssl.key`).CombinedOutput()
	return string(bs1) == string(bs2)
}
