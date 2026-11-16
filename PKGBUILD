pkgname=run-puppet
pkgver=0.1.9
pkgrel=1
pkgdesc="Wrapper to run puppet in my homelab"
arch=('x86_64')
url="https://github.com/gentoomaniac/run-puppet"
license=('MIT')
options=('!debug')

# Point this to where your GoReleaser uploads the tarball
source=("${pkgname}::${url}/releases/download/v${pkgver}/${pkgname}_${pkgver}_linux-amd64")
sha256sums=('b2033be1592e3d9d1f9c22d3f1bc9bfe0525ee86e6d3c02de6f272be982c9365')

package() {
  # GoReleaser usually puts the binary in the root of the tarball
  install -Dm755 "${srcdir}/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
}
