pkgname=run-puppet
pkgver=0.1.10
pkgrel=1
pkgdesc="Wrapper to run puppet in my homelab"
arch=('x86_64')
url="https://github.com/gentoomaniac/run-puppet"
license=('MIT')
options=('!debug')

# Point this to where your GoReleaser uploads the tarball
source=("${pkgname}::${url}/releases/download/v${pkgver}/${pkgname}_${pkgver}_linux-amd64")
sha256sums=('639846ae396d250937877b920ac6dfab54586eed7079ed4328fd0983dcb2f934')

package() {
  # GoReleaser usually puts the binary in the root of the tarball
  install -Dm755 "${srcdir}/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
}
