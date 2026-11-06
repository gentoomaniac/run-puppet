pkgname=run-puppet
pkgver=0.1.7
pkgrel=1
pkgdesc="Wrapper to run puppet in my homelab"
arch=('x86_64')
url="https://github.com/gentoomaniac/run-puppet"
license=('MIT')
options=('!debug')

# Point this to where your GoReleaser uploads the tarball
source=("${pkgname}::${url}/releases/download/v${pkgver}/${pkgname}_${pkgver}_linux-amd64")
sha256sums=('d8c29f01762d46181734fe417dff624ad5fc9ed068c972b44b13972b7f53450a') 

package() {
  # GoReleaser usually puts the binary in the root of the tarball
  install -Dm755 "${srcdir}/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
}